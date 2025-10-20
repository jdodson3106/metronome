package metronome

import (
	"fmt"
	"time"
)

const (
	MICROS_PER_MIN = 60_000_000
)

type Beat struct {
	Number int
	tone   *Sample
}

func (b *Beat) PlayTone() {
	b.tone.Play()
}

type TimeSignature struct {
	Beats int
	Notes Note
}

func (t *TimeSignature) BeatsFromTS() []Beat {
	var err error
	b := make([]Beat, t.Beats)
	for i := range t.Beats {
		var s *Sample
		if i == 0 {
			s, err = NewSample(PITCH_THREE)
		} else {
			s, err = NewSample(PITCH_ONE)
		}
		if err != nil {
			fmt.Printf("error loading sample file :: %s\n", err)
		}
		b[i] = Beat{Number: i + 1, tone: s}
	}
	return b
}

func (t *TimeSignature) Pretty() string {
	n, _ := t.Notes.Int()
	return fmt.Sprintf("%d/%d", t.Beats, n)
}

type MetronomeSettings struct {
	SoundOn   bool
	LoopCount int
}

type Metronome struct {
	BPM           int
	Beats         []Beat
	TimeSignature TimeSignature
	Settings      MetronomeSettings

	Ticker chan int64

	done chan bool
	mpb  int64
}

func Initialize(bpm int, ts TimeSignature) *Metronome {
	mpb := int64(MICROS_PER_MIN / bpm)

	return &Metronome{
		TimeSignature: ts,
		BPM:           bpm,
		Beats:         ts.BeatsFromTS(),
		Ticker:        make(chan int64),
		Settings:      MetronomeSettings{true, -1},
		done:          make(chan bool),
		mpb:           mpb,
	}
}

func NewMetronome(BPM int, ts TimeSignature) *Metronome {
	mpb := int64(MICROS_PER_MIN / BPM)
	return &Metronome{
		BPM:    BPM,
		done:   make(chan bool),
		mpb:    mpb,
		Ticker: make(chan int64),
	}
}

func (m *Metronome) Start() {
	next := time.Now().Add(time.Duration(m.mpb) * time.Microsecond)
	driftWatcher := m.mpb

	for {
		select {
		case <-m.done:
			// TODO: Handle freeing any file resources here
			return
		default:
			dt := time.Since(next)
			if dt >= time.Duration(driftWatcher) {
				driftWatcher -= int64(dt)
				driftWatcher += m.mpb
				next = time.Now().Add(time.Duration(driftWatcher) * time.Microsecond)
				m.Ticker <- int64(driftWatcher)
			}
		}
	}
}

func (m *Metronome) Pause() {

}

func (m *Metronome) Stop() {
	m.done <- true
}

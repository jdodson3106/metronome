package metronome

import (
	"fmt"
	"time"
)

type Metronome struct {
	BPM int

	mpb    int64
	ticker chan int64
	done   chan bool
}

func NewMetronome(BPM int) *Metronome {
	mpb := int64((1_000_000 * 60) / BPM)
	//mpb := int64((1_000 * 60) / BPM)
	return &Metronome{
		BPM:    BPM,
		mpb:    mpb,
		ticker: make(chan int64),
		done:   make(chan bool),
	}
}

func (m *Metronome) start() {
	startTimer(m.mpb, m.ticker, m.done)
}

func startTimer(interval int64, ticker chan int64, done chan bool) {
	next := time.Now().Add(time.Duration(interval) * time.Microsecond)
	fmt.Println(interval)
	driftWatcher := interval

	for {
		select {
		case <-done:
			fmt.Printf("\nstopping timer\n")
			return
		default:
			dt := time.Since(next)
			if dt >= time.Duration(driftWatcher) {
				//fmt.Printf("\r\033[2K%d", dt)
				driftWatcher -= int64(dt)
				driftWatcher += interval
				next = time.Now().Add(time.Duration(driftWatcher) * time.Microsecond)
				ticker <- int64(driftWatcher)
			}
		}
	}
}

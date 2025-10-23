package metronome

import (
	"bytes"
	_ "embed"
	"io"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
	mp "github.com/jdodson3106/metronome/internal/mp3"
	//"github.com/jdodson3106/metronome/lib/utils"
	"github.com/pkg/errors"
)

//go:embed audio_samples/c6-tone.mp3
var tone3 []byte

//go:embed audio_samples/c5-tone.mp3
var tone2 []byte

//go:embed audio_samples/c4-tone.mp3
var tone1 []byte

const (
	sampleRate       = 44100
	mp3NumChannels   = 2
	mp3Precision     = 2
	mp3BytesPerFrame = mp3NumChannels * mp3Precision

	PITCH_ONE = iota
	PITCH_TWO
	PITCH_THREE
)

type MetronomeSound int

var ctx *Context

type Context struct {
	c  *oto.Context
	mu sync.Mutex
}

func (c *Context) getPlayer(d *mp3.Decoder) (*oto.Player, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	d.Seek(0, io.SeekStart)
	// TODO: This trimming needs to be pulled to an external method that scans the mp3 file and cuts off the
	// dead space, and then trims the tone based on time not percentage
	//size := utils.TrimmedTo(10, d.Length())
	player, err := mp.NewMP3ReadSeeker(d, 0, d.Length())
	if err != nil {
		return nil, errors.Wrap(err, "error getting new MP3ReadSeeker")
	}
	return c.c.NewPlayer(player), nil
}

// Sample represents the audio
// file used for the
// metronome tick sounds
type Sample struct {
	TimeMicros int64
	rc         *io.ReadCloser
	d          *mp3.Decoder
	player     *oto.Player
}

func NewSample(tone MetronomeSound, mpb int64) (*Sample, error) {
	s := &Sample{TimeMicros: mpb}

	//f, err := os.Open("")
	//if err != nil {
	//	return s, errors.Wrap(err, "os.Open")
	//}

	// create a new decoder
	d, err := getDecoderForTone(tone)
	if err != nil {
		return s, errors.Wrap(err, "mp3")
	}
	s.d = d

	// init/get the context and grab a new player
	c, err := getContext()
	s.player, err = c.getPlayer(s.d)
	if err != nil {
		return s, errors.Wrap(err, "getPlayer()")
	}

	return s, nil
}

func (s *Sample) Play() {
	end := time.Now().Add(time.Duration(s.TimeMicros) * time.Microsecond)
	s.player.Play()
	for end.After(time.Now()) {
	}
	s.player.Pause()
	s.player.Seek(0, io.SeekStart)
}

func (s *Sample) Close() {
}

func getDecoderForTone(t MetronomeSound) (*mp3.Decoder, error) {
	var r io.Reader
	switch t {
	case PITCH_ONE:
		r = bytes.NewReader(tone1)
	case PITCH_TWO:
		r = bytes.NewReader(tone2)
	case PITCH_THREE:
		r = bytes.NewReader(tone3)
	default:
		return nil, errors.New("invalid MetronomeSound type")
	}

	return mp3.NewDecoder(r)
}

// getContext retrieves the wrapped oto.Context
// object. There can only be 1 context so I wrap
// and manage the concurrency here.
// TODO: Dig through the oto lib to see if the context is already thread safe during access
func getContext() (*Context, error) {
	if ctx == nil {
		op := &oto.NewContextOptions{}
		op.SampleRate = sampleRate
		op.ChannelCount = mp3NumChannels
		op.Format = oto.FormatSignedInt16LE

		c, ready, err := oto.NewContext(op)
		if err != nil {
			return nil, errors.Wrap(err, "oto context")
		}

		// block until we get the message the context is loaded
		<-ready

		ctx = &Context{c: c}
	}

	return ctx, nil
}

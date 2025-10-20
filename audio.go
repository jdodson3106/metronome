package metronome

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
	mp "github.com/jdodson3106/metronome/internal/mp3"
	"github.com/jdodson3106/metronome/lib/utils"
	"github.com/pkg/errors"
)

const (
	sampleRate       = 44100
	mp3NumChannels   = 2
	mp3Precision     = 2
	mp3BytesPerFrame = mp3NumChannels * mp3Precision

	PITCH_ONE   MetronomeSound = "audio_samples/c4-tone.mp3"
	PITCH_TWO   MetronomeSound = "audio_samples/c5-tone.mp3"
	PITCH_THREE MetronomeSound = "audio_samples/c6-tone.mp3"
)

type MetronomeSound string

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
	size := utils.TrimmedTo(10, d.Length())
	fmt.Printf("full :: %d | 10%% %d\n", d.Length(), size)
	halfPlayer, err := mp.NewMP3ReadSeeker(d, 0, size)
	if err != nil {
		return nil, errors.Wrap(err, "error getting new MP3ReadSeeker")
	}
	return c.c.NewPlayer(halfPlayer), nil
}

// Sample represents the audio
// file used for the
// metronome tick sounds
type Sample struct {
	rc     *io.ReadCloser
	d      *mp3.Decoder
	player *oto.Player
}

func NewSample(file MetronomeSound) (Sample, error) {
	s := Sample{}

	f, err := os.Open(string(file))
	if err != nil {
		return s, errors.Wrap(err, "os.Open")
	}

	// create a new decoder
	d, err := mp3.NewDecoder(f)
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

func (s Sample) Play() {
	s.player.Play()
	for s.player.IsPlaying() {
	}
	s.player.Seek(0, io.SeekStart)
}

func (s *Sample) Close() {
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

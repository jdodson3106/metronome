package metronome

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
	"github.com/pkg/errors"
)

const (
	sampleRate       = 44100
	mp3NumChannels   = 2
	mp3Precision     = 2
	mp3BytesPerFrame = mp3NumChannels * mp3Precision
	frameSize        = 4608 // the frame size used to buffer in the oto/mp3 libs

	PITCH_ONE   MetronomeSound = "audio_samples/c4-tone.mp3"
	PITCH_TWO   MetronomeSound = "audio_samples/c5-tone.mp3"
	PITCH_THREE MetronomeSound = "audio_samples/c6-tone.mp3"
)

type MetronomeSound string

var ctx *Context

type MP3ReadSeeker struct {
	rs    io.ReadSeeker
	start int64 // absolute start
	limit int64 // number of bites allowed to read from start
	pos   int64 // current position from start (start..limit)
}

func NewMP3ReadSeeker(r io.ReadSeeker, start, limit int64) (*MP3ReadSeeker, error) {
	// first setup the seeker to start at the provided start pos
	if _, err := r.Seek(start, io.SeekStart); err != nil {
		return nil, err
	}

	return &MP3ReadSeeker{rs: r, start: start, limit: limit, pos: 0}, nil
}

func (rs *MP3ReadSeeker) Read(p []byte) (int, error) {
	// first make sure we are not at the end of the buffer already

	if rs.pos >= rs.limit {
		fmt.Println("ALREADY READ THE ENTIRE FILE!")
		return 0, io.EOF
	}

	// see how many bytes we have remaining to read
	// and compare that to the buffer being asked to read
	remaining := rs.limit - rs.pos
	toRead := int64(len(p))
	fmt.Printf("remaining :: %d | toRead :: %d\n", remaining, toRead)
	if remaining < toRead {
		toRead = remaining
	}

	// read the data from start to the max we can support
	n, err := rs.rs.Read(p[:toRead])
	fmt.Printf("read %d bytes\n", n)

	rs.pos += int64(n)
	if err != nil {
		return n, err
	}

	// if we read to the end or over then
	// surface the io.EOF 10% of the total file bytes
	if rs.pos >= rs.limit {
		len64 := int64(len(p))
		trimmed := trimmedTo(10, len64)
		return int(trimmed), io.EOF
	}

	return n, nil
}

func (rs *MP3ReadSeeker) Seek(offset int64, whence int) (int64, error) {
	fmt.Println("seek()...")
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = rs.start + offset
	case io.SeekCurrent:
		abs = rs.start + rs.pos + offset
	case io.SeekEnd:
		abs = rs.start + rs.limit + offset
	default:
		return 0, fmt.Errorf("invalid whence value provided")
	}

	if abs < rs.start {
		return 0, errors.New("seeking before start")
	}

	if _, err := rs.rs.Seek(abs, whence); err != nil {
		return 0, errors.Wrap(err, "mp3 seek failed")
	}

	rs.pos = abs - rs.start
	fmt.Printf("abs: %d, pos: %d\n", abs, rs.pos)
	return rs.pos, nil
}

type Context struct {
	c  *oto.Context
	mu sync.Mutex
}

func (c *Context) getPlayer(d *mp3.Decoder) (*oto.Player, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	d.Seek(0, io.SeekStart)
	// TOOD: figure out how to make this size a paramter
	size := trimmedTo(10, d.Length())
	fmt.Printf("full :: %d | 10%% %d\n", d.Length(), size)
	halfPlayer, err := NewMP3ReadSeeker(d, 0, size)
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

func NewSample(file MetronomeSound) (*Sample, error) {
	s := &Sample{}

	f, err := os.Open(string(file))
	if err != nil {
		return nil, errors.Wrap(err, "os.Open")
	}

	// create a new decoder
	d, err := mp3.NewDecoder(f)
	if err != nil {
		return nil, errors.Wrap(err, "mp3")
	}
	s.d = d

	// init/get the context and grab a new player
	c, err := getContext()
	s.player, err = c.getPlayer(s.d)
	if err != nil {
		return nil, errors.Wrap(err, "getPlayer()")
	}

	return s, nil
}

func (s *Sample) Play() {
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

func trimmedTo(perc, total int64) int64 {
	return perc * total / 100
}

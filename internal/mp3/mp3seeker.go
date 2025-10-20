package mp3

import (
	"fmt"
	"io"

	"github.com/jdodson3106/metronome/lib/utils"
	"github.com/pkg/errors"
)

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
		return 0, io.EOF
	}

	// see how many bytes we have remaining to read
	// and compare that to the buffer being asked to read
	remaining := rs.limit - rs.pos
	toRead := int64(len(p))
	if remaining < toRead {
		toRead = remaining
	}

	// read the data from start to the max we can support
	n, err := rs.rs.Read(p[:toRead])
	rs.pos += int64(n)
	if err != nil {
		return n, err
	}

	// if we read to the end or over then
	// surface the io.EOF 10% of the total file bytes
	if rs.pos >= rs.limit {
		len64 := int64(len(p))
		trimmed := utils.TrimmedTo(10, len64)
		return int(trimmed), io.EOF
	}

	return n, nil
}

func (rs *MP3ReadSeeker) Seek(offset int64, whence int) (int64, error) {
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
	return rs.pos, nil
}

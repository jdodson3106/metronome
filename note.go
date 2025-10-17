package metronome

import "fmt"

const (
	WHOLE        Note = 1
	HALF         Note = 2
	QUARTER      Note = 4
	EIGHTH       Note = 8
	SIXTEENTH    Note = 16
	THIRTYSECOND Note = 32
)

var notes = [6]Note{WHOLE, HALF, QUARTER, EIGHTH, SIXTEENTH, THIRTYSECOND}
var noteInts = [6]int{1, 2, 4, 8, 16, 32}

type Note int

func (n Note) Int() (int, error) {
	switch n {
	case WHOLE:
		return 1, nil
	case HALF:
		return 2, nil
	case QUARTER:
		return 4, nil
	case EIGHTH:
		return 8, nil
	case SIXTEENTH:
		return 16, nil
	case THIRTYSECOND:
		return 32, nil
	default:
		return -1, fmt.Errorf("invalid note value")
	}
}

func ToNote(n int) (Note, error) {
	switch n {
	case 1:
		return WHOLE, nil
	case 2:
		return HALF, nil
	case 4:
		return QUARTER, nil
	case 8:
		return EIGHTH, nil
	case 16:
		return SIXTEENTH, nil
	case 32:
		return THIRTYSECOND, nil
	default:
		return -1, fmt.Errorf("invalid note value %d", n)
	}
}

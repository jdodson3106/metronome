package metronome

// Sample represents the audio
// file used for the
// metronome tick sounds
type Sample struct {
	Speaker *Speaker
}

// Speaker is used to actually
// playback the tone
type Speaker struct {
}

func (s *Speaker) Play() error {
	return nil
}

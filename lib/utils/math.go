package utils

func TrimmedTo(perc, total int64) int64 {
	return perc * total / 100
}

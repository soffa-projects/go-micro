package dates

import (
	"time"
)

func Now() time.Time {
	return time.Now().UTC()
}

func NowPrt() *time.Time {
	value := time.Now().UTC()
	return &value
}

func NowPtrPlus(d time.Duration) *time.Time {
	value := time.Now().UTC()
	value = value.Add(d)
	return &value
}

func NowPlus(d time.Duration) time.Time {
	value := time.Now().UTC()
	value = value.Add(d)
	return value
}

func Days(value int) time.Duration {
	return time.Hour * 24 * time.Duration(value)
}

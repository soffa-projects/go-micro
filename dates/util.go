package dates

import "time"

func Now() time.Time {
	return time.Now().UTC()
}

package utils

import "time"

func ToLocal(t time.Time, tz string) time.Time {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return t.UTC()
	}
	return t.In(loc)
}

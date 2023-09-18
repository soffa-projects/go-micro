package h

import (
	"strings"
	"time"
)

func ParseDate(value *string) (*time.Time, error) {
	if value == nil || *value == "" {
		return nil, nil
	}
	tx, err := time.Parse("2006-01-02", strings.Split(*value, "T")[0])
	return &tx, err
}

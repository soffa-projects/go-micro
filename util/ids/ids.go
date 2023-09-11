package ids

import (
	"github.com/rs/xid"
	"strings"
)

func NewId(prefix string) string {
	guid := xid.New()
	if !strings.HasSuffix(prefix, "_") && !strings.HasSuffix(prefix, "-") {
		prefix += "_"
	}
	return prefix + guid.String()
}

func NewIdPtr(prefix string) *string {
	value := NewId(prefix)
	return &value
}

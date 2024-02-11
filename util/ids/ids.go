package ids

import (
	"github.com/rs/xid"
	"github.com/soffa-projects/go-micro/util/h"
	"strings"
)

func NewId(prefix string) string {
	guid := xid.New()
	if h.IsNotEmpty(prefix) && !strings.HasSuffix(prefix, "_") && !strings.HasSuffix(prefix, "-") {
		prefix += "_"
	}
	return prefix + guid.String()
}

func NewIdPtr(prefix string) *string {
	value := NewId(prefix)
	return &value
}

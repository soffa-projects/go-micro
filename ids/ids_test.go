package ids

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestId(t *testing.T) {
	value := NewId("test_")
	assert.NotEmpty(t, value)
	assert.True(t, strings.HasPrefix(value, "test_"))
	assert.True(t, len(value) > 10)

}

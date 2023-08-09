package h

import (
	testifyAssert "github.com/stretchr/testify/assert"
	"testing"
)

func TestRoundFloat(t *testing.T) {
	assert := testifyAssert.New(t)

	assert.Equal(560.99, RoundFloat(560.993, 2))
	assert.Equal(193.00, RoundFloat(192.995, 2))
	assert.Equal(8928391938.12346, RoundFloat(8928391938.123456, 5))
	assert.Equal(8928391938.12345, RoundFloat(8928391938.123454, 5))
}

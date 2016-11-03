package vitotrol

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResultHeader(t *testing.T) {
	assert := assert.New(t)

	rh := &ResultHeader{
		ErrorNum: 0,
		ErrorStr: "No error",
	}

	var _ error = rh

	assert.Equal("No error [#0]", rh.Error())
	assert.False(rh.IsError())

	rh = &ResultHeader{
		ErrorNum: 42,
		ErrorStr: "Big error",
	}
	assert.Equal("Big error [#42]", rh.Error())
	assert.True(rh.IsError())
}

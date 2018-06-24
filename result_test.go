package vitotrol

import (
	"testing"

	td "github.com/maxatome/go-testdeep"
)

func TestResultHeader(tt *testing.T) {
	t := td.NewT(tt)

	rh := &ResultHeader{
		ErrorNum: 0,
		ErrorStr: "No error",
	}

	var _ error = rh

	t.CmpDeeply(rh.Error(), "No error [#0]")
	t.False(rh.IsError())

	rh = &ResultHeader{
		ErrorNum: 42,
		ErrorStr: "Big error",
	}
	t.CmpDeeply(rh.Error(), "Big error [#42]")
	t.True(rh.IsError())
}

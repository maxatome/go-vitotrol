package vitotrol

import (
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestTimeslot(t *testing.T) {
	assert := assert.New(t)

	ts := Timeslot{
		From: 9*100 + 11,
		To:   22*100 + 33,
	}

	assert.Equal("9:11 - 22:33", ts.String())

	tss := TimeslotSlice{
		{From: 5*100 + 34, To: 5*100 + 35},
		{From: 2*100 + 34, To: 2*100 + 35},
		{From: 4*100 + 34, To: 4*100 + 35},
		{From: 3*100 + 34, To: 3*100 + 35},
		{From: 0*100 + 34, To: 0*100 + 35},
	}
	sort.Sort(tss)

	assert.Equal(TimeslotSlice{
		{From: 0*100 + 34, To: 0*100 + 35},
		{From: 2*100 + 34, To: 2*100 + 35},
		{From: 3*100 + 34, To: 3*100 + 35},
		{From: 4*100 + 34, To: 4*100 + 35},
		{From: 5*100 + 34, To: 5*100 + 35},
	}, tss)
}

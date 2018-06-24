package vitotrol

import (
	"sort"
	"testing"

	td "github.com/maxatome/go-testdeep"
)

func TestTimeslot(tt *testing.T) {
	t := td.NewT(tt)

	ts := Timeslot{
		From: 9*100 + 11,
		To:   22*100 + 33,
	}

	t.CmpDeeply(ts.String(), "9:11 - 22:33")

	tss := TimeslotSlice{
		{From: 5*100 + 34, To: 5*100 + 35},
		{From: 2*100 + 34, To: 2*100 + 35},
		{From: 4*100 + 34, To: 4*100 + 35},
		{From: 3*100 + 34, To: 3*100 + 35},
		{From: 0*100 + 34, To: 0*100 + 35},
	}
	sort.Sort(tss)

	t.CmpDeeply(tss, TimeslotSlice{
		{From: 0*100 + 34, To: 0*100 + 35},
		{From: 2*100 + 34, To: 2*100 + 35},
		{From: 3*100 + 34, To: 3*100 + 35},
		{From: 4*100 + 34, To: 4*100 + 35},
		{From: 5*100 + 34, To: 5*100 + 35},
	})
}

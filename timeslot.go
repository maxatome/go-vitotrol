package vitotrol

import (
	"fmt"
)

// Timeslot represents a time slot. Hours and minutes are packed on 16
// bits by multiplying hours by 100 before adding them to minutes.
type Timeslot struct {
	From uint16 `json:"from"`
	To   uint16 `json:"to"`
}

// String returns a string representing the time slot.
func (t *Timeslot) String() string {
	return fmt.Sprintf("%d:%02d - %d:%02d",
		t.From/100, t.From%100,
		t.To/100, t.To%100)
}

type TimeslotSlice []Timeslot

func (t TimeslotSlice) Len() int           { return len(t) }
func (t TimeslotSlice) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t TimeslotSlice) Less(i, j int) bool { return t[i].From < t[j].From }

package vitotrol

import (
	"fmt"
)

// A TimesheetID allows to reference a specific timesheet. See
// *Timesheet consts.
type TimesheetID uint16

// All available/recognized TimesheetID values.
const (
	HotWaterLoopTimesheet TimesheetID = 7193 // Programmation bouclage ECS
	HotWaterTimesheet     TimesheetID = 7192 // Programmation ECS
	HeatingTimesheet      TimesheetID = 7191 // Programmation chauffage
)

// A TimesheetRef describe a time program reference.
type TimesheetRef struct {
	Name string
	Doc  string
}

// String returns a string describing a time program reference.
func (t *TimesheetRef) String() string {
	return fmt.Sprintf("%s: %s", t.Name, t.Doc)
}

// TimesheetsRef lists the reference for each timesheet ID.
var TimesheetsRef = map[TimesheetID]*TimesheetRef{
	HotWaterLoopTimesheet: {
		Name: "HotWaterLoopTimesheet",
		Doc:  "Time program for domestic hot water recirculation pump",
	},
	HotWaterTimesheet: {
		Name: "HotWaterTimesheet",
		Doc:  "Time program for domestic hot water heating",
	},
	HeatingTimesheet: {
		Name: "HeatingTimesheet",
		Doc:  "Time program for central heating",
	},
}

// TimesheetsNames2IDs maps the timesheet names to their TimesheetID
// counterpart.
var TimesheetsNames2IDs = func() map[string]TimesheetID {
	ret := make(map[string]TimesheetID, len(TimesheetsRef))
	for timesheetID, pTimesheetRef := range TimesheetsRef {
		ret[pTimesheetRef.Name] = timesheetID
	}
	return ret
}()

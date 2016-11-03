package vitotrol

import (
	"fmt"
)

type TimesheetId uint16

const (
	HotWaterLoopTimesheet TimesheetId = 7193 // Programmation bouclage ECS
	HotWaterTimesheet     TimesheetId = 7192 // Programmation ECS
	HeatingTimesheet      TimesheetId = 7191 // Programmation chauffage
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
var TimesheetsRef = map[TimesheetId]*TimesheetRef{
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

// TimesheetsNames2Ids maps the timesheet names to their TimesheetId
// counterpart.
var TimesheetsNames2Ids = func() map[string]TimesheetId {
	ret := make(map[string]TimesheetId, len(TimesheetsRef))
	for timesheetId, pTimesheetRef := range TimesheetsRef {
		ret[pTimesheetRef.Name] = timesheetId
	}
	return ret
}()

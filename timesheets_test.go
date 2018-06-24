package vitotrol

import (
	"testing"

	td "github.com/maxatome/go-testdeep"
)

func TestTimesheetRef(tt *testing.T) {
	t := td.NewT(tt)

	tsr := TimesheetRef{
		Name: "Foo",
		Doc:  "Documentation...",
	}
	t.CmpDeeply(tsr.String(), "Foo: Documentation...")
}

func TestTimesheetsVars(tt *testing.T) {
	t := td.NewT(tt)

	for name, tID := range TimesheetsNames2IDs {
		tRef, ok := TimesheetsRef[tID]
		if t.True(ok, "TimesheetsRef[%d] exists", tID) {
			t.CmpDeeply(name, tRef.Name,
				"ID %d: same name in TimesheetsRef and TimesheetsNames2IDs", tID)
		}
	}

	for tID, tRef := range TimesheetsRef {
		tID2, ok := TimesheetsNames2IDs[tRef.Name]
		if t.True(ok, "TimesheetsNames2IDs[%s] exists", tRef.Name) {
			t.CmpDeeply(tID, tID2,
				tRef.Name+": same ID in TimesheetsRef and TimesheetsNames2IDs")
		}
	}
}

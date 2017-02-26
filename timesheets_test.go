package vitotrol

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTimesheetRef(t *testing.T) {
	assert := assert.New(t)

	tsr := TimesheetRef{
		Name: "Foo",
		Doc:  "Documentation...",
	}
	assert.Equal("Foo: Documentation...", tsr.String())
}

func TestTimesheetsVars(t *testing.T) {
	assert := assert.New(t)

	for name, tID := range TimesheetsNames2IDs {
		tRef, ok := TimesheetsRef[tID]
		if assert.True(ok, "TimesheetsRef[%d] exists", tID) {
			assert.Equal(tRef.Name, name,
				"ID %d: same name in TimesheetsRef and TimesheetsNames2IDs", tID)
		}
	}

	for tID, tRef := range TimesheetsRef {
		tID2, ok := TimesheetsNames2IDs[tRef.Name]
		if assert.True(ok, "TimesheetsNames2IDs[%s] exists", tRef.Name) {
			assert.Equal(tID2, tID,
				tRef.Name+": same ID in TimesheetsRef and TimesheetsNames2IDs")
		}
	}
}

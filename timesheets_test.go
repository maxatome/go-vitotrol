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

	for name, tId := range TimesheetsNames2Ids {
		tRef, ok := TimesheetsRef[tId]
		if assert.True(ok, "TimesheetsRef[%d] exists", tId) {
			assert.Equal(tRef.Name, name,
				"ID %d: same name in TimesheetsRef and TimesheetsNames2Ids", tId)
		}
	}

	for tId, tRef := range TimesheetsRef {
		tId2, ok := TimesheetsNames2Ids[tRef.Name]
		if assert.True(ok, "TimesheetsNames2Ids[%s] exists", tRef.Name) {
			assert.Equal(tId2, tId,
				tRef.Name+": same ID in TimesheetsRef and TimesheetsNames2Ids")
		}
	}
}

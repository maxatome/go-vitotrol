package vitotrol

import (
	"encoding/xml"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	assert := assert.New(t)

	// UnmarshalXML
	type Value struct {
		XMLName xml.Name `xml:"Test"`
		Value   string   `xml:"Wert"`
		Time    Time     `xml:"Zeitstempel"`
	}
	var value Value
	err := xml.Unmarshal(
		[]byte(`
<Test>
  <Wert>1</Wert>
  <Zeitstempel>2016-10-30 22:57:18</Zeitstempel>
</Test>`),
		&value)
	if assert.Nil(err) {
		assert.Equal(
			Time(time.Date(2016, time.October, 30, 22, 57, 18, 0, vitodataTZ)),
			value.Time)
	}

	err = xml.Unmarshal(
		[]byte(`
<Test>
  <Wert>1</Wert>
  <Zeitstempel>
</Test>`),
		&value)
	assert.NotNil(err)

	// String
	tm := Time(time.Date(2016, time.October, 30, 22, 57, 18, 0, vitodataTZ))
	assert.Equal("2016-10-30 22:57:18", tm.String())

	tm2, err := ParseVitotrolTime("2016-10-30 22:57:18")
	assert.Nil(err)
	assert.Equal(tm, tm2)

	_, err = ParseVitotrolTime("2016-10-30 22:57:foo")
	assert.NotNil(err)
}

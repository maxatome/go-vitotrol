package vitotrol

import (
	"encoding/xml"
	"testing"
	"time"

	td "github.com/maxatome/go-testdeep"
)

func TestTime(tt *testing.T) {
	t := td.NewT(tt)

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
	if t.CmpNoError(err) {
		t.CmpDeeply(
			value.Time,
			Time(time.Date(2016, time.October, 30, 22, 57, 18, 0, vitodataTZ)))
	}

	err = xml.Unmarshal(
		[]byte(`
<Test>
  <Wert>1</Wert>
  <Zeitstempel>
</Test>`),
		&value)
	t.CmpError(err)

	// String
	tm := Time(time.Date(2016, time.October, 30, 22, 57, 18, 0, vitodataTZ))
	t.CmpDeeply(tm.String(), "2016-10-30 22:57:18")

	tm2, err := ParseVitotrolTime("2016-10-30 22:57:18")
	if t.CmpNoError(err) {
		t.CmpDeeply(tm2, tm)
	}

	_, err = ParseVitotrolTime("2016-10-30 22:57:foo")
	t.CmpError(err)
}

package vitotrol

import (
	"encoding/xml"
	"time"
)

const vitodataTimeFormat = "2006-01-02 15:04:05"

var vitodataTZ = time.Local

// Time handle the Vitotrol™ time format.
type Time time.Time

// UnmarshalXML decodes a Vitotrol™ time embedded in XML.
func (t *Time) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var str string
	err := d.DecodeElement(&str, &start)
	if err != nil {
		return err
	}

	*t, err = ParseVitotrolTime(str)
	return err
}

// String returns the time formatted using the format string
//	2006-01-02 15:04:05
// considered as being a localtime value.
func (t Time) String() string {
	return time.Time(t).Format(vitodataTimeFormat)
}

// ParseVitotrolTime parses a Vitotrol™ time information. Without a
// time zone it is considered as a local time.
func ParseVitotrolTime(value string) (Time, error) {
	tm, err := time.ParseInLocation(vitodataTimeFormat, value, vitodataTZ)
	return Time(tm), err
}

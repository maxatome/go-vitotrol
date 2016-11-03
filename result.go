package vitotrol

import (
	"fmt"
)

// ResultHeader included in each Result part of each Vitotrol™ Response
// message.
type ResultHeader struct {
	ErrorNum int    `xml:"Ergebnis"`
	ErrorStr string `xml:"ErgebnisText"`
}

// Error returns the result as a string.
func (e *ResultHeader) Error() string {
	return fmt.Sprintf("%s [#%d]", e.ErrorStr, e.ErrorNum)
}

// IsError allows to know if this result is an error or not from the
// Vitotrol™ point of view.
func (e *ResultHeader) IsError() bool {
	return e.ErrorNum != 0
}

// HasResultHeader is the interface for abstrating Result part of each
// Vitotrol™ Response message.
type HasResultHeader interface {
	ResultHeader() *ResultHeader
}

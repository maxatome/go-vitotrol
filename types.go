package vitotrol

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Singletons matching Vitodata™ types.
var (
	TypeDouble    = (*VitodataDouble)(nil)
	TypeInteger   = (*VitodataInteger)(nil)
	TypeDate      = (*VitodataDate)(nil)
	TypeString    = (*VitodataString)(nil)
	TypeOnOffEnum = NewEnum([]string{ // 0 -> 1
		"off",
		"on",
	})
	TypeEnabledEnum = NewEnum([]string{ // 0 -> 1
		"disabled",
		"enabled",
	})

	TypeNames = map[string]VitodataType{
		TypeDouble.Type():  TypeDouble,
		TypeInteger.Type(): TypeInteger,
		TypeDate.Type():    TypeDate,
		TypeString.Type():  TypeString,
	}
)

// ErrEnumInvalidValue is returned when trying to convert to an enum a
// value that cannot match any value of this enum.
var ErrEnumInvalidValue = errors.New("Invalid Enum value")

// VitodataType is the interface implemented by each Vitodata™ type.
type VitodataType interface {
	Type() string
	Human2VitodataValue(string) (string, error)
	Vitodata2HumanValue(string) (string, error)
	Vitodata2NativeValue(string) (interface{}, error)
}

// A VitodataDouble represent the Vitodata™ Double type.
type VitodataDouble struct{}

// Type returns the "human" name of the type.
func (v *VitodataDouble) Type() string {
	return "Double"
}

// Human2VitodataValue checks that the value is a float number and
// returns it after reformatting.
func (v *VitodataDouble) Human2VitodataValue(value string) (string, error) {
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return "", err
	}
	return strings.Replace(strconv.FormatFloat(num, 'f', -1, 64), ".", ",", 1), nil
}

// Vitodata2HumanValue checks that the value is a float number and
// returns it after reformatting.
func (v *VitodataDouble) Vitodata2HumanValue(value string) (string, error) {
	num, err := strconv.ParseFloat(strings.Replace(value, ",", ".", 1), 64)
	if err != nil {
		return "", err
	}
	return strconv.FormatFloat(num, 'f', -1, 64), nil
}

// Vitodata2NativeValue extract the number from the passed string and
// returns it as a float64.
func (v *VitodataDouble) Vitodata2NativeValue(value string) (interface{}, error) {
	num, err := strconv.ParseFloat(strings.Replace(value, ",", ".", 1), 64)
	if err != nil {
		return nil, err
	}
	return num, nil
}

// A VitodataInteger represent the Vitodata™ Integer type.
type VitodataInteger struct{}

// Type returns the "human" name of the type.
func (v *VitodataInteger) Type() string {
	return "Integer"
}

// Human2VitodataValue checks that the value is an integer and returns
// it after reformatting.
func (v *VitodataInteger) Human2VitodataValue(value string) (string, error) {
	num, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(num, 10), nil
}

// Vitodata2HumanValue checks that the value is an integer and returns
// it after reformatting.
func (v *VitodataInteger) Vitodata2HumanValue(value string) (string, error) {
	return v.Human2VitodataValue(value)
}

// Vitodata2NativeValue extract the number from the passed string and
// returns it as an int64.
func (v *VitodataInteger) Vitodata2NativeValue(value string) (interface{}, error) {
	num, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, err
	}
	return num, nil
}

// A VitodataDate represent the Vitodata™ Date type.
type VitodataDate struct{}

// Type returns the "human" name of the type.
func (v *VitodataDate) Type() string {
	return "Date"
}

// Human2VitodataValue checks that the value is Vitodata™ formatted date and
// returns it after reformatting.
func (v *VitodataDate) Human2VitodataValue(value string) (string, error) {
	tm, err := ParseVitotrolTime(value)
	if err != nil {
		return "", err
	}
	return tm.String(), nil
}

// Vitodata2HumanValue checks that the value is Vitodata™ formatted date and
// returns it after reformatting.
func (v *VitodataDate) Vitodata2HumanValue(value string) (string, error) {
	return v.Human2VitodataValue(value)
}

// Vitodata2NativeValue extract the Vitodata™ date from
// the passed string and returns it as a vitotrol.Time.
func (v *VitodataDate) Vitodata2NativeValue(value string) (interface{}, error) {
	tm, err := ParseVitotrolTime(value)
	if err != nil {
		return nil, err
	}
	return tm, nil
}

// A VitodataString represent the Vitodata™ String type.
type VitodataString struct{}

// Type returns the "human" name of the type.
func (v *VitodataString) Type() string {
	return "String"
}

// Human2VitodataValue is a no-op here, returning its argument.
func (v *VitodataString) Human2VitodataValue(value string) (string, error) {
	return value, nil
}

// Vitodata2HumanValue is a no-op here, returning its argument.
func (v *VitodataString) Vitodata2HumanValue(value string) (string, error) {
	return value, nil
}

// Vitodata2NativeValue is a no-op here, returning its argument.
func (v *VitodataString) Vitodata2NativeValue(value string) (interface{}, error) {
	return value, nil
}

// VitodataEnum represents any Vitodata™ Enum type. See NewEnum to
// specialize it.
type VitodataEnum struct {
	values    map[string]uint16
	revValues []string
}

// NewEnum specializes an enum to a set of values and returns it.
func NewEnum(values []string) *VitodataEnum {
	pEnum := &VitodataEnum{
		values:    make(map[string]uint16, len(values)),
		revValues: make([]string, len(values)),
	}

	for idx, value := range values {
		pEnum.values[value] = uint16(idx)
		pEnum.revValues[idx] = value
	}

	return pEnum
}

// Type returns the "human" name of the type.
func (v *VitodataEnum) Type() string {
	return fmt.Sprintf("Enum%d", len(v.revValues))
}

// Human2VitodataValue checks that the value is a Vitodata™ enum value
// and returns its numeric counterpart.
func (v *VitodataEnum) Human2VitodataValue(value string) (string, error) {
	// String version ?
	if num, ok := v.values[value]; ok {
		return strconv.FormatUint(uint64(num), 10), nil
	}

	// Numeric one ?
	num, err := v.Vitodata2NativeValue(value)
	if err != nil {
		return "", err
	}

	return strconv.FormatUint(num.(uint64), 10), nil
}

// Vitodata2HumanValue check that the (numeric) value is a Vitodata™
// enum value and returns its string counterpart.
func (v *VitodataEnum) Vitodata2HumanValue(value string) (string, error) {
	num, err := v.Vitodata2NativeValue(value)
	if err != nil {
		return "", err
	}
	return v.revValues[num.(uint64)], nil
}

// Vitodata2NativeValue extract the numeric Vitodata™ enum value from
// the passed string and returns it as a uint64.
func (v *VitodataEnum) Vitodata2NativeValue(value string) (interface{}, error) {
	num, err := strconv.ParseUint(value, 10, 64)
	if err != nil || num >= uint64(len(v.revValues)) {
		return nil, ErrEnumInvalidValue
	}
	return num, nil
}

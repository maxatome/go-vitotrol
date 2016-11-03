package vitotrol

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	TypeDouble    = (*VitodataDouble)(nil)
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
)

var EnumInvalidValue = errors.New("Invalid Enum value")

// VitodataType is the interface implemented by each Vitodata™ type.
type VitodataType interface {
	Type() string
	Human2VitodataValue(string) (string, error)
	Vitodata2HumanValue(string) (string, error)
	Vitodata2NativeValue(string) (interface{}, error)
}

// A VitodataDouble represent the Vitodata™ Double type.
type VitodataDouble struct{}

func (v *VitodataDouble) Type() string {
	return "Double"
}

func (v *VitodataDouble) Human2VitodataValue(value string) (string, error) {
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return "", err
	}
	return strconv.FormatFloat(num, 'f', -1, 64), nil
}

func (v *VitodataDouble) Vitodata2HumanValue(value string) (string, error) {
	return v.Human2VitodataValue(value)
}

func (v *VitodataDouble) Vitodata2NativeValue(value string) (interface{}, error) {
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, err
	}
	return num, nil
}

// A VitodataDate represent the Vitodata™ Date type.
type VitodataDate struct{}

func (v *VitodataDate) Type() string {
	return "Date"
}

func (v *VitodataDate) Human2VitodataValue(value string) (string, error) {
	tm, err := ParseVitotrolTime(value)
	if err != nil {
		return "", err
	}
	return tm.String(), nil
}

func (v *VitodataDate) Vitodata2HumanValue(value string) (string, error) {
	return v.Human2VitodataValue(value)
}

func (v *VitodataDate) Vitodata2NativeValue(value string) (interface{}, error) {
	tm, err := ParseVitotrolTime(value)
	if err != nil {
		return nil, err
	}
	return tm, nil
}

// A VitodataString represent the Vitodata™ String type.
type VitodataString struct{}

func (v *VitodataString) Type() string {
	return "String"
}

func (v *VitodataString) Human2VitodataValue(value string) (string, error) {
	return value, nil
}

func (v *VitodataString) Vitodata2HumanValue(value string) (string, error) {
	return value, nil
}

func (v *VitodataString) Vitodata2NativeValue(value string) (interface{}, error) {
	return value, nil
}

// A vitodataString represent any Vitodata™ Enum type. It must be composed.
type vitodataEnum struct {
	values    map[string]uint16
	revValues []string
}

func NewEnum(values []string) *vitodataEnum {
	pEnum := &vitodataEnum{
		values:    make(map[string]uint16, len(values)),
		revValues: make([]string, len(values)),
	}

	for idx, value := range values {
		pEnum.values[value] = uint16(idx)
		pEnum.revValues[idx] = value
	}

	return pEnum
}

func (v *vitodataEnum) Type() string {
	return fmt.Sprintf("Enum%d", len(v.revValues))
}

func (v *vitodataEnum) Human2VitodataValue(value string) (string, error) {
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

func (v *vitodataEnum) Vitodata2HumanValue(value string) (string, error) {
	num, err := v.Vitodata2NativeValue(value)
	if err != nil {
		return "", err
	}
	return v.revValues[num.(uint64)], nil
}

func (v *vitodataEnum) Vitodata2NativeValue(value string) (interface{}, error) {
	num, err := strconv.ParseUint(value, 10, 64)
	if err != nil || num >= uint64(len(v.revValues)) {
		return nil, EnumInvalidValue
	}
	return num, nil
}

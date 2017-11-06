package vitotrol

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestVitodataDouble(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("Double", TypeDouble.Type())

	// Human2VitodataValue
	str, err := TypeDouble.Human2VitodataValue("1.200")
	assert.Equal("1.2", str)
	assert.Nil(err)

	str, err = TypeDouble.Human2VitodataValue("foo")
	assert.Equal("", str)
	assert.NotNil(err)

	// Vitodata2HumanValue
	str, err = TypeDouble.Vitodata2HumanValue("1.200")
	assert.Equal("1.2", str)
	assert.Nil(err)

	str, err = TypeDouble.Vitodata2HumanValue("foo")
	assert.Equal("", str)
	assert.NotNil(err)

	// Vitodata2NativeValue
	num, err := TypeDouble.Vitodata2NativeValue("1.200")
	assert.Equal(float64(1.2), num)
	assert.Nil(err)

	num, err = TypeDouble.Vitodata2NativeValue("foo")
	assert.Nil(num)
	assert.NotNil(err)
}

func TestVitodataInteger(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("Integer", TypeInteger.Type())

	// Human2VitodataValue
	str, err := TypeInteger.Human2VitodataValue("12")
	assert.Equal("12", str)
	assert.Nil(err)

	str, err = TypeInteger.Human2VitodataValue("foo")
	assert.Equal("", str)
	assert.NotNil(err)

	// Vitodata2HumanValue
	str, err = TypeInteger.Vitodata2HumanValue("00012")
	assert.Equal("12", str)
	assert.Nil(err)

	str, err = TypeInteger.Vitodata2HumanValue("foo")
	assert.Equal("", str)
	assert.NotNil(err)

	// Vitodata2NativeValue
	num, err := TypeInteger.Vitodata2NativeValue("00012")
	assert.Equal(int64(12), num)
	assert.Nil(err)

	num, err = TypeInteger.Vitodata2NativeValue("foo")
	assert.Nil(num)
	assert.NotNil(err)
}

func TestVitodataDate(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("Date", TypeDate.Type())

	const refDate = "2016-09-26 11:22:33"

	// Human2VitodataValue
	str, err := TypeDate.Human2VitodataValue(refDate)
	assert.Equal(refDate, str)
	assert.Nil(err)

	str, err = TypeDate.Human2VitodataValue("foo")
	assert.Equal("", str)
	assert.NotNil(err)

	// Vitodata2HumanValue
	str, err = TypeDate.Vitodata2HumanValue(refDate)
	assert.Equal(refDate, str)
	assert.Nil(err)

	str, err = TypeDate.Vitodata2HumanValue("foo")
	assert.Equal("", str)
	assert.NotNil(err)

	// Vitodata2NativeValue
	date, err := TypeDate.Vitodata2NativeValue(refDate)
	assert.Equal(
		Time(time.Date(2016, time.September, 26, 11, 22, 33, 0, vitodataTZ)),
		date)
	assert.Nil(err)

	date, err = TypeDate.Vitodata2NativeValue("foo")
	assert.Nil(date)
	assert.NotNil(err)
}

func TestVitodataString(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("String", TypeString.Type())

	const refString = "foobar"

	// Human2VitodataValue
	str, err := TypeString.Human2VitodataValue(refString)
	assert.Equal(refString, str)
	assert.Nil(err)

	// Vitodata2HumanValue
	str, err = TypeString.Vitodata2HumanValue(refString)
	assert.Equal(refString, str)
	assert.Nil(err)

	// Vitodata2NativeValue
	strIf, err := TypeString.Vitodata2NativeValue(refString)
	assert.Equal(refString, strIf)
	assert.Nil(err)
}

func TestVitodataEnum(t *testing.T) {
	assert := assert.New(t)

	typeEnumTest := NewEnum([]string{
		"zero",
		"one",
		"two",
	})

	assert.Equal("Enum3", typeEnumTest.Type())

	// Human2VitodataValue
	str, err := typeEnumTest.Human2VitodataValue("one")
	assert.Equal("1", str)
	assert.Nil(err)

	str, err = typeEnumTest.Human2VitodataValue("1")
	assert.Equal("1", str)
	assert.Nil(err)

	str, err = typeEnumTest.Human2VitodataValue("foo")
	assert.Equal("", str)
	assert.NotNil(err)

	// Vitodata2HumanValue
	str, err = typeEnumTest.Vitodata2HumanValue("2")
	assert.Equal("two", str)
	assert.Nil(err)

	str, err = typeEnumTest.Vitodata2HumanValue("42")
	assert.Equal("", str)
	assert.Equal(ErrEnumInvalidValue, err)

	str, err = typeEnumTest.Vitodata2HumanValue("foo")
	assert.Equal("", str)
	assert.Equal(ErrEnumInvalidValue, err)

	// Vitodata2NativeValue
	num, err := typeEnumTest.Vitodata2NativeValue("1")
	assert.Equal(uint64(1), num)
	assert.Nil(err)

	num, err = typeEnumTest.Vitodata2NativeValue("42")
	assert.Nil(num)
	assert.Equal(ErrEnumInvalidValue, err)

	num, err = typeEnumTest.Vitodata2NativeValue("foo")
	assert.Nil(num)
	assert.Equal(ErrEnumInvalidValue, err)
}

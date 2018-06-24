package vitotrol

import (
	"testing"
	"time"

	td "github.com/maxatome/go-testdeep"
)

func TestVitodataDouble(tt *testing.T) {
	t := td.NewT(tt)

	t.CmpDeeply(TypeDouble.Type(), "Double")

	// Human2VitodataValue
	str, err := TypeDouble.Human2VitodataValue("1.200")
	t.CmpDeeply(str, "1.2")
	t.CmpNoError(err)

	str, err = TypeDouble.Human2VitodataValue("foo")
	t.Empty(str)
	t.CmpError(err)

	// Vitodata2HumanValue
	str, err = TypeDouble.Vitodata2HumanValue("1.200")
	t.CmpDeeply(str, "1.2")
	t.CmpNoError(err)

	str, err = TypeDouble.Vitodata2HumanValue("foo")
	t.Empty(str)
	t.CmpError(err)

	// Vitodata2NativeValue
	num, err := TypeDouble.Vitodata2NativeValue("1.200")
	t.CmpDeeply(num, float64(1.2))
	t.CmpNoError(err)

	num, err = TypeDouble.Vitodata2NativeValue("foo")
	t.Nil(num)
	t.CmpError(err)
}

func TestVitodataInteger(tt *testing.T) {
	t := td.NewT(tt)

	t.CmpDeeply(TypeInteger.Type(), "Integer")

	// Human2VitodataValue
	str, err := TypeInteger.Human2VitodataValue("12")
	t.CmpDeeply(str, "12")
	t.CmpNoError(err)

	str, err = TypeInteger.Human2VitodataValue("foo")
	t.Empty(str)
	t.CmpError(err)

	// Vitodata2HumanValue
	str, err = TypeInteger.Vitodata2HumanValue("00012")
	t.CmpDeeply(str, "12")
	t.CmpNoError(err)

	str, err = TypeInteger.Vitodata2HumanValue("foo")
	t.Empty(str)
	t.CmpError(err)

	// Vitodata2NativeValue
	num, err := TypeInteger.Vitodata2NativeValue("00012")
	t.CmpDeeply(num, int64(12))
	t.CmpNoError(err)

	num, err = TypeInteger.Vitodata2NativeValue("foo")
	t.Nil(num)
	t.CmpError(err)
}

func TestVitodataDate(tt *testing.T) {
	t := td.NewT(tt)

	t.CmpDeeply(TypeDate.Type(), "Date")

	const refDate = "2016-09-26 11:22:33"

	// Human2VitodataValue
	str, err := TypeDate.Human2VitodataValue(refDate)
	t.CmpDeeply(str, refDate)
	t.CmpNoError(err)

	str, err = TypeDate.Human2VitodataValue("foo")
	t.Empty(str)
	t.CmpError(err)

	// Vitodata2HumanValue
	str, err = TypeDate.Vitodata2HumanValue(refDate)
	t.CmpDeeply(str, refDate)
	t.CmpNoError(err)

	str, err = TypeDate.Vitodata2HumanValue("foo")
	t.Empty(str)
	t.CmpError(err)

	// Vitodata2NativeValue
	date, err := TypeDate.Vitodata2NativeValue(refDate)
	t.CmpDeeply(
		date,
		Time(time.Date(2016, time.September, 26, 11, 22, 33, 0, vitodataTZ)))
	t.CmpNoError(err)

	date, err = TypeDate.Vitodata2NativeValue("foo")
	t.Nil(date)
	t.CmpError(err)
}

func TestVitodataString(tt *testing.T) {
	t := td.NewT(tt)

	t.CmpDeeply(TypeString.Type(), "String")

	const refString = "foobar"

	// Human2VitodataValue
	str, err := TypeString.Human2VitodataValue(refString)
	t.CmpDeeply(str, refString)
	t.CmpNoError(err)

	// Vitodata2HumanValue
	str, err = TypeString.Vitodata2HumanValue(refString)
	t.CmpDeeply(str, refString)
	t.CmpNoError(err)

	// Vitodata2NativeValue
	strIf, err := TypeString.Vitodata2NativeValue(refString)
	t.CmpDeeply(strIf, refString)
	t.CmpNoError(err)
}

func TestVitodataEnum(tt *testing.T) {
	t := td.NewT(tt)

	typeEnumTest := NewEnum([]string{
		"zero",
		"one",
		"two",
	})

	t.CmpDeeply(typeEnumTest.Type(), "Enum3")

	// Human2VitodataValue
	str, err := typeEnumTest.Human2VitodataValue("one")
	t.CmpDeeply(str, "1")
	t.CmpNoError(err)

	str, err = typeEnumTest.Human2VitodataValue("1")
	t.CmpDeeply(str, "1")
	t.CmpNoError(err)

	str, err = typeEnumTest.Human2VitodataValue("foo")
	t.Empty(str)
	t.CmpError(err)

	// Vitodata2HumanValue
	str, err = typeEnumTest.Vitodata2HumanValue("2")
	t.CmpDeeply(str, "two")
	t.CmpNoError(err)

	str, err = typeEnumTest.Vitodata2HumanValue("42")
	t.Empty(str)
	t.CmpDeeply(err, ErrEnumInvalidValue)

	str, err = typeEnumTest.Vitodata2HumanValue("foo")
	t.Empty(str)
	t.CmpDeeply(err, ErrEnumInvalidValue)

	// Vitodata2NativeValue
	num, err := typeEnumTest.Vitodata2NativeValue("1")
	t.CmpDeeply(num, uint64(1))
	t.CmpNoError(err)

	num, err = typeEnumTest.Vitodata2NativeValue("42")
	t.Nil(num)
	t.CmpDeeply(err, ErrEnumInvalidValue)

	num, err = typeEnumTest.Vitodata2NativeValue("foo")
	t.Nil(num)
	t.CmpDeeply(err, ErrEnumInvalidValue)
}

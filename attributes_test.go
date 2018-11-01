package vitotrol

import (
	"testing"

	td "github.com/maxatome/go-testdeep"
)

func TestAttrRef(tt *testing.T) {
	t := td.NewT(tt)

	ar := AttrRef{
		Type:   TypeString,
		Access: ReadOnly,
		Name:   "Foo",
		Doc:    "Documentation...",
	}
	t.CmpDeeply(ar.String(), "Foo: Documentation... (String - read-only)")
}

func TestAttributesVars(tt *testing.T) {
	t := td.NewT(tt)

	for name, aID := range AttributesNames2IDs {
		tRef, ok := AttributesRef[aID]
		if t.True(ok, "AttributesRef[%d] exists", aID) {
			t.CmpDeeply(name, tRef.Name,
				"ID %d: same name in AttributesRef and AttributesNames2IDs", aID)
		}
	}

	for aID, tRef := range AttributesRef {
		aID2, ok := AttributesNames2IDs[tRef.Name]
		if t.True(ok, "AttributesNames2IDs[%s] exists", tRef.Name) {
			t.CmpDeeply(aID, aID2,
				tRef.Name+": same ID in AttributesRef and AttributesNames2IDs")
		}
	}

	t.CmpDeeply(len(Attributes), len(AttributesRef))
	for _, aID := range Attributes {
		_, ok := AttributesRef[aID]
		t.True(ok, "ID %d of Attributes is present in AttributesRef")
	}

	// Add custom attribute
	require := t.FailureIsFatal()

	require.CmpDeeply(AttributesRef, td.Not(td.ContainsKey(9999)))

	AddAttributeRef(9999, AttrRef{
		Type:   TypeDouble,
		Access: ReadOnly,
		Name:   "FooBar",
	})

	require.CmpDeeply(AttributesRef, td.ContainsKey(AttrID(9999)))

	t.CmpDeeply(AttributesNames2IDs["FooBar"], AttrID(9999))
	t.CmpDeeply(Attributes, td.Contains(AttrID(9999)))

	t.True(AttributesRef[AttrID(9999)].Custom)
}

func TestValue(tt *testing.T) {
	t := td.NewT(tt)

	v := Value{Value: "34"}
	t.CmpDeeply(v.Num(), float64(34))

	v = Value{Value: "foo"}
	t.CmpDeeply(v.Num(), float64(0))
}

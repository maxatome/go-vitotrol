package vitotrol

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAttrRef(t *testing.T) {
	assert := assert.New(t)

	ar := AttrRef{
		Type:   TypeString,
		Access: ReadOnly,
		Name:   "Foo",
		Doc:    "Documentation...",
	}
	assert.Equal("Foo: Documentation... (String - read-only)", ar.String())
}

func TestAttributesVars(t *testing.T) {
	assert := assert.New(t)

	for name, aID := range AttributesNames2IDs {
		tRef, ok := AttributesRef[aID]
		if assert.True(ok, "AttributesRef[%d] exists", aID) {
			assert.Equal(tRef.Name, name,
				"ID %d: same name in AttributesRef and AttributesNames2IDs", aID)
		}
	}

	for aID, tRef := range AttributesRef {
		aID2, ok := AttributesNames2IDs[tRef.Name]
		if assert.True(ok, "AttributesNames2IDs[%s] exists", tRef.Name) {
			assert.Equal(aID2, aID,
				tRef.Name+": same ID in AttributesRef and AttributesNames2IDs")
		}
	}

	assert.Equal(len(AttributesRef), len(Attributes))
	for _, aID := range Attributes {
		_, ok := AttributesRef[aID]
		assert.True(ok, "ID %d of Attributes is present in AttributesRef")
	}
}

func TestValue(t *testing.T) {
	assert := assert.New(t)

	v := Value{Value: "34"}
	assert.Equal(float64(34), v.Num())

	v = Value{Value: "foo"}
	assert.Equal(float64(0), v.Num())
}

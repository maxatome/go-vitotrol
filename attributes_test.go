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

	for name, aId := range AttributesNames2Ids {
		tRef, ok := AttributesRef[aId]
		if assert.True(ok, "AttributesRef[%d] exists", aId) {
			assert.Equal(tRef.Name, name,
				"ID %d: same name in AttributesRef and AttributesNames2Ids", aId)
		}
	}

	for aId, tRef := range AttributesRef {
		aId2, ok := AttributesNames2Ids[tRef.Name]
		if assert.True(ok, "AttributesNames2Ids[%s] exists", tRef.Name) {
			assert.Equal(aId2, aId,
				tRef.Name+": same ID in AttributesRef and AttributesNames2Ids")
		}
	}

	assert.Equal(len(AttributesRef), len(Attributes))
	for _, aId := range Attributes {
		_, ok := AttributesRef[aId]
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

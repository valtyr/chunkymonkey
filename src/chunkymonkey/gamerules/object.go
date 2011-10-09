// Defines non-block movable objects such as arrows in flight, boats and
// minecarts.

package gamerules

import (
	"io"
	"os"

	"chunkymonkey/physics"
	"chunkymonkey/proto"
	"chunkymonkey/types"
	"nbt"
)

// TODO Object sub-types?

type Object struct {
	types.EntityId
	types.ObjTypeId
	physics.PointObject
	orientation types.OrientationBytes
}

func NewObject(objType types.ObjTypeId) (object *Object) {
	object = &Object{
		// TODO: proper orientation
		orientation: types.OrientationBytes{0, 0, 0},
	}
	object.ObjTypeId = objType
	return
}

func (object *Object) UnmarshalNbt(tag *nbt.Compound) (err os.Error) {
	if err = object.PointObject.UnmarshalNbt(tag); err != nil {
		return
	}

	var typeName string
	if entityObjectId, ok := tag.Lookup("id").(*nbt.String); !ok {
		return os.NewError("missing object type id")
	} else {
		typeName = entityObjectId.Value
	}

	var ok bool
	if object.ObjTypeId, ok = types.ObjTypeByName[typeName]; !ok {
		return os.NewError("unknown object type id")
	}

	// TODO load orientation

	return
}

func (object *Object) MarshalNbt(tag *nbt.Compound) (err os.Error) {
	objTypeName, ok := types.ObjNameByType[object.ObjTypeId]
	if !ok {
		return os.NewError("unknown object type")
	}
	if err = object.PointObject.MarshalNbt(tag); err != nil {
		return
	}
	tag.Set("id", &nbt.String{objTypeName})
	// TODO unknown fields
	return
}

func (object *Object) SendSpawn(writer io.Writer) (err os.Error) {
	// TODO: Send non-nil ObjectData (is there any?)
	err = proto.WriteObjectSpawn(writer, object.EntityId, object.ObjTypeId, &object.PointObject.LastSentPosition, nil)
	if err != nil {
		return
	}

	err = proto.WriteEntityVelocity(writer, object.EntityId, &object.PointObject.LastSentVelocity)
	return
}

func (object *Object) SendUpdate(writer io.Writer) (err os.Error) {
	if err = proto.WriteEntity(writer, object.EntityId); err != nil {
		return
	}

	// TODO: Should this be the Rotation information?
	err = object.PointObject.SendUpdate(writer, object.EntityId, &types.LookBytes{0, 0})

	return
}

func NewBoat() INonPlayerEntity {
	return NewObject(types.ObjTypeIdBoat)
}

func NewMinecart() INonPlayerEntity {
	return NewObject(types.ObjTypeIdMinecart)
}

func NewStorageCart() INonPlayerEntity {
	return NewObject(types.ObjTypeIdStorageCart)
}

func NewPoweredCart() INonPlayerEntity {
	return NewObject(types.ObjTypeIdPoweredCart)
}

func NewActivatedTnt() INonPlayerEntity {
	return NewObject(types.ObjTypeIdActivatedTnt)
}

func NewArrow() INonPlayerEntity {
	return NewObject(types.ObjTypeIdArrow)
}

func NewThrownSnowball() INonPlayerEntity {
	return NewObject(types.ObjTypeIdThrownSnowball)
}

func NewThrownEgg() INonPlayerEntity {
	return NewObject(types.ObjTypeIdThrownEgg)
}

func NewFallingSand() INonPlayerEntity {
	return NewObject(types.ObjTypeIdFallingSand)
}

func NewFallingGravel() INonPlayerEntity {
	return NewObject(types.ObjTypeIdFallingGravel)
}

func NewFishingFloat() INonPlayerEntity {
	return NewObject(types.ObjTypeIdFishingFloat)
}

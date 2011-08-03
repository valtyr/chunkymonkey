// Defines non-block movable objecs such as arrows in flight, boats and
// minecarts.

package gamerules

import (
	"io"
	"os"

	"chunkymonkey/physics"
	"chunkymonkey/proto"
	. "chunkymonkey/types"
	"nbt"
)

// TODO Object sub-types?

type Object struct {
	EntityId
	ObjTypeId
	physics.PointObject
	orientation OrientationBytes
}

func NewObject(objType ObjTypeId) (object *Object) {
	object = &Object{
		// TODO: proper orientation
		orientation: OrientationBytes{0, 0, 0},
	}
	object.ObjTypeId = objType
	return
}

func (object *Object) ReadNbt(tag nbt.ITag) (err os.Error) {
	if err = object.PointObject.ReadNbt(tag); err != nil {
		return
	}

	var typeName string
	if entityObjectId, ok := tag.Lookup("id").(*nbt.String); !ok {
		return os.NewError("missing object type id")
	} else {
		typeName = entityObjectId.Value
	}

	var ok bool
	if object.ObjTypeId, ok = ObjTypeMap[typeName]; !ok {
		return os.NewError("unknown object type id")
	}

	// TODO load orientation

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
	err = object.PointObject.SendUpdate(writer, object.EntityId, &LookBytes{0, 0})

	return
}

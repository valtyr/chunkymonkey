// Defines interfaces for entities in the world (including pick up items,
// mobs, players, and non-block objecs such as arrows in flight, boats and
// minecarts.
package object

import (
	"io"
	"os"

	"chunkymonkey/physics"
	"chunkymonkey/proto"
	. "chunkymonkey/types"
)

// ISpawn represents common elements to all types of entities that can be
// present in a chunk.
type IEntity interface {
	GetEntityId() EntityId
	SendSpawn(io.Writer) os.Error
	SendUpdate(io.Writer) os.Error
	Position() *AbsXyz
}

type INonPlayerEntity interface {
	IEntity
	SetEntityId(EntityId)
	Tick(physics.BlockQueryFn) (leftBlock bool)
}

type Object struct {
	EntityId
	ObjTypeId
	physics.PointObject
	orientation OrientationBytes
}

func NewObject(objType ObjTypeId, position *AbsXyz, velocity *AbsVelocity) (object *Object) {
	object = &Object{
		// TODO: proper orientation
		orientation: OrientationBytes{0, 0, 0},
	}
	object.ObjTypeId = objType
	object.PointObject.Init(position, velocity)
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

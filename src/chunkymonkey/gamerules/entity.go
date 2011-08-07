// Defines interfaces for entities in the world (including pick up items,
// mobs, players, and other non-block objects).

package gamerules

import (
	"io"
	"os"

	"chunkymonkey/physics"
	. "chunkymonkey/types"
	"nbt"
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
	ReadNbt(nbt.ITag) os.Error
	// WriteNbt creates an NBT tag representing the entity. This can be nil if
	// the entity cannot be serialized.
	WriteNbt() nbt.ITag
	SetEntityId(EntityId)
	Tick(physics.IBlockQuerier) (leftBlock bool)
}

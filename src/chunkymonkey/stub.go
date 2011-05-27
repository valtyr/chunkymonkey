package stub

import (
	"io"
	"os"

	"chunkymonkey/entity"
	"chunkymonkey/physics"
	"chunkymonkey/slot"
	. "chunkymonkey/types"
)

// ISpawn represents common elements to all types of entities that can be
// present in a chunk.
type ISpawn interface {
	GetEntityId() EntityId
	SendSpawn(io.Writer) os.Error
	SendUpdate(io.Writer) os.Error
	Position() *AbsXyz
}

type INonPlayerSpawn interface {
	ISpawn
	GetEntity() *entity.Entity
	Tick(physics.BlockQueryFn) (leftBlock bool)
}

// IPlayerConnection is the interface by which shards communicate to players on
// the frontend.
type IPlayerConnection interface {
	GetEntityId() EntityId

	TransmitPacket(packet []byte)

	InventorySubscribed(shardInvId int32, invTypeId InvTypeId)

	InventoryUpdate(shardInvId int32, slotIds []SlotId, slots []slot.Slot)

	// RequestPlaceHeldItem requests that the player frontend take one item from
	// the held item stack and send it in a RequestPlaceItem to the target block.
	// The player code may *not* honour this request (e.g there might be no
	// suitable held item).
	RequestPlaceHeldItem(target BlockXyz, wasHeld slot.Slot)

	// RequestOfferItem requests that the player check if it can take the item.
	// If it can then it should RequestTakeItem from the chunk.
	RequestOfferItem(fromChunk ChunkXz, entityId EntityId, item slot.Slot)

	RequestGiveItem(atPosition AbsXyz, item slot.Slot)
}

// IShardConnection is the interface by which shards can be communicated to by
// player frontend code.
type IShardConnection interface {
	// Removes connection to shard, and removes all subscriptions to chunks in
	// the shard. Note that this does *not* send packets to tell the client to
	// unload the subscribed chunks.
	Disconnect()

	// TODO better method to send events to chunks from player frontend.
	Enqueue(fn func())

	// The following methods are requests upon chunks.

	SubscribeChunk(chunkLoc ChunkXz)

	UnsubscribeChunk(chunkLoc ChunkXz)

	MulticastPlayers(chunkLoc ChunkXz, exclude EntityId, packet []byte)

	AddPlayerData(chunkLoc ChunkXz, name string, position AbsXyz, look LookBytes, held ItemTypeId)

	RemovePlayerData(chunkLoc ChunkXz, isDisconnect bool)

	SetPlayerPositionLook(chunkLoc ChunkXz, position AbsXyz, look LookBytes, moved bool)

	// RequestHitBlock requests that the targetted block be hit.
	RequestHitBlock(held slot.Slot, target BlockXyz, digStatus DigStatus, face Face)

	// RequestHitBlock requests that the targetted block be interacted with.
	RequestInteractBlock(held slot.Slot, target BlockXyz, face Face)

	// RequestPlaceItem requests that the item passed be placed at the given
	// target location. The shard *may* choose not to do this, but if it cannot,
	// then it *must* account for the item in some way (maybe hand it back to the
	// player or just drop it on the ground).
	RequestPlaceItem(target BlockXyz, slot slot.Slot)

	// RequestTakeItem requests that the item with the specified entityId is
	// given to the player. The chunk doesn't have to respect this (particularly
	// if the item no longer exists).
	RequestTakeItem(chunkLoc ChunkXz, entityId EntityId)

	// RequestDropItem requests that an item be created.
	RequestDropItem(content slot.Slot, position AbsXyz, velocity AbsVelocity)
}

// IShardConnecter is used to look up shards and connect to them.
type IShardConnecter interface {
	// Must currently be called from with the owning IGame's Enqueue:
	ShardConnect(entityId EntityId, player IPlayerConnection, shardLoc ShardXz) IShardConnection
}

// TODO remove this interface when Enqueue* removed from IShardConnection
type IChunk interface {
	// Everything below must be called from within the containing shard's
	// goroutine.
	MulticastPlayers(exclude EntityId, packet []byte)
}

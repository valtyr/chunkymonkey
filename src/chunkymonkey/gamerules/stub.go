package gamerules

import (
	"chunkymonkey/proto"
	"chunkymonkey/types"
)

// IShardConnecter is used to look up shards and connect to them.
type IShardConnecter interface {
	// PlayerShardConnect makes a connection from a player to a shard.
	PlayerShardConnect(entityId types.EntityId, player IPlayerClient, shardLoc types.ShardXz) IPlayerShardClient

	// ShardShardConnect makes a connection from one shard to another.
	// TODO Consider making this package-private to shardserver.
	ShardShardConnect(shardLoc types.ShardXz) IShardShardClient
}

// IPlayerShardClient is the interface by which shards can be communicated to by
// player frontend code.
type IPlayerShardClient interface {
	// Removes connection to shard, and removes all subscriptions to chunks in
	// the shard. Note that this does *not* send packets to tell the client to
	// unload the subscribed chunks.
	Disconnect()

	// The following methods are requests upon chunks.

	ReqSubscribeChunk(chunkLoc types.ChunkXz, notify bool)

	ReqUnsubscribeChunk(chunkLoc types.ChunkXz)

	ReqMulticastPlayers(chunkLoc types.ChunkXz, exclude types.EntityId, packet []byte)

	ReqAddPlayerData(chunkLoc types.ChunkXz, name string, position types.AbsXyz, look types.LookBytes, held types.ItemTypeId)

	ReqRemovePlayerData(chunkLoc types.ChunkXz, isDisconnect bool)

	ReqSetPlayerPosition(chunkLoc types.ChunkXz, position types.AbsXyz)

	ReqSetPlayerLook(chunkLoc types.ChunkXz, look types.LookBytes)

	// ReqHitBlock requests that the targetted block be hit.
	ReqHitBlock(held Slot, target types.BlockXyz, digStatus types.DigStatus, face types.Face)

	// ReqHitBlock requests that the targetted block be interacted with.
	ReqInteractBlock(held Slot, target types.BlockXyz, face types.Face)

	// ReqPlaceItem requests that the item passed be placed at the given target
	// location. The shard *may* choose not to do this, but if it cannot, then it
	// *must* account for the item in some way (maybe hand it back to the player
	// or just drop it on the ground).
	ReqPlaceItem(target types.BlockXyz, slot Slot)

	// ReqTakeItem requests that the item with the specified entityId is given to
	// the player. The chunk doesn't have to respect this (particularly if the
	// item no longer exists).
	ReqTakeItem(chunkLoc types.ChunkXz, entityId types.EntityId)

	// ReqDropItem requests that an item be created.
	ReqDropItem(content Slot, position types.AbsXyz, velocity types.AbsVelocity, pickupImmunity types.Ticks)

	// ReqInventoryClick requests that the given cursor be "clicked" onto the
	// inventory. The chunk should send a replying ReqInventoryCursorUpdate to
	// reflect the new state of the cursor afterwards - in addition to any
	// ReqInventorySlotUpdate to all subscribers to the inventory.
	ReqInventoryClick(block types.BlockXyz, click Click)

	// ReqInventoryUnsubscribed requests that the inventory for the block be
	// unsubscribed to.
	ReqInventoryUnsubscribed(block types.BlockXyz)
}

// IShardShardClient provides an interface for shards to make requests against
// another shard.
// TODO Consider making this package-private to shardserver.
type IShardShardClient interface {
	Disconnect()

	ReqSetActiveBlocks(blocks []types.BlockXyz)

	ReqTransferEntity(loc types.ChunkXz, entity INonPlayerEntity)
}

// IGame provide an interface for interacting with and taking action on the
// game, including getting information about the game state, etc.
type IGame interface {
	// Broadcast a packet to all players on the server.
	BroadcastPacket(packet []byte)

	// Broadcast a message to all players on the server.
	BroadcastMessage(msg string)

	// Return a player from their name.
	PlayerByName(name string) IPlayerClient

	// Return a player from an EntityId.
	PlayerByEntityId(id types.EntityId) IPlayerClient

	// Return an ItemType from a numeric item. The boolean flag indicates
	// whether or not 'id' was a valid item type.
	ItemTypeById(id int) (ItemType, bool)
}

// IShardClient is the interface by which shards communicate to players on
// the frontend.
type IPlayerClient interface {
	GetEntityId() types.EntityId

	TransmitPacket(packet []byte)

	// NotifyChunkLoad informs Player that a chunk subscription request with
	// notify=true has completed.
	NotifyChunkLoad()

	// InventorySubscribed informs the player that an inventory has been
	// opened.
	InventorySubscribed(block types.BlockXyz, invTypeId types.InvTypeId, slots []proto.WindowSlot)

	// InventorySlotUpdate informs the player of a change to a slot in the
	// open inventory.
	InventorySlotUpdate(block types.BlockXyz, slot Slot, slotId types.SlotId)

	// InventoryProgressUpdate informs the player of a change of a progress
	// bar in a window.
	InventoryProgressUpdate(block types.BlockXyz, prgBarId types.PrgBarId, value types.PrgBarValue)

	// InventoryCursorUpdate informs the player of their new cursor contents.
	InventoryCursorUpdate(block types.BlockXyz, cursor Slot)

	// InventoryTxState requests that the player report the transaction state
	// as accepted or not. This is used by remote inventories when
	// TxStateDeferred is returned from Click.
	InventoryTxState(block types.BlockXyz, txId types.TxId, accepted bool)

	// InventorySubscribed informs the player that an inventory has been
	// closed.
	InventoryUnsubscribed(block types.BlockXyz)

	// PlaceHeldItem requests that the player frontend take one item from the
	// held item stack and send it in a ReqPlaceItem to the target block.  The
	// player code may *not* honour this request (e.g there might be no suitable
	// held item).
	PlaceHeldItem(target types.BlockXyz, wasHeld Slot)

	// OfferItem requests that the player check if it can take the item.  If
	// it can then it should ReqTakeItem from the chunk.
	OfferItem(fromChunk types.ChunkXz, entityId types.EntityId, item Slot)

	// GiveItemAtPosition requests that the player takes the item contents
	// into their inventory. If they cannot, then the player should drop the
	// item at the given position.
	GiveItemAtPosition(atPosition types.AbsXyz, item Slot)

	// GiveItem is a wrapper for GiveItemAtPosition that uses the player's
	// current position as the 'atPosition'.
	GiveItem(item Slot)

	// PositionLook returns the player's current position and look
	PositionLook() (types.AbsXyz, types.LookDegrees)

	// SetPositionLook changes the player's position and look
	SetPositionLook(types.AbsXyz, types.LookDegrees)

	// EchoMessage displays a message to the player
	EchoMessage(msg string)
}

type ICommandFramework interface {
	Prefix() string
	Process(player IPlayerClient, cmd string, game IGame)
}

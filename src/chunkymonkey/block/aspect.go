package block

import (
	"rand"

	"chunkymonkey/itemtype"
	"chunkymonkey/recipe"
	"chunkymonkey/slot"
	"chunkymonkey/stub"
	. "chunkymonkey/types"
)

// The distance from the edge of a block that items spawn at in fractional
// blocks.
const blockItemSpawnFromEdge = 4.0 / PixelsPerBlock

// The interface required of a chunk by block behaviour.
type IChunkBlock interface {
	Rand() *rand.Rand
	ItemType(itemTypeId ItemTypeId) (itemType *itemtype.ItemType, ok bool)
	AddSpawn(s stub.INonPlayerSpawn)
	BlockExtra(subLoc *SubChunkXyz) interface{}
	SetBlockExtra(subLoc *SubChunkXyz, extra interface{})
	RecipeSet() *recipe.RecipeSet
	AddOnUnsubscribe(entityId EntityId, observer IUnsubscribed)
	RemoveOnUnsubscribe(entityId EntityId, observer IUnsubscribed)
}

// IUnsubscribed is the interface by which blocks (and potentially other
// things) can register themselves to be called when a player unsubscribes from
// a chunk.
type IUnsubscribed interface {
	Unsubscribed(entityId EntityId)
}

// BlockInstance represents the instance of a block within a chunk. It is used
// to pass context to a IBlockAspect method call. BlockInstances must not be
// modified after creation.
type BlockInstance struct {
	Chunk    IChunkBlock
	BlockLoc BlockXyz
	SubLoc   SubChunkXyz
	Index    BlockIndex
	// TODO decide if *BlockType belongs in here as well.
	// Note that only the lower nibble of data is stored.
	Data byte
}

// Defines the behaviour of a block.
type IBlockAspect interface {
	Name() string
	Hit(instance *BlockInstance, player stub.IPlayerConnection, digStatus DigStatus) (destroyed bool)
	Interact(instance *BlockInstance, player stub.IPlayerConnection)
	InventoryClick(instance *BlockInstance, player stub.IPlayerConnection, slotId SlotId, cursor *slot.Slot, rightClick bool, shiftClick bool, txId TxId, expectedSlot *slot.Slot)
	InventoryUnsubscribed(instance *BlockInstance, player stub.IPlayerConnection)
	Destroy(instance *BlockInstance)
}

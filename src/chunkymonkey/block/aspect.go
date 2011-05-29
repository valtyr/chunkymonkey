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
	GetRand() *rand.Rand
	GetItemType(itemTypeId ItemTypeId) (itemType *itemtype.ItemType, ok bool)
	AddSpawn(s stub.INonPlayerSpawn)
	GetBlockExtra(subLoc *SubChunkXyz) interface{}
	SetBlockExtra(subLoc *SubChunkXyz, extra interface{})
	GetRecipeSet() *recipe.RecipeSet

	// The above methods are freely callable in the goroutine context of a call
	// to a IBlockAspect method (as the chunk itself calls that). But from any
	// other goroutine they must be called via EnqueueGeneric().
	EnqueueGeneric(f func())
}

// BlockInstance represents the instance of a block within a chunk.
type BlockInstance struct {
	Chunk    IChunkBlock
	BlockLoc BlockXyz
	SubLoc   SubChunkXyz
	// TODO decide if *BlockType belongs in here as well.
	// Note that only the lower nibble of data is stored.
	Data byte
}

// Defines the behaviour of a block.
type IBlockAspect interface {
	Name() string
	Hit(instance *BlockInstance, player stub.IPlayerConnection, digStatus DigStatus) (destroyed bool)
	Interact(instance *BlockInstance, player stub.IPlayerConnection)
	Click(instance *BlockInstance, player stub.IPlayerConnection, cursor *slot.Slot, rightClick bool, shiftClick bool, slotId SlotId)
}

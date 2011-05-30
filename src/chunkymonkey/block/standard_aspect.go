package block

import (
	"chunkymonkey/slot"
	"chunkymonkey/stub"
	. "chunkymonkey/types"
)

func makeStandardAspect() (aspect IBlockAspect) {
	return &StandardAspect{}
}

// Behaviour of a "standard" block. A StandardAspect block is one that is
// diggable, and drops items in a simple manner. StandardAspect blocks do not
// use block metadata.
type StandardAspect struct {
	// Items, up to one of which will potentially spawn when block destroyed.
	DroppedItems []blockDropItem
	BreakOn      DigStatus
}

func (aspect *StandardAspect) Name() string {
	return "Standard"
}

func (aspect *StandardAspect) Hit(instance *BlockInstance, player stub.IPlayerConnection, digStatus DigStatus) (destroyed bool) {
	if aspect.BreakOn != digStatus {
		return
	}

	destroyed = true

	if len(aspect.DroppedItems) > 0 {
		rand := instance.Chunk.Rand()
		// Possibly drop item(s)
		r := byte(rand.Intn(100))
		for _, dropItem := range aspect.DroppedItems {
			if dropItem.Probability > r {
				dropItem.drop(instance)
				break
			}
			r -= dropItem.Probability
		}
	}

	return
}

func (aspect *StandardAspect) Interact(instance *BlockInstance, player stub.IPlayerConnection) {
}

func (aspect *StandardAspect) Click(instance *BlockInstance, player stub.IPlayerConnection, cursor *slot.Slot, rightClick bool, shiftClick bool, slotId SlotId) {
}

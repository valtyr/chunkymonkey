package block

import (
	"chunkymonkey/slot"
	"chunkymonkey/stub"
	. "chunkymonkey/types"
)

func makeVoidAspect() (aspect IBlockAspect) {
	return &VoidAspect{}
}

// Behaviour of a "void" block which has no behaviour.
type VoidAspect struct{}

func (aspect *VoidAspect) Name() string {
	return "Void"
}

func (aspect *VoidAspect) Hit(instance *BlockInstance, player stub.IPlayerConnection, digStatus DigStatus) (destroyed bool) {
	destroyed = false
	return
}

func (aspect *VoidAspect) Interact(instance *BlockInstance, player stub.IPlayerConnection) {
}

func (aspect *VoidAspect) Click(instance *BlockInstance, player stub.IPlayerConnection, cursor *slot.Slot, rightClick bool, shiftClick bool, slotId SlotId) {
}

func (aspect *VoidAspect) Destroy(instance *BlockInstance) {
}

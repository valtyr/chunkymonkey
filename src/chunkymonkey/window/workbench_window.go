package window

import (
	"chunkymonkey/inventory"
	"chunkymonkey/slot"
	. "chunkymonkey/types"
)

const (
	workbenchInvCraftNum   = 1 + inventory.WorkbenchInvCraftWidth*inventory.WorkbenchInvCraftHeight
	workbenchInvCraftStart = 0
	workbenchInvCraftEnd   = workbenchInvCraftStart + workbenchInvCraftNum

	workbenchInvMainStart = workbenchInvCraftEnd
	workbenchInvMainEnd   = workbenchInvMainStart + playerInvMainNum

	workbenchInvHoldingStart = workbenchInvMainEnd
	workbenchInvHoldingEnd   = workbenchInvHoldingStart + playerInvHoldingNum
)

type WorkbenchWindow struct {
	Window
	crafting IInventory
	main     *inventory.Inventory
	holding  *inventory.Inventory
}

func NewWorkbenchWindow(entityId EntityId, viewer IWindowViewer, windowId WindowId, crafting IInventory, main, holding *inventory.Inventory) (w *WorkbenchWindow) {
	w = &WorkbenchWindow{
		crafting: crafting,
		main:     main,
		holding:  holding,
	}
	w.Window.Init(
		windowId,
		InvTypeIdWorkbench,
		viewer,
		"Crafting",
		w.crafting,
		main,
		holding)

	return
}

func (w *WorkbenchWindow) Click(slotId SlotId, cursor *slot.Slot, rightClick bool, shiftClick bool, txId TxId, expectedSlot *slot.Slot) TxState {
	switch {
	case slotId < 0:
		break
	case slotId < workbenchInvCraftEnd:
		return w.crafting.Click(
			slotId-workbenchInvCraftStart,
			cursor, rightClick, shiftClick, txId, expectedSlot)
	case slotId < workbenchInvMainEnd:
		return w.main.Click(
			slotId-workbenchInvMainStart,
			cursor, rightClick, shiftClick, txId, expectedSlot)
	case slotId < workbenchInvHoldingEnd:
		return w.holding.Click(
			slotId-workbenchInvHoldingStart,
			cursor, rightClick, shiftClick, txId, expectedSlot)
	}
	return TxStateRejected
}

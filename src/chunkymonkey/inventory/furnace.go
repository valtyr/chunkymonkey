package inventory

import (
	"chunkymonkey/recipe"
	"chunkymonkey/slot"
	. "chunkymonkey/types"
)

const (
	furnaceSlotReagent = SlotId(0)
	furnaceSlotFuel    = SlotId(1)
	furnaceSlotOutput  = SlotId(2)
	furnaceNumSlots    = 3
)

type FurnaceInventory struct {
	Inventory
	furnaceData *recipe.FurnaceData
}

// NewFurnaceInventory creates a furnace inventory.
func NewFurnaceInventory(furnaceData *recipe.FurnaceData) (inv *FurnaceInventory) {
	inv = &FurnaceInventory{
		furnaceData: furnaceData,
	}
	inv.Inventory.Init(furnaceNumSlots)
	return
}

func (inv *FurnaceInventory) Click(slotId SlotId, cursor *slot.Slot, rightClick bool, shiftClick bool, txId TxId, expectedSlot *slot.Slot) (txState TxState) {

	cursorItemId := cursor.GetItemTypeId()

	switch (slotId) {
	case furnaceSlotReagent:
		_, cursorIsReagent := inv.furnaceData.Reactions[cursorItemId]
		if cursorIsReagent || cursor.IsEmpty() {
			txState = inv.Inventory.Click(
				slotId, cursor, rightClick, shiftClick, txId, expectedSlot)
			// TODO If the reagent changes, the reaction should restart.
		} else {
			return TxStateRejected
		}
	case furnaceSlotFuel:
		_, cursorIsFuel := inv.furnaceData.Fuels[cursorItemId]
		if cursorIsFuel || cursor.IsEmpty() {
			txState = inv.Inventory.Click(
				slotId, cursor, rightClick, shiftClick, txId, expectedSlot)
		}
		// TODO If fuel has been added to an empty slot, set the furnace burning.
	case furnaceSlotOutput:
		// Player may only *take* the *whole* stack from the output slot.
		txState = inv.Inventory.TakeOnlyClick(
			slotId, cursor, rightClick, shiftClick, txId, expectedSlot)
	}

	return
}

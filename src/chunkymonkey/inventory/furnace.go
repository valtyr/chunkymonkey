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
	active      bool
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

	switch slotId {
	case furnaceSlotReagent:
		txState = inv.Inventory.Click(
			slotId, cursor, rightClick, shiftClick, txId, expectedSlot)
		// TODO If the reagent type changes, the reaction should restart.
	case furnaceSlotFuel:
		cursorItemId := cursor.GetItemTypeId()
		_, cursorIsFuel := inv.furnaceData.Fuels[cursorItemId]
		if cursorIsFuel || cursor.IsEmpty() {
			txState = inv.Inventory.Click(
				slotId, cursor, rightClick, shiftClick, txId, expectedSlot)
		}
	case furnaceSlotOutput:
		// Player may only *take* the *whole* stack from the output slot.
		txState = inv.Inventory.TakeOnlyClick(
			slotId, cursor, rightClick, shiftClick, txId, expectedSlot)
	}

	// If the fuel and reagent slots are non-empty, make the furnace active.
	if !inv.slots[furnaceSlotFuel].IsEmpty() && !inv.slots[furnaceSlotReagent].IsEmpty() {
		inv.active = true
	}

	return
}

func (inv *FurnaceInventory) IsActive() bool {
	return inv.active
}

// Tick runs the furnace for a single tick.
func (inv *FurnaceInventory) Tick() {
	return
}

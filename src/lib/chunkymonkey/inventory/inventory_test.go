package inventory

import (
    "testing"

    .   "chunkymonkey/types"
)

func TestInventory_Init(t *testing.T) {
    var inv Inventory
    inv.Init(10, make([]SlotId, 0))

    for i, slot := range inv.slots {
        if slot.ItemTypeId != ItemTypeIdNull || slot.Count != 0 || slot.Data != 0 {
            t.Errorf("Slot %d not initialized: %+v", i, slot)
        }
    }
}

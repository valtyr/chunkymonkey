package gamerules

import (
	"testing"

	. "chunkymonkey/types"
)

func TestInventory_Init(t *testing.T) {
	var inv Inventory
	inv.Init(10)

	for i, slot := range inv.slots {
		if slot.ItemTypeId != ItemTypeIdNull || slot.Count != 0 || slot.Data != 0 {
			t.Errorf("Slot %d not initialized: %+v", i, slot)
		}
	}
}

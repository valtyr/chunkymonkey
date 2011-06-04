package block

import (
	"chunkymonkey/inventory"
	. "chunkymonkey/types"
)

func makeFurnaceAspect() IBlockAspect {
	return &FurnaceAspect{
		InventoryAspect{
			name:                 "Furnace",
			createBlockInventory: createFurnaceInventory,
		},
	}
}

type FurnaceAspect struct {
	InventoryAspect
}

func createFurnaceInventory(instance *BlockInstance) *blockInventory {
	return newBlockInventory(
		instance,
		inventory.NewFurnaceInventory(),
		false,
		InvTypeIdFurnace,
	)
}

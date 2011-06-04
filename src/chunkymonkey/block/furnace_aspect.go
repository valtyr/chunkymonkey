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
		inventory.NewFurnaceInventory(instance.Chunk.FurnaceData()),
		false,
		InvTypeIdFurnace,
	)
}

func (aspect *FurnaceAspect) Tick(instance *BlockInstance) bool {
	// TODO
	return false
}

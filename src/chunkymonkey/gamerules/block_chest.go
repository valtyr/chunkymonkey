package gamerules

import (
	. "chunkymonkey/types"
)

func makeChestAspect() (aspect IBlockAspect) {
	return &InventoryAspect{
		name:                 "Chest",
		createBlockInventory: createChestInventory,
	}
}

// Creates a new tile entity for a chest. ReadNbt and SetChunk must be called
// before any other methods.
func NewChestTileEntity() ITileEntity {
	return createChestInventory(nil)
}

func createChestInventory(instance *BlockInstance) *blockInventory {
	inv := new(Inventory)
	inv.InitChestInventory()
	return newBlockInventory(
		instance,
		inv,
		false,
		InvTypeIdChest,
	)
}

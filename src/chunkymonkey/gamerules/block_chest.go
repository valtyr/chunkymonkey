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

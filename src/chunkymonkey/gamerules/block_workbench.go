package gamerules

import (
	. "chunkymonkey/types"
)

func makeWorkbenchAspect() (aspect IBlockAspect) {
	return &InventoryAspect{
		name:                 "Workbench",
		createBlockInventory: createWorkbenchInventory,
	}
}

func createWorkbenchInventory(instance *BlockInstance) *blockInventory {
	inv := NewWorkbenchInventory(instance.Chunk.RecipeSet())
	return newBlockInventory(
		instance,
		inv,
		true,
		InvTypeIdWorkbench,
	)
}

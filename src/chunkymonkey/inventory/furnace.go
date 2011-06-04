package inventory

import (
	"chunkymonkey/recipe"
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
	inv.Inventory.Init(3)
	return
}

// TODO special behaviour

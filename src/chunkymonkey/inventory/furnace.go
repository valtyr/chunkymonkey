package inventory

type FurnaceInventory struct {
	Inventory
	// TODO furnace recipe data
}

// NewFurnaceInventory creates a furnace inventory.
func NewFurnaceInventory() (inv *FurnaceInventory) {
	inv = &FurnaceInventory{}
	inv.Inventory.Init(3)
	return
}

// TODO special behaviour

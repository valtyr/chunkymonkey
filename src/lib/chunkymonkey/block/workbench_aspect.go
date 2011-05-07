package block

import (
	"chunkymonkey/inventory"
	. "chunkymonkey/types"
)

func makeWorkbenchAspect() (aspect IBlockAspect) {
	return &WorkbenchAspect{}
}

// WorkbenchAspect is the behaviour for the workbench block that allows 3x3
// crafting.
type WorkbenchAspect struct {
	StandardAspect
}

func (aspect *WorkbenchAspect) Name() string {
	return "Workbench"
}

func (aspect *WorkbenchAspect) Interact(instance *BlockInstance, player IBlockPlayer) {
	workbenchInv, ok := instance.Chunk.GetBlockExtra(&instance.SubLoc).(*inventory.WorkbenchInventory)
	if !ok {
		// TODO have the inventory stop existing when all players unsubscribe from
		// it, and have it throw items out into the chunk it's in.
		workbenchInv = inventory.NewWorkbenchInventory(instance.Chunk.GetRecipeSet())
		instance.Chunk.SetBlockExtra(&instance.SubLoc, workbenchInv)
		return
	}
	player.OpenWindow(InvTypeIdWorkbench, workbenchInv)
}

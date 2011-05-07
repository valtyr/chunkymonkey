package inventory

import (
	"chunkymonkey/recipe"
	"chunkymonkey/slot"
	. "chunkymonkey/types"
)

const (
	workbenchInvCraftWidth  = 3
	workbenchInvCraftHeight = 3
	workbenchInvCraftNum    = 1 + workbenchInvCraftWidth*workbenchInvCraftHeight
	workbenchInvCraftStart  = 0
	workbenchInvCraftEnd    = workbenchInvCraftStart + workbenchInvCraftNum

	workbenchInvMainStart = workbenchInvCraftEnd
	workbenchInvMainEnd   = workbenchInvMainStart + playerInvMainNum

	workbenchInvHoldingStart = workbenchInvMainEnd
	workbenchInvHoldingEnd   = workbenchInvHoldingStart + playerInvHoldingNum
)

type WorkbenchWindow struct {
	Window
	recipes   *recipe.RecipeSet
	crafting  CraftingInventory
	main      *Inventory
	holding   *Inventory
}

func NewWorkbenchWindow(entityId EntityId, viewer IWindowViewer, recipes *recipe.RecipeSet, windowId WindowId, main, holding *Inventory) (w *WorkbenchWindow) {
	w = &WorkbenchWindow{
		recipes: recipes,
		main:    main,
		holding: holding,
	}
	w.crafting.Init(workbenchInvCraftWidth, workbenchInvCraftHeight, recipes)
	w.Window.Init(
		windowId,
		InvTypeIdWorkbench,
		viewer,
		"Crafting",
		&w.crafting.Inventory,
		main,
		holding)

	return
}

func (w *WorkbenchWindow) Click(slotId SlotId, cursor *slot.Slot, rightClick bool, shiftClick bool) (accepted bool) {
	switch {
	case slotId < 0:
		return false
	case slotId < workbenchInvCraftEnd:
		accepted = w.crafting.Click(
			slotId-workbenchInvCraftStart,
			cursor, rightClick, shiftClick)
	case slotId < workbenchInvMainEnd:
		accepted = w.main.StandardClick(
			slotId-workbenchInvMainStart,
			cursor, rightClick, shiftClick)
	case slotId < workbenchInvHoldingEnd:
		accepted = w.holding.StandardClick(
			slotId-workbenchInvHoldingStart,
			cursor, rightClick, shiftClick)
	}
	return
}

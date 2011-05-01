package inventory

import (
    "io"
    "os"

    "chunkymonkey/slot"
    . "chunkymonkey/types"
)

const (
    playerInvCraftStart = SlotId(0)
    playerInvCraftEnd   = SlotId(5)
    playerInvCraftNum   = int(playerInvCraftEnd - playerInvCraftStart)

    playerInvArmorStart = SlotId(5)
    playerInvArmorEnd   = SlotId(9)
    playerInvArmorNum   = int(playerInvArmorEnd - playerInvArmorStart)

    playerInvMainStart = SlotId(9)
    playerInvMainEnd   = SlotId(36)
    playerInvMainNum   = int(playerInvMainEnd - playerInvMainStart)

    playerInvHoldingStart = SlotId(36)
    playerInvHoldingEnd   = SlotId(45)
    playerInvHoldingNum   = int(playerInvHoldingEnd - playerInvHoldingStart)

    playerInvSize = playerInvCraftNum + playerInvArmorNum + playerInvMainNum + playerInvHoldingNum
)

type PlayerInventory struct {
    Window
    entityId     EntityId
    crafting     Inventory
    armor        Inventory
    main         Inventory
    holding      Inventory
    holdingIndex SlotId
}

// Init initializes PlayerInventory.
// entityId - The EntityId of the player who holds the inventory.
func (inv *PlayerInventory) Init(entityId EntityId, viewer IWindowViewer) {
    inv.crafting.Init(playerInvCraftNum)
    inv.armor.Init(playerInvArmorNum)
    inv.main.Init(playerInvMainNum)
    inv.holding.Init(playerInvHoldingNum)
    inv.Window.Init(
        WindowIdInventory,
        // Note that we have no known value for invTypeId - but it's only used
        // in WriteWindowOpen which isn't used for PlayerInventory.
        -1,
        viewer,
        "Inventory",
        &inv.crafting,
        &inv.armor,
        &inv.main,
        &inv.holding,
    )
    inv.holdingIndex = 0
    inv.entityId = entityId
}

// SetHolding chooses the held item (0-8). Out of range values have no effect.
func (inv *PlayerInventory) SetHolding(holding SlotId) {
    if holding >= 0 && holding < SlotId(playerInvHoldingNum) {
        inv.holdingIndex = holding
    }
}

// HeldItem returns the slot that is the current "held" item.
// TODO need any changes to the held item slot to create notifications to
// players.
func (inv *PlayerInventory) HeldItem() (slot *slot.Slot, slotId SlotId) {
    slotId = playerInvHoldingStart + inv.holdingIndex
    slot = &inv.holding.slots[inv.holdingIndex]
    return
}

// Writes packets for other players to see the equipped items.
func (inv *PlayerInventory) SendFullEquipmentUpdate(writer io.Writer) (err os.Error) {
    slot, _ := inv.HeldItem()
    err = slot.SendEquipmentUpdate(writer, inv.entityId, 0)
    if err != nil {
        return
    }

    for i := range inv.armor.slots {
        err = inv.armor.slots[i].SendEquipmentUpdate(writer, inv.entityId, SlotId(i+1))
        if err != nil {
            return
        }
    }
    return
}

// PutItem attempts to put the item stack into the player's inventory. The item
// will be modified as a result.
func (inv *PlayerInventory) PutItem(item *slot.Slot) {
    inv.holding.PutItem(item)
    inv.main.PutItem(item)
    return
}

func (inv *PlayerInventory) Click(slotId SlotId, cursor *slot.Slot, rightClick bool, shiftClick bool) (accepted bool) {
    switch {
    case slotId < 0:
        return false
    case slotId < playerInvCraftEnd:
        // TODO - handle crafting
        return false
    case slotId < playerInvArmorEnd:
        // TODO - handle armor
        return false
    case slotId < playerInvMainEnd:
        accepted = inv.main.StandardClick(
            slotId-playerInvMainStart,
            cursor, rightClick, shiftClick)
    case slotId < playerInvHoldingEnd:
        accepted = inv.holding.StandardClick(
            slotId-playerInvHoldingStart,
            cursor, rightClick, shiftClick)
    }
    return
}

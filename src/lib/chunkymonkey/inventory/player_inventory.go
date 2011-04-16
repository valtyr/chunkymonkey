package inventory

import (
    "io"
    "os"

    .   "chunkymonkey/types"
)

const (
    playerInvSize = 45

    playerInvArmorStart = SlotId(5)
    playerInvArmorEnd   = SlotId(8)

    playerInvMainStart = SlotId(9)
    playerInvMainEnd   = SlotId(35)
    playerInvMainNum   = 1 + playerInvMainEnd - playerInvMainStart

    playerInvHeldStart = SlotId(36)
    playerInvHeldEnd   = SlotId(44)
    playerInvHeldNum   = 1 + playerInvHeldEnd - playerInvHeldStart
)

// Determines the order in which items automatically "land" in the player
// inventory.
var playerInvOrder []SlotId

type PlayerInventory struct {
    Inventory
    holding SlotId // Note that this is 0-8.
}

func (inv *PlayerInventory) Init() {
    inv.Inventory.Init(playerInvSize, playerInvOrder)
    inv.holding = 0
}

// Writes packets for other players to see the equipped items.
func (inv *PlayerInventory) SendFullEquipmentUpdate(writer io.Writer, entityId EntityId) (err os.Error) {
    err = inv.HeldItem().SendEquipmentUpdate(writer, entityId, 0)
    if err != nil {
        return
    }

    equipSlot := SlotId(1)
    for i := playerInvArmorStart; i <= playerInvArmorEnd; i++ {
        err = inv.Slot(i).SendEquipmentUpdate(writer, entityId, equipSlot)
        if err != nil {
            return
        }
        equipSlot++
    }
    return
}

// Chooses the held item (0-8). Out of range values have no effect.
func (inv *PlayerInventory) SetHolding(holding SlotId) {
    if holding >= 0 && holding < playerInvHeldNum {
        inv.holding = holding
    }
}

// Returns the slot that is the current "held" item.
func (inv *PlayerInventory) HeldItem() *Slot {
    return inv.Slot(inv.holding + playerInvHeldStart)
}

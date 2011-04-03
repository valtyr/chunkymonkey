package inventory

import (
    "io"
    "os"

    .   "chunkymonkey/types"
)

const (
    playerInvSize = 45

    playerInvHeldStart = SlotID(36)
    playerInvHeldEnd   = SlotID(44)
    playerInvHeldNum   = 1 + playerInvHeldEnd - playerInvHeldStart

    playerInvArmorStart = SlotID(5)
    playerInvArmorEnd   = SlotID(8)
)

type PlayerInventory struct {
    Inventory
    holding SlotID // Note that this is 0-8.
}

func (inv *PlayerInventory) Init() {
    inv.Inventory.Init(playerInvSize)
    inv.holding = 0
}

// Writes packets for other players to see the equipped items.
func (inv *PlayerInventory) SendFullEquipmentUpdate(writer io.Writer, entityID EntityID) (err os.Error) {
    err = inv.HeldItem().SendEquipmentUpdate(writer, entityID, 0)
    if err != nil {
        return
    }

    equipSlot := SlotID(1)
    for i := playerInvArmorStart; i <= playerInvArmorEnd; i++ {
        err = inv.Slot(i).SendEquipmentUpdate(writer, entityID, equipSlot)
        if err != nil {
            return
        }
        equipSlot++
    }
    return
}

// Chooses the held item (0-8). Out of range values have no effect.
func (inv *PlayerInventory) SetHolding(holding SlotID) {
    if holding >= 0 && holding < playerInvHeldNum {
        inv.holding = holding
    }
}

// Returns the slot that is the current "held" item.
func (inv *PlayerInventory) HeldItem() *Slot {
    return inv.Slot(inv.holding + playerInvHeldStart)
}

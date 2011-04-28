package inventory

import (
    "io"
    "os"

    "chunkymonkey/proto"
    "chunkymonkey/slot"
    . "chunkymonkey/types"
)

type Inventory struct {
    slots      []slot.Slot
    slotsProto []proto.IWindowSlot // Holds same items as `slots`.
}

func (inv *Inventory) Init(size int) {
    inv.slots = make([]slot.Slot, size)
    inv.slotsProto = make([]proto.IWindowSlot, size)
    for i := range inv.slots {
        inv.slots[i].Init()
        inv.slotsProto[i] = &inv.slots[i]
    }
}

func (inv *Inventory) Slot(slotId SlotId) *slot.Slot {
    if slotId < 0 || int(slotId) > len(inv.slots) {
        return nil
    }
    return &inv.slots[slotId]
}

func (inv *Inventory) SendUpdate(writer io.Writer, windowId WindowId) os.Error {
    return proto.WriteWindowItems(writer, windowId, inv.slotsProto)
}

func init() {
    // Set up playerInvOrder
    numPlayerAutoSlots := playerInvHeldNum + playerInvMainNum
    playerInvOrder = make([]SlotId, 0, numPlayerAutoSlots)
    for slotId := playerInvHeldStart; slotId <= playerInvHeldEnd; slotId++ {
        playerInvOrder = append(playerInvOrder, slotId)
    }
    for slotId := playerInvMainStart; slotId <= playerInvMainEnd; slotId++ {
        playerInvOrder = append(playerInvOrder, slotId)
    }
}

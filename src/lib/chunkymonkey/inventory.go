package inventory

import (
    "io"
    "os"

    "chunkymonkey/proto"
    .   "chunkymonkey/types"
)

type Inventory struct {
    slots      []Slot
    slotsProto []proto.IWindowSlot // Holds same items as `slots`.
}

func (inv *Inventory) Init(size int) {
    inv.slots = make([]Slot, size)
    inv.slotsProto = make([]proto.IWindowSlot, size)
    for i := range inv.slots {
        inv.slots[i].Init()
        inv.slotsProto[i] = &inv.slots[i]
    }
}

func (inv *Inventory) Slot(slotID SlotID) *Slot {
    if slotID < 0 || int(slotID) > len(inv.slots) {
        return nil
    }
    return &inv.slots[slotID]
}

func (inv *Inventory) SendUpdate(writer io.Writer, windowID WindowID) os.Error {
    return proto.WriteWindowItems(writer, windowID, inv.slotsProto)
}

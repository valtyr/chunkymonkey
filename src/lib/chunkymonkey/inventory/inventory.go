package inventory

import (
    "io"
    "os"

    .   "chunkymonkey/interfaces"
    "chunkymonkey/proto"
    "chunkymonkey/slot"
    .   "chunkymonkey/types"
)

type Inventory struct {
    slots      []slot.Slot
    slotsProto []proto.IWindowSlot // Holds same items as `slots`.
    slotOrder  []SlotId            // The order to try to drop items into slots.
}

func (inv *Inventory) Init(size int, slotOrder []SlotId) {
    inv.slots = make([]slot.Slot, size)
    inv.slotsProto = make([]proto.IWindowSlot, size)
    for i := range inv.slots {
        inv.slots[i].Init()
        inv.slotsProto[i] = &inv.slots[i]
    }
    inv.slotOrder = slotOrder
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

// Returns taken=true if any count was removed from the item. For each slot
// that changes, the slotChanged function is called.
func (inv *Inventory) PutItem(item IItem, slotChanged func(slotId SlotId, slot *slot.Slot)) (taken bool) {
    // TODO optimize this algorithm, maybe by maintaining a map of non-full
    // slots containing an item of various item type IDs.
    srcSlot := item.GetSlot()
    for _, slotIndex := range inv.slotOrder {
        slot := &inv.slots[slotIndex]
        if srcSlot.Count <= 0 {
            break
        }
        if slot.ItemType == ItemIdNull || slot.ItemType == srcSlot.ItemType {
            if slot.Add(srcSlot) {
                taken = true
                if slotChanged != nil {
                    slotChanged(SlotId(slotIndex), slot)
                }
            }
        }
    }
    return
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

package inventory

import (
    "sync"

    "chunkymonkey/proto"
    "chunkymonkey/slot"
    . "chunkymonkey/types"
)

type IInventorySubscriber interface {
    SlotUpdate(slot *slot.Slot, slotId SlotId)
}

type Inventory struct {
    lock        sync.Mutex
    slots       []slot.Slot
    slotsProto  []proto.IWindowSlot // Holds same items as `slots`.
    subscribers map[IInventorySubscriber]bool
}

func (inv *Inventory) Init(size int) {
    inv.slots = make([]slot.Slot, size)
    inv.slotsProto = make([]proto.IWindowSlot, size)
    for i := range inv.slots {
        inv.slots[i].Init()
        inv.slotsProto[i] = &inv.slots[i]
    }
    inv.subscribers = make(map[IInventorySubscriber]bool)
}

func (inv *Inventory) AddSubscriber(subscriber IInventorySubscriber) {
    inv.lock.Lock()
    defer inv.lock.Unlock()
    inv.subscribers[subscriber] = true
}

func (inv *Inventory) RemoveSubscriber(subscriber IInventorySubscriber) {
    inv.lock.Lock()
    defer inv.lock.Unlock()
    inv.subscribers[subscriber] = false, false
}

// StandardClick takes the default actions upon a click event from a player.
func (inv *Inventory) StandardClick(slotId SlotId, cursor *slot.Slot, rightClick bool, shiftClick bool) (accepted bool) {
    inv.lock.Lock()
    defer inv.lock.Unlock()

    if slotId < 0 || int(slotId) > len(inv.slots) {
        return false
    }

    clickedSlot := &inv.slots[slotId]

    // Apply the change.
    if cursor.Count == 0 {
        if rightClick {
            accepted = clickedSlot.Split(cursor)
        } else {
            accepted = clickedSlot.Swap(cursor)
        }
    } else {
        if rightClick {
            accepted = clickedSlot.AddOne(cursor)
        } else {
            accepted = clickedSlot.Add(cursor)
        }

        if !accepted {
            accepted = clickedSlot.Swap(cursor)
        }
    }

    // We send slot updates in case we have custom max counts that differ
    // from the client's idea.
    inv.slotUpdate(clickedSlot, slotId)

    return
}

// PutItem attempts to put the given item into the inventory.
func (inv *Inventory) PutItem(item *slot.Slot) {
    inv.lock.Lock()
    defer inv.lock.Unlock()
    // TODO optimize this algorithm, maybe by maintaining a map of non-full
    // slots containing an item of various item type IDs. Additionally, it
    // should prefer to put stackable items into stacks of the same type,
    // rather than in empty slots.
    for slotIndex := range inv.slots {
        if item.Count <= 0 {
            break
        }
        slot := &inv.slots[slotIndex]
        if slot.ItemType == nil || slot.ItemType == item.ItemType {
            if slot.Add(item) {
                inv.slotUpdate(slot, SlotId(slotIndex))
            }
        }
    }
    return
}

// Send message about the slot change to the relevant places.
func (inv *Inventory) slotUpdate(slot *slot.Slot, slotId SlotId) {
    for subscriber := range inv.subscribers {
        subscriber.SlotUpdate(slot, slotId)
    }
}

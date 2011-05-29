package inventory

import (
	"chunkymonkey/proto"
	"chunkymonkey/slot"
	. "chunkymonkey/types"
)

type IInventorySubscriber interface {
	// SlotUpdate is called when a slot inside the inventory changes its
	// contents.
	SlotUpdate(slot *slot.Slot, slotId SlotId)

	// Unsubscribed is called when the inventory is cutting the subscription.
	// This will typically be when the inventory is destroyed.
	Unsubscribed()
}

type Inventory struct {
	slots      []slot.Slot
	subscriber IInventorySubscriber
}

// Init initializes the inventory. onUnsubscribed is called in a new goroutine
// when the number of subscribers to the inventory reaches zero (but is not
// called initially).
func (inv *Inventory) Init(size int) {
	inv.slots = make([]slot.Slot, size)
	for i := range inv.slots {
		inv.slots[i].Init()
	}
}

func (inv *Inventory) Destroy() {
	// TODO call this method from the appropriate place(s).
	if inv.subscriber != nil {
		inv.subscriber.Unsubscribed()
	}
}

func (inv *Inventory) NumSlots() SlotId {
	return SlotId(len(inv.slots))
}

func (inv *Inventory) SetSubscriber(subscriber IInventorySubscriber) {
	inv.subscriber = subscriber
}

// StandardClick takes the default actions upon a click event from a player.
func (inv *Inventory) StandardClick(slotId SlotId, cursor *slot.Slot, rightClick bool, shiftClick bool) (accepted bool) {
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

	// We send slot updates in case we have custom max counts that differ from
	// the client's idea.
	inv.slotUpdate(clickedSlot, slotId)

	return
}

// TakeOnlyClick only allows items to be taken from the slot, and it only
// allows the *whole* stack to be taken, otherwise no items are taken at all.
func (inv *Inventory) TakeOnlyClick(slotId SlotId, cursor *slot.Slot, rightClick, shiftClick bool) (accepted bool) {
	if slotId < 0 || int(slotId) > len(inv.slots) {
		return false
	}

	clickedSlot := &inv.slots[slotId]

	// Apply the change.
	accepted = cursor.AddWhole(clickedSlot)

	// We send slot updates in case we have custom max counts that differ from
	// the client's idea.
	inv.slotUpdate(clickedSlot, slotId)

	return
}

func (inv *Inventory) Slot(slotId SlotId) slot.Slot {
	return inv.slots[slotId]
}

func (inv *Inventory) TakeOneItem(slotId SlotId, into *slot.Slot) {
	slot := &inv.slots[slotId]
	if into.AddOne(slot) {
		inv.slotUpdate(slot, slotId)
	}
}

// PutItem attempts to put the given item into the inventory.
func (inv *Inventory) PutItem(item *slot.Slot) {
	// TODO optimize this algorithm, maybe by maintaining a map of non-full
	// slots containing an item of various item type IDs. Additionally, it
	// should prefer to put stackable items into stacks of the same type,
	// rather than in empty slots.
	for slotIndex := range inv.slots {
		if item.Count <= 0 {
			break
		}
		slot := &inv.slots[slotIndex]
		if slot.Add(item) {
			inv.slotUpdate(slot, SlotId(slotIndex))
		}
	}
}

// CanTakeItem returns true if it can take at least one item from the passed
// Slot.
func (inv *Inventory) CanTakeItem(item *slot.Slot) bool {
	if item.Count <= 0 {
		return false
	}

	itemCopy := *item

	for slotIndex := range inv.slots {
		slotCopy := inv.slots[slotIndex]

		if slotCopy.Add(&itemCopy) {
			return true
		}
	}

	return false
}

func (inv *Inventory) MakeProtoSlots() []proto.WindowSlot {
	slots := make([]proto.WindowSlot, len(inv.slots))
	inv.WriteProtoSlots(slots)
	return slots
}

// WriteProtoSlots stores into the slots parameter the proto version of the
// item data in the inventory.
// Precondition: len(slots) == len(inv.slots)
func (inv *Inventory) WriteProtoSlots(slots []proto.WindowSlot) {
	for i := range inv.slots {
		src := &inv.slots[i]
		itemTypeId := ItemTypeIdNull
		if src.ItemType != nil {
			itemTypeId = src.ItemType.Id
		}
		slots[i] = proto.WindowSlot{
			ItemTypeId: itemTypeId,
			Count:      src.Count,
			Data:       src.Data,
		}
	}
}

// Send message about the slot change to the relevant places.
func (inv *Inventory) slotUpdate(slot *slot.Slot, slotId SlotId) {
	if inv.subscriber != nil {
		inv.subscriber.SlotUpdate(slot, slotId)
	}
}

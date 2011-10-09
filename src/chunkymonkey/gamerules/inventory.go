package gamerules

import (
	"os"

	"chunkymonkey/proto"
	"chunkymonkey/types"
	"nbt"
)

type IInventorySubscriber interface {
	// SlotUpdate is called when a slot inside the inventory changes its
	// contents.
	SlotUpdate(slot *Slot, slotId types.SlotId)

	// ProgressUpdate is called when a progress bar inside the inventory changes.
	ProgressUpdate(prgBarId types.PrgBarId, value types.PrgBarValue)
}

// IInventory is the general interface provided by inventory implementations.
type IInventory interface {
	NumSlots() types.SlotId
	Click(click *Click) (txState types.TxState)
	SetSubscriber(subscriber IInventorySubscriber)
	MakeProtoSlots() []proto.WindowSlot
	WriteProtoSlots(slots []proto.WindowSlot)
	TakeAllItems() (items []Slot)
	UnmarshalNbt(tag *nbt.Compound) (err os.Error)
	MarshalNbt(tag *nbt.Compound) (err os.Error)
	SlotUnmarshalNbt(tag *nbt.Compound, slotId types.SlotId) (err os.Error)
}

type Click struct {
	SlotId       types.SlotId
	Cursor       Slot
	RightClick   bool
	ShiftClick   bool
	TxId         types.TxId
	ExpectedSlot Slot
}

type Inventory struct {
	slots      []Slot
	subscriber IInventorySubscriber
}

// Init initializes the inventory. onUnsubscribed is called in a new goroutine
// when the number of subscribers to the inventory reaches zero (but is not
// called initially).
func (inv *Inventory) Init(size int) {
	inv.slots = make([]Slot, size)
}

func (inv *Inventory) NumSlots() types.SlotId {
	return types.SlotId(len(inv.slots))
}

func (inv *Inventory) SetSubscriber(subscriber IInventorySubscriber) {
	inv.subscriber = subscriber
}

// Click takes the default actions upon a click event from a player. The Cursor
// attribute of click may be modified to represent the cursors new contents.
func (inv *Inventory) Click(click *Click) types.TxState {
	if click.SlotId < 0 || int(click.SlotId) > len(inv.slots) {
		return types.TxStateRejected
	}

	clickedSlot := &inv.slots[click.SlotId]

	if !click.ExpectedSlot.Equals(clickedSlot) {
		return types.TxStateRejected
	}

	// Apply the change.
	if click.Cursor.Count == 0 {
		if click.RightClick {
			clickedSlot.Split(&click.Cursor)
		} else {
			clickedSlot.Swap(&click.Cursor)
		}
	} else {
		var changed bool

		if click.RightClick {
			changed = clickedSlot.AddOne(&click.Cursor)
		} else {
			changed = clickedSlot.Add(&click.Cursor)
		}

		if !changed {
			clickedSlot.Swap(&click.Cursor)
		}
	}

	// We send slot updates in case we have custom max counts that differ from
	// the client's idea.
	inv.slotUpdate(clickedSlot, click.SlotId)

	return types.TxStateAccepted
}

// TakeOnlyClick is similar to Click, but only allows items to be taken from
// the slot, and it only allows the *whole* stack to be taken, otherwise no
// items are taken at all. This is intended for use by crafting/furnace output
// slots.
func (inv *Inventory) TakeOnlyClick(click *Click) types.TxState {
	if click.SlotId < 0 || int(click.SlotId) > len(inv.slots) {
		return types.TxStateRejected
	}

	clickedSlot := &inv.slots[click.SlotId]

	if !click.ExpectedSlot.Equals(clickedSlot) {
		return types.TxStateRejected
	}

	// Apply the change.
	click.Cursor.AddWhole(clickedSlot)

	// We send slot updates in case we have custom max counts that differ from
	// the client's idea.
	inv.slotUpdate(clickedSlot, click.SlotId)

	return types.TxStateAccepted
}

func (inv *Inventory) Slot(slotId types.SlotId) Slot {
	return inv.slots[slotId]
}

func (inv *Inventory) TakeOneItem(slotId types.SlotId, into *Slot) {
	slot := &inv.slots[slotId]
	if into.AddOne(slot) {
		inv.slotUpdate(slot, slotId)
	}
}

// PutItem attempts to put the given item into the inventory.
func (inv *Inventory) PutItem(item *Slot) {
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
			inv.slotUpdate(slot, types.SlotId(slotIndex))
		}
	}
}

// CanTakeItem returns true if it can take at least one item from the passed
// Slot.
func (inv *Inventory) CanTakeItem(item *Slot) bool {
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
		slots[i] = proto.WindowSlot{
			ItemTypeId: src.ItemTypeId,
			Count:      src.Count,
			Data:       src.Data,
		}
	}
}

// TakeAllItems empties the inventory, and returns all items that were inside
// it inside a slice of Slots.
func (inv *Inventory) TakeAllItems() (items []Slot) {
	items = make([]Slot, 0, len(inv.slots)-1)

	for i := range inv.slots {
		curSlot := &inv.slots[i]
		if curSlot.Count > 0 {
			var taken Slot
			taken.Swap(curSlot)
			items = append(items, taken)
			inv.slotUpdate(curSlot, types.SlotId(i))
		}
	}

	return
}

// Send message about the slot change to the relevant places.
func (inv *Inventory) slotUpdate(slot *Slot, slotId types.SlotId) {
	if inv.subscriber != nil {
		inv.subscriber.SlotUpdate(slot, slotId)
	}
}

func (inv *Inventory) UnmarshalNbt(tag *nbt.Compound) (err os.Error) {
	itemList, ok := tag.Lookup("Items").(*nbt.List)
	if !ok {
		return os.NewError("bad inventory - not a list")
	}

	for _, slotTagITag := range itemList.Value {
		slotTag, ok := slotTagITag.(*nbt.Compound)
		if !ok {
			return os.NewError("inventory slot not a compound")
		}

		var slotIdTag *nbt.Byte
		if slotIdTag, ok = slotTag.Lookup("Slot").(*nbt.Byte); !ok {
			return os.NewError("Slot ID not a byte")
		}
		slotId := types.SlotId(slotIdTag.Value)

		if err = inv.SlotUnmarshalNbt(slotTag, slotId); err != nil {
			return
		}
	}

	return nil
}

func (inv *Inventory) MarshalNbt(tag *nbt.Compound) (err os.Error) {
	occupiedSlots := 0
	for i := range inv.slots {
		if inv.slots[i].Count > 0 {
			occupiedSlots++
		}
	}

	itemList := &nbt.List{nbt.TagCompound, make([]nbt.ITag, 0, occupiedSlots)}
	for i := range inv.slots {
		slot := &inv.slots[i]
		if slot.Count > 0 {
			slotTag := nbt.NewCompound()
			slotTag.Set("Slot", &nbt.Byte{int8(i)})
			if err = slot.MarshalNbt(slotTag); err != nil {
				return
			}
			itemList.Value = append(itemList.Value, slotTag)
		}
	}

	tag.Set("Items", itemList)

	return nil
}

func (inv *Inventory) SlotUnmarshalNbt(tag *nbt.Compound, slotId types.SlotId) (err os.Error) {
	if slotId < 0 || int(slotId) >= len(inv.slots) {
		return os.NewError("Bad slot ID")
	}
	return inv.slots[slotId].UnmarshalNbt(tag)
}

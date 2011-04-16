package inventory

import (
    "io"
    "os"

    "chunkymonkey/proto"
    .   "chunkymonkey/types"
)

const SlotCountMax = ItemCount(64)

// Represents an inventory slot, e.g in a player's inventory, their cursor, a
// chest.
type Slot struct {
    ItemType ItemId
    Count    ItemCount
    Data     ItemData
}

func (s *Slot) Init() {
    s.ItemType = ItemIdNull
    s.Count = 0
    s.Data = 0
}

func (s *Slot) GetAttr() (ItemId, ItemCount, ItemData) {
    return s.ItemType, s.Count, s.Data
}

func (s *Slot) SendUpdate(writer io.Writer, windowId WindowId, slotId SlotId) os.Error {
    return proto.WriteWindowSetSlot(writer, windowId, slotId, s.ItemType, s.Count, s.Data)
}

func (s *Slot) SendEquipmentUpdate(writer io.Writer, entityId EntityId, slotId SlotId) os.Error {
    return proto.WriteEntityEquipment(writer, entityId, slotId, s.ItemType, s.Data)
}

func (s *Slot) setCount(count ItemCount) {
    s.Count = count
    if s.Count == 0 {
        s.ItemType = ItemIdNull
        s.Data = 0
    }
}

// Adds as many items from the passed slot to the destination (subject) slot as
// possible, depending on stacking allowances and item types etc.
// Returns true if the destination slot changed as a result.
func (s *Slot) Add(src *Slot) (changed bool) {
    // NOTE: This code assumes that 2*SlotCountMax will not overflow
    // the ItemCount type.

    if s.ItemType != ItemIdNull {
        if s.ItemType != src.ItemType {
            return
        }
        if s.Data != src.Data {
            return
        }
    }
    if s.Count >= SlotCountMax {
        return
    }

    s.ItemType = src.ItemType

    toTransfer := src.Count
    if s.Count+toTransfer > SlotCountMax {
        toTransfer = SlotCountMax - s.Count
    }
    if toTransfer != 0 {
        changed = true

        s.Data = src.Data

        s.setCount(s.Count + toTransfer)
        src.setCount(src.Count - toTransfer)
    }
    return
}

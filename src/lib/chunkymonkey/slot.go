package inventory

import (
    "io"
    "os"

    "chunkymonkey/proto"
    .   "chunkymonkey/types"
)

const SlotQuantityMax = ItemCount(64)

// Represents an inventory slot, e.g in a player's inventory, their cursor, a
// chest.
type Slot struct {
    ItemType ItemID
    Quantity ItemCount
    Uses     ItemUses
}

func (s *Slot) Init() {
    s.ItemType = ItemIDNull
    s.Quantity = 0
    s.Uses = 0
}

func (s *Slot) GetAttr() (ItemID, ItemCount, ItemUses) {
    return s.ItemType, s.Quantity, s.Uses
}

func (s *Slot) SendUpdate(writer io.Writer, windowID WindowID, slotID SlotID) os.Error {
    return proto.WriteWindowSetSlot(writer, windowID, slotID, s.ItemType, s.Quantity, s.Uses)
}

func (s *Slot) SendEquipmentUpdate(writer io.Writer, entityID EntityID, slotID SlotID) os.Error {
    return proto.WriteEntityEquipment(writer, entityID, slotID, s.ItemType, s.Uses)
}

func (s *Slot) setQuantity(quantity ItemCount) {
    s.Quantity = quantity
    if s.Quantity == 0 {
        s.ItemType = ItemIDNull
        s.Uses = 0
    }
}

// Adds as many items from the passed slot to the destination (subject) slot as
// possible, depending on stacking allowances and item types etc.
func (s *Slot) Add(src *Slot) {
    // NOTE: This code assumes that 2*SlotQuantityMax will not overflow
    // the ItemCount type.

    if s.ItemType != ItemIDNull {
        if s.ItemType != src.ItemType {
            return
        }
        if s.Uses != src.Uses {
            return
        }
    }
    if s.Quantity >= SlotQuantityMax {
        return
    }

    s.ItemType = src.ItemType

    toTransfer := src.Quantity
    if s.Quantity+toTransfer > SlotQuantityMax {
        toTransfer = SlotQuantityMax - s.Quantity
    }

    s.Uses = src.Uses

    s.setQuantity(s.Quantity + toTransfer)
    src.setQuantity(src.Quantity - toTransfer)
}

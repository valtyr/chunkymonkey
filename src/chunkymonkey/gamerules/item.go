package gamerules

import (
	"io"
	"os"

	"chunkymonkey/physics"
	"chunkymonkey/proto"
	"chunkymonkey/types"
	"nbt"
)

type Item struct {
	types.EntityId
	Slot
	physics.PointObject
	orientation    types.OrientationBytes
	PickupImmunity types.Ticks
}

func NewBlankItem() INonPlayerEntity {
	return new(Item)
}

func NewItem(itemTypeId types.ItemTypeId, count types.ItemCount, data types.ItemData, position *types.AbsXyz, velocity *types.AbsVelocity, pickupImmunity types.Ticks) (item *Item) {
	item = &Item{
		Slot: Slot{
			ItemTypeId: itemTypeId,
			Count:      count,
			Data:       data,
		},
		PickupImmunity: pickupImmunity,
	}
	item.PointObject.Init(position, velocity)
	return
}

func (item *Item) UnmarshalNbt(tag *nbt.Compound) (err os.Error) {
	if err = item.PointObject.UnmarshalNbt(tag); err != nil {
		return
	}

	itemInfo, ok := tag.Lookup("Item").(*nbt.Compound)
	if !ok {
		return os.NewError("bad item data")
	}

	// Grab the basic item data
	id, idOk := itemInfo.Lookup("id").(*nbt.Short)
	count, countOk := itemInfo.Lookup("Count").(*nbt.Byte)
	data, dataOk := itemInfo.Lookup("Damage").(*nbt.Short)
	if !idOk || !countOk || !dataOk {
		return os.NewError("bad item data")
	}

	item.Slot = Slot{
		ItemTypeId: types.ItemTypeId(id.Value),
		Count:      types.ItemCount(count.Value),
		Data:       types.ItemData(data.Value),
	}

	return nil
}

func (item *Item) MarshalNbt(tag *nbt.Compound) (err os.Error) {
	if err = item.PointObject.MarshalNbt(tag); err != nil {
		return
	}
	tag.Set("id", &nbt.String{"Item"})
	tag.Set("Item", &nbt.Compound{map[string]nbt.ITag{
		"id":     &nbt.Short{int16(item.ItemTypeId)},
		"Count":  &nbt.Byte{int8(item.Count)},
		"Damage": &nbt.Short{int16(item.Data)},
	}})
	return nil
}

func (item *Item) GetSlot() *Slot {
	return &item.Slot
}

func (item *Item) SendSpawn(writer io.Writer) (err os.Error) {
	err = proto.WriteItemSpawn(
		writer, item.EntityId, item.ItemTypeId, item.Slot.Count, item.Slot.Data,
		&item.PointObject.LastSentPosition, &item.orientation)
	if err != nil {
		return
	}

	err = proto.WriteEntityVelocity(writer, item.EntityId, &item.PointObject.LastSentVelocity)
	if err != nil {
		return
	}

	return
}

func (item *Item) SendUpdate(writer io.Writer) (err os.Error) {
	if err = proto.WriteEntity(writer, item.EntityId); err != nil {
		return
	}

	err = item.PointObject.SendUpdate(writer, item.EntityId, &types.LookBytes{0, 0})

	return
}

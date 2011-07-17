package gamerules

import (
	"io"
	"os"

	"chunkymonkey/physics"
	"chunkymonkey/proto"
	. "chunkymonkey/types"
)

type Item struct {
	EntityId
	Slot
	physics.PointObject
	orientation OrientationBytes
}

func NewItem(itemTypeId ItemTypeId, count ItemCount, data ItemData, position *AbsXyz, velocity *AbsVelocity) (item *Item) {
	item = &Item{
		Slot: Slot{
			ItemTypeId: itemTypeId,
			Count:      count,
			Data:       data,
		},
	}
	item.PointObject.Init(position, velocity)
	return
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

	err = item.PointObject.SendUpdate(writer, item.EntityId, &LookBytes{0, 0})

	return
}

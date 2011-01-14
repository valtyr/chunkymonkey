package chunkymonkey

import (
    "io"
    "os"
    "chunkymonkey/proto"
    .   "chunkymonkey/types"
)

type PickupItem struct {
    Entity
    itemType    ItemID
    count       ItemCount
    position    XYZInteger
    orientation OrientationPacked
}

func NewPickupItem(game *Game, itemType ItemID, count ItemCount, position XYZInteger) {
    item := &PickupItem{
        itemType: itemType,
        count:    count,
        position: position,
        // TODO proper orientation
        orientation: OrientationPacked{0, 0, 0},
    }

    game.Enqueue(func(game *Game) {
        game.AddPickupItem(item)
    })
}

func (item *PickupItem) SendSpawn(writer io.Writer) os.Error {
    return proto.WritePickupSpawn(writer, item.EntityID, item.itemType, item.count, &item.position, &item.orientation)
}

package gamerules

import (
	"fmt"
	"os"

	. "chunkymonkey/types"
)

// spawnItemInBlock creates an item in a block. It must be run within
// instance.Chunk's goroutine.
func spawnItemInBlock(instance *BlockInstance, itemTypeId ItemTypeId, count ItemCount, data ItemData) {
	rand := instance.Chunk.Rand()
	position := instance.BlockLoc.ToAbsXyz()
	position.X += AbsCoord(blockItemSpawnFromEdge + rand.Float64()*(1-2*blockItemSpawnFromEdge))
	position.Y += AbsCoord(blockItemSpawnFromEdge)
	position.Z += AbsCoord(blockItemSpawnFromEdge + rand.Float64()*(1-2*blockItemSpawnFromEdge))
	instance.Chunk.AddEntity(
		NewItem(
			itemTypeId, count, data,
			position,
			&AbsVelocity{0, 0, 0},
			0,
		),
	)
}

type blockDropItem struct {
	DroppedItem ItemTypeId
	Probability byte // Probabilities specified as a percentage
	Count       ItemCount
	CopyData    bool
}

func (bdi *blockDropItem) drop(instance *BlockInstance) {
	var itemData ItemData
	if bdi.CopyData {
		itemData = ItemData(instance.Data)
	} else {
		itemData = 0
	}

	for i := bdi.Count; i > 0; i-- {
		spawnItemInBlock(instance, bdi.DroppedItem, 1, itemData)
	}
}

func (bdi *blockDropItem) check() os.Error {
	if _, ok := Items[bdi.DroppedItem]; !ok {
		return fmt.Errorf("dropped item type %d does not exist", bdi.DroppedItem)
	}

	if bdi.Count <= 0 {
		return fmt.Errorf("dropped item has Count %d", bdi.Count)
	}

	return nil
}

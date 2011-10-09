package gamerules

import (
	"fmt"
	"os"

	"chunkymonkey/types"
)

// spawnItemInBlock creates an item in a block. It must be run within
// instance.Chunk's goroutine.
func spawnItemInBlock(chunk IChunkBlock, blockLoc types.BlockXyz, itemTypeId types.ItemTypeId, count types.ItemCount, data types.ItemData) {
	rand := chunk.Rand()
	position := blockLoc.ToAbsXyz()
	position.X += types.AbsCoord(blockItemSpawnFromEdge + rand.Float64()*(1-2*blockItemSpawnFromEdge))
	position.Y += types.AbsCoord(blockItemSpawnFromEdge)
	position.Z += types.AbsCoord(blockItemSpawnFromEdge + rand.Float64()*(1-2*blockItemSpawnFromEdge))
	chunk.AddEntity(
		NewItem(
			itemTypeId, count, data,
			position,
			&types.AbsVelocity{0, 0, 0},
			0,
		),
	)
}

type blockDropItem struct {
	DroppedItem types.ItemTypeId
	Probability byte // Probabilities specified as a percentage
	Count       types.ItemCount
	CopyData    bool
}

func (bdi *blockDropItem) drop(chunk IChunkBlock, blockLoc types.BlockXyz, blockData byte) {
	var itemData types.ItemData
	if !bdi.CopyData {
		itemData = 0
	} else {
		itemData = types.ItemData(blockData)
	}

	spawnItemInBlock(chunk, blockLoc, bdi.DroppedItem, bdi.Count, itemData)
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

package block

import (
    "chunkymonkey/item"
    . "chunkymonkey/types"
)

type BlockDropItem struct {
    DroppedItem ItemTypeId
    Probability byte // Probabilities specified as a percentage
    Count       ItemCount
}

func makeStandardAspect() (aspect IBlockAspect) {
    return &StandardAspect{}
}

// Behaviour of a "standard" block. A StandardAspect block is one that is
// diggable, and drops items in a simple manner. StandardAspect blocks do not
// use block metadata.
type StandardAspect struct {
    // Items, up to one of which will potentially spawn when block destroyed.
    DroppedItems []BlockDropItem
    BreakOn      DigStatus
}

func (aspect *StandardAspect) Name() string {
    return "Standard"
}

func (aspect *StandardAspect) Dig(chunk IChunkBlock, blockLoc *BlockXyz, digStatus DigStatus) (destroyed bool) {
    if aspect.BreakOn != digStatus {
        return
    }

    destroyed = true

    if len(aspect.DroppedItems) > 0 {
        rand := chunk.GetRand()
        // Possibly drop item(s)
        r := byte(rand.Intn(100))
        for _, dropItem := range aspect.DroppedItems {
            if dropItem.Probability > r {
                for i := dropItem.Count; i > 0; i-- {
                    position := blockLoc.ToAbsXyz()
                    position.X += AbsCoord(blockItemSpawnFromEdge + rand.Float64()*(1-2*blockItemSpawnFromEdge))
                    position.Y += AbsCoord(blockItemSpawnFromEdge)
                    position.Z += AbsCoord(blockItemSpawnFromEdge + rand.Float64()*(1-2*blockItemSpawnFromEdge))
                    chunk.AddItem(
                        item.NewItem(
                            dropItem.DroppedItem, 1,
                            position,
                            &AbsVelocity{0, 0, 0}))
                }
                break
            }
            r -= dropItem.Probability
        }
    }

    return
}

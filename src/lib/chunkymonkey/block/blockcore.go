package block

import (
    "rand"

    "chunkymonkey/item"
    . "chunkymonkey/types"
)

const (
    BlockIdMin = BlockId(0)
    BlockIdAir = BlockId(0)
    BlockIdMax = BlockId(255)
)

// Defines the behaviour of a block.
type IBlockAspect interface {
    Name() string
    Dig(chunk IChunkBlock, blockLoc *BlockXyz, digStatus DigStatus) (destroyed bool)
}

type BlockAttrs struct {
    Name         string
    Opacity      int8
    defined      bool
    Destructable bool
    Solid        bool
    Replaceable  bool
    Attachable   bool
}

// The core information about any block type.
type BlockType struct {
    BlockAttrs
    Aspect IBlockAspect
}

// Lookup table of blocks.
type BlockTypeList []BlockType

func (btl *BlockTypeList) Get(id BlockId) (block *BlockType, ok bool) {
    if id < 0 || int(id) > len(*btl) {
        ok = false
        return
    }
    block = &(*btl)[id]
    ok = block.defined
    return
}

// The interface required of a chunk by block behaviour.
type IChunkBlock interface {
    GetRand() *rand.Rand
    AddItem(item *item.Item)
}

// The distance from the edge of a block that items spawn at in fractional
// blocks.
const blockItemSpawnFromEdge = 4.0 / PixelsPerBlock

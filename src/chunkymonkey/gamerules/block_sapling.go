package gamerules

import (
	"os"
	"rand"
	"time"
	. "chunkymonkey/types"
)

// Behaviour of a sapling block, takes care of growing or dying depending on
// world conditions.
func makeSaplingAspect() (aspect IBlockAspect) {
	return &SaplingAspect{
		&StandardAspect{},
	}
}

type SaplingAspect struct {
	*StandardAspect
}

func (aspect *SaplingAspect) Name() string {
	return "Sapling"
}

func (aspect *SaplingAspect) Check() os.Error {
	return nil
}

var rnd = rand.New(rand.NewSource(time.Nanoseconds()))

func (aspect *SaplingAspect) Tick(instance *BlockInstance) bool {
	if rnd.Intn(1e4) >= 1e4-1 {
		// Turn this block into a tree
		return aspect.makeTree(instance)
	}
	return true
}

func (aspect *SaplingAspect) makeTree(instance *BlockInstance) bool {
	loc := instance.SubLoc
	minheight := 3
	maxheight := 6
	height := loc.Y + SubChunkCoord(minheight+rand.Intn(maxheight-minheight))

	for y := loc.Y; y < height; y++ {
		loc.Y = y
		index, ok := loc.BlockIndex()
		if !ok {
			return false
		}
		instance.Chunk.SetBlockByIndex(index, BlockId(17), byte(0))
	}
	return true
}

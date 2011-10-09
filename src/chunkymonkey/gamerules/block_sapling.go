package gamerules

import (
	"rand"
	"time"
	"chunkymonkey/types"
	"log"
)

// Behaviour of a sapling block, takes care of growing or dying depending on
// world conditions.
func makeSaplingAspect() (aspect IBlockAspect) {
	return &SaplingAspect{}
}

type SaplingAspect struct {
	StandardAspect
}

func (aspect *SaplingAspect) Name() string {
	return "Sapling"
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
	height := minheight + rand.Intn(maxheight-minheight)
	maxy := loc.Y + types.SubChunkCoord(height)

	for y := loc.Y; y < maxy; y++ {
		loc.Y = y
		index, ok := loc.BlockIndex()
		if !ok {
			// TODO: Can't place a block outside chunk boundaries
			log.Printf("Couldn't place a tree block (%v,%v,%v)", loc.X, loc.Y, loc.Z)
		} else {
			instance.Chunk.SetBlockByIndex(index, types.BlockId(17), byte(0))
		}
	}

	// Store the location at the top block of the tree
	treex, treey, treez := int(loc.X), int(loc.Y), int(loc.Z)
	cradius := height / 2

	// Start one block above the tree and move down
	for y := treey + 1; y >= treey-1; y-- {
		var radius int

		// Slightly round out the canopy
		if y > treey {
			radius = cradius - 1
		} else if y == treey {
			radius = cradius
		} else if y < treey {
			radius = cradius - 1
		}

		for x := treex - radius; x <= treex+radius; x++ {
			for z := treez - radius; z <= treez+radius; z++ {
				if y > treey || x != treex || z != treez {
					loc.X = types.SubChunkCoord(x)
					loc.Y = types.SubChunkCoord(y)
					loc.Z = types.SubChunkCoord(z)
					index, ok := loc.BlockIndex()
					if !ok {
						// TODO: Can't place a block outside chunk boundaries
						log.Printf("Couldn't place a leaf block (%v,%v,%v)", loc.X, loc.Y, loc.Z)
					} else {
						instance.Chunk.SetBlockByIndex(index, types.BlockId(18), byte(0))
					}
				}
			}
		}
	}

	return true
}

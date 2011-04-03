package chunk

import (
    .   "chunkymonkey/interfaces"
    .   "chunkymonkey/types"
)

// Encapsulates logic for looking up block data from neighbouring chunks, and
// for updating neighbouring chunks.
type neighboursCache struct {
    sideCache   [4]chunkSideCache   // Caches neighbouring blocks.
    sideUpdater [4]neighbourUpdater // Updates neighbours' caches.
}

func (n *neighboursCache) init() {
    for i, _ := range n.sideCache {
        n.sideCache[i].init(ChunkSideDir(i))
        n.sideUpdater[i].init(ChunkSideDir(i))
    }
}

func (n *neighboursCache) setBlock(subLoc *SubChunkXYZ, blockType BlockID) {
    if subLoc.X == 0 {
        n.sideUpdater[ChunkSideNorth].blockChanged(subLoc, blockType)
    } else if subLoc.X == ChunkSizeH-1 {
        n.sideUpdater[ChunkSideSouth].blockChanged(subLoc, blockType)
    }

    if subLoc.Z == 0 {
        n.sideUpdater[ChunkSideWest].blockChanged(subLoc, blockType)
    } else if subLoc.Z == ChunkSizeH-1 {
        n.sideUpdater[ChunkSideEast].blockChanged(subLoc, blockType)
    }
}

func (n *neighboursCache) flush() {
    for _, updater := range n.sideUpdater {
        updater.flush()
    }
}

func (n *neighboursCache) sideCacheUpdate(side ChunkSideDir, update *chunkSideCacheUpdate) {
    update.updateCache(&n.sideCache[side])
}

func (n *neighboursCache) sideCacheFullUpdate(side ChunkSideDir, blocks sideBlockData) {
    n.sideCache[side].blocks = blocks
    n.sideCache[side].active = true
}

func (n *neighboursCache) sideCacheSetNeighbour(side ChunkSideDir, neighbour *Chunk, blocks []byte) {
    n.sideUpdater[side].neighbour = neighbour

    // Construct and send a full side cache update.

    var subLoc SubChunkXYZ
    var h *SubChunkCoord
    switch side {
    case ChunkSideNorth:
        subLoc.X = 0
        h = &subLoc.Z
    case ChunkSideSouth:
        subLoc.X = ChunkSizeH - 1
        h = &subLoc.Z
    case ChunkSideWest:
        subLoc.Z = 0
        h = &subLoc.X
    case ChunkSideEast:
        subLoc.Z = ChunkSizeH - 1
        h = &subLoc.X
    }

    var update sideBlockData
    blocksIndex := 0
    for *h = 0; *h < ChunkSizeH; *h++ {
        subLoc.Y = 0
        index, _, _ := blockIndex(&subLoc)
        for y := 0; y < ChunkSizeY; y++ {
            update[blocksIndex] = blocks[index]
            index++
            blocksIndex++
        }
    }

    neighbour.Enqueue(func(_ IChunk) {
        neighbour.neighbours.sideCacheFullUpdate(side.GetOpposite(), update)
    })
}

func (n *neighboursCache) GetCachedBlock(dx, dz ChunkCoord, subLoc *SubChunkXYZ) (ok bool, blockTypeID BlockID) {
    dir, isNeighbour := DXzToDir(int32(dx), int32(dz))
    ok = false

    if !isNeighbour {
        // Do not have information except about directly adjacent neighbours.
        return
    }

    sideCache := n.sideCache[dir]
    if !sideCache.active {
        // We don't have any information about the neighbouring chunk.
        return
    }

    blockTypeID, ok = sideCache.GetCachedBlock(subLoc)
    if !ok {
        return
    }
    ok = true

    return
}

type sideBlockData [ChunkSizeH * ChunkSizeY]byte

// Contains a cache of the blocks on the side of a neighbouring chunk.
type chunkSideCache struct {
    side   ChunkSideDir
    active bool // Has data been received from the chunk yet?
    blocks sideBlockData
}

func (cache *chunkSideCache) init(side ChunkSideDir) {
    cache.side = side
    cache.active = false
}

func (cache *chunkSideCache) GetCachedBlock(subLoc *SubChunkXYZ) (blockType BlockID, ok bool) {
    index, ok := getSideBlockIndex(cache.side, subLoc)
    if !ok {
        return
    }
    blockType = BlockID(cache.blocks[index])
    return
}

// Represents a single block change on the side of a chunk.
type blockChange struct {
    index     int16
    blockType BlockID
}

// Represents a set of individual block changes on the side of a chunk.
type chunkSideCacheUpdate struct {
    changes []blockChange
}

func (update *chunkSideCacheUpdate) updateCache(cache *chunkSideCache) {
    data := cache.blocks
    for _, change := range update.changes {
        data[change.index] = byte(change.blockType)
    }
}

// Used to update one of the 4 neighbouring chunks with changes to the sides of
// the chunk.
type neighbourUpdater struct {
    side      ChunkSideDir
    changes   []blockChange
    neighbour *Chunk
}

func (updater *neighbourUpdater) init(side ChunkSideDir) {
    updater.side = side
    updater.changes = nil
    updater.neighbour = nil
}

func (updater *neighbourUpdater) blockChanged(subLoc *SubChunkXYZ, blockType BlockID) {
    if updater.neighbour == nil {
        // Don't remember changes to send if the neighbouring chunk is not present.
        return
    }

    index, ok := getSideBlockIndex(updater.side, subLoc)
    if !ok {
        return
    }

    updater.changes = append(updater.changes, blockChange{int16(index), blockType})
}

func (updater *neighbourUpdater) flush() {
    if len(updater.changes) > 0 && updater.neighbour != nil {
        update := &chunkSideCacheUpdate{
            changes: updater.changes,
        }
        neighbour := updater.neighbour
        neighbour.Enqueue(func(_ IChunk) {
            neighbour.neighbours.sideCacheUpdate(updater.side, update)
        })
        updater.changes = nil
    }
}

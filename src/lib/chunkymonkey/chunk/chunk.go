// Map chunks

package chunk

import (
    "bytes"
    "log"
    "rand"
    "sync"
    "time"

    "chunkymonkey/block"
    .   "chunkymonkey/interfaces"
    "chunkymonkey/proto"
    .   "chunkymonkey/types"
)

// A chunk is slice of the world map
type Chunk struct {
    mainQueue    chan func(IChunk)
    mgr          *ChunkManager
    loc          ChunkXZ
    blocks       []byte
    blockData    []byte
    blockLight   []byte
    skyLight     []byte
    heightMap    []byte
    items        map[EntityID]IItem
    rand         *rand.Rand
    neighbours   neighboursCache
    cachedPacket []byte                 // Cached packet data for this block.
    subscribers  map[IPacketSender]bool // Subscribers getting updates from the chunk
}

func newChunk(loc *ChunkXZ, mgr *ChunkManager, blocks, blockData, skyLight, blockLight, heightMap []byte) (chunk *Chunk) {
    chunk = &Chunk{
        mainQueue:   make(chan func(IChunk), 256),
        mgr:         mgr,
        loc:         *loc,
        blocks:      blocks,
        blockData:   blockData,
        skyLight:    skyLight,
        blockLight:  blockLight,
        heightMap:   heightMap,
        items:       make(map[EntityID]IItem),
        rand:        rand.New(rand.NewSource(time.UTC().Seconds())),
        subscribers: make(map[IPacketSender]bool),
    }
    chunk.neighbours.init()
    go chunk.mainLoop()
    return
}

func blockIndex(subLoc *SubChunkXYZ) (index int32, shift byte, ok bool) {
    if subLoc.X < 0 || subLoc.Y < 0 || subLoc.Z < 0 || subLoc.X >= ChunkSizeH || subLoc.Y >= ChunkSizeY || subLoc.Z >= ChunkSizeH {
        ok = false
        index = 0
    } else {
        ok = true

        index = int32(subLoc.Y) + (int32(subLoc.Z) * ChunkSizeY) + (int32(subLoc.X) * ChunkSizeY * ChunkSizeH)

        if index%2 == 0 {
            // Low nibble
            shift = 0
        } else {
            // High nibble
            shift = 4
        }
    }
    return
}

// Sets a block and its data. Returns true if the block was not changed.
func (chunk *Chunk) setBlock(blockLoc *BlockXYZ, subLoc *SubChunkXYZ, index int32, shift byte, blockType BlockID, blockMetadata byte) {

    // Invalidate cached packet
    chunk.cachedPacket = nil

    chunk.blocks[index] = byte(blockType)

    mask := byte(0x0f) << shift
    twoBlockData := chunk.blockData[index/2]
    twoBlockData = ((blockMetadata << shift) & mask) | (twoBlockData & ^mask)
    chunk.blockData[index/2] = twoBlockData

    // Tell players that the block changed
    packet := &bytes.Buffer{}
    proto.WriteBlockChange(packet, blockLoc, blockType, blockMetadata)
    chunk.mgr.game.MulticastChunkPacket(packet.Bytes(), &chunk.loc)

    // Update neighbour caches of this change
    chunk.neighbours.setBlock(subLoc, blockType)

    return
}

func (chunk *Chunk) GetLoc() *ChunkXZ {
    return &chunk.loc
}

func (chunk *Chunk) Enqueue(f func(IChunk)) {
    chunk.mainQueue <- f
}

func (chunk *Chunk) mainLoop() {
    for {
        f := <-chunk.mainQueue
        f(chunk)
    }
}

func (chunk *Chunk) GetRand() *rand.Rand {
    return chunk.rand
}

func (chunk *Chunk) AddItem(item IItem) {
    wg := &sync.WaitGroup{}
    wg.Add(1)
    chunk.mgr.game.Enqueue(func(game IGame) {
        entity := item.GetEntity()
        game.AddEntity(entity)
        chunk.items[entity.EntityID] = item
        wg.Done()
    })
    wg.Wait()

    // Spawn new item for players
    buf := &bytes.Buffer{}
    item.SendSpawn(buf)
    chunk.multicastSubscribers(buf.Bytes())
}

func (chunk *Chunk) TransferItem(item IItem) {
    chunk.items[item.GetEntity().EntityID] = item
}

func (chunk *Chunk) GetBlock(subLoc *SubChunkXYZ) (blockType BlockID, ok bool) {
    index, _, ok := blockIndex(subLoc)
    if !ok {
        return
    }

    blockType = BlockID(chunk.blocks[index])

    return
}

func (chunk *Chunk) DestroyBlock(subLoc *SubChunkXYZ) (ok bool) {
    index, shift, ok := blockIndex(subLoc)
    if !ok {
        return
    }

    blockTypeID := BlockID(chunk.blocks[index])
    blockLoc := chunk.loc.ToBlockXYZ(subLoc)

    if blockType, ok := chunk.mgr.blockTypes[blockTypeID]; ok {
        if blockType.Destroy(chunk, blockLoc) {
            chunk.setBlock(blockLoc, subLoc, index, shift, block.BlockIDAir, 0)
        }
    } else {
        log.Printf("Attempted to destroy unknown block ID %d", blockTypeID)
        ok = false
    }

    return
}

func (chunk *Chunk) Tick() {
    // Update neighbouring chunks of block changes in this chunk
    chunk.neighbours.flush()

    blockQuery := func(blockLoc *BlockXYZ) (isSolid bool, isWithinChunk bool) {
        chunkLoc, subLoc := blockLoc.ToChunkLocal()

        // If we are in doubt, we assume that the block asked about is solid
        // (this way items don't fly off the side of the map needlessly)
        isSolid = true

        var blockTypeID BlockID
        if chunkLoc.X == chunk.loc.X && chunkLoc.Z == chunk.loc.Z {
            // The item is asking about this chunk.
            blockTypeID, _ = chunk.GetBlock(subLoc)
            isWithinChunk = true
        } else {
            // The item is asking about a separate chunk.
            isWithinChunk = false

            var ok bool
            ok, blockTypeID = chunk.neighbours.GetCachedBlock(
                chunk.loc.X-chunkLoc.X,
                chunk.loc.Z-chunkLoc.Z,
                subLoc)

            if !ok {
                return
            }
        }

        blockType, ok := chunk.mgr.blockTypes[blockTypeID]
        if !ok {
            log.Printf(
                "game.physicsTick/blockQuery found unknown block type ID %d at %+v",
                blockTypeID, blockLoc)
        } else {
            isSolid = blockType.IsSolid()
        }
        return
    }

    destroyedEntityIDs := []EntityID{}
    leftItems := []IItem{}

    for _, item := range chunk.items {
        if item.Tick(blockQuery) {
            if item.GetPosition().Y <= 0 {
                // Item fell out of the world
                destroyedEntityIDs = append(
                    destroyedEntityIDs, item.GetEntity().EntityID)
            } else {
                leftItems = append(leftItems, item)
            }
        }
    }

    if len(leftItems) > 0 {
        // Remove items from this chunk
        for _, item := range leftItems {
            chunk.items[item.GetEntity().EntityID] = nil, false
        }

        // Send items to new chunk
        chunk.mgr.game.Enqueue(func(game IGame) {
            mgr := game.GetChunkManager()
            for _, item := range leftItems {
                chunkLoc := item.GetPosition().ToChunkXZ()
                blockChunk := mgr.Get(chunkLoc)
                blockChunk.Enqueue(func(blockChunk IChunk) {
                    blockChunk.AddItem(item)
                })
            }
        })
    }

    if len(destroyedEntityIDs) > 0 {
        buf := &bytes.Buffer{}
        for _, entityID := range destroyedEntityIDs {
            proto.WriteEntityDestroy(buf, entityID)
            chunk.items[entityID] = nil, false
        }
        chunk.multicastSubscribers(buf.Bytes())
    }
}

func (chunk *Chunk) AddSubscriber(subscriber IPacketSender) {
    chunk.subscribers[subscriber] = true
    subscriber.TransmitPacket(chunk.chunkPacket())

    // Send spawns of all items in the chunk
    if len(chunk.items) > 0 {
        buf := &bytes.Buffer{}
        for _, item := range chunk.items {
            item.SendSpawn(buf)
        }
        subscriber.TransmitPacket(buf.Bytes())
    }
}

func (chunk *Chunk) RemoveSubscriber(subscriber IPacketSender, sendPacket bool) {
    chunk.subscribers[subscriber] = false, false
    if sendPacket {
        buf := &bytes.Buffer{}
        proto.WritePreChunk(buf, &chunk.loc, ChunkUnload)
        subscriber.TransmitPacket(buf.Bytes())
    }
}

func (chunk *Chunk) multicastSubscribers(packet []byte) {
    for subscriber, _ := range chunk.subscribers {
        subscriber.TransmitPacket(packet)
    }
}

func (chunk *Chunk) chunkPacket() []byte {
    if chunk.cachedPacket == nil {
        buf := &bytes.Buffer{}
        proto.WriteMapChunk(buf, &chunk.loc, chunk.blocks, chunk.blockData, chunk.blockLight, chunk.skyLight)
        chunk.cachedPacket = buf.Bytes()
    }

    return chunk.cachedPacket
}

func (chunk *Chunk) SendUpdate() {
    buf := &bytes.Buffer{}
    for _, item := range chunk.items {
        item.SendUpdate(buf)
    }
    chunk.multicastSubscribers(buf.Bytes())
}

// Used in chunk side caching code:
func getSideBlockIndex(side ChunkSideDir, subLoc *SubChunkXYZ) (index int, ok bool) {
    var h, h2, y SubChunkCoord

    if y >= ChunkSizeY {
        ok = false
        return
    }

    switch side {
    case ChunkSideEast, ChunkSideWest:
        h = subLoc.X
    case ChunkSideNorth, ChunkSideSouth:
        h = subLoc.X
    }

    if h >= ChunkSizeH {
        ok = false
        return
    }

    switch side {
    case ChunkSideWest, ChunkSideNorth:
        if h2 != 0 {
            ok = false
            return
        }
    case ChunkSideEast, ChunkSideSouth:
        if h2 != (ChunkSizeH - 1) {
            ok = false
            return
        }
    }

    ok = true

    y = subLoc.Y

    index = (int(h) * ChunkSizeH) + int(y)

    return
}

func (chunk *Chunk) sideCacheSetNeighbour(side ChunkSideDir, neighbour *Chunk) {
    chunk.neighbours.sideCacheSetNeighbour(side, neighbour, chunk.blocks)
}

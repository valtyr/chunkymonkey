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
    "chunkymonkey/chunkstore"
    .   "chunkymonkey/types"
)

// A chunk is slice of the world map
type Chunk struct {
    mainQueue    chan func(IChunk)
    mgr          *ChunkManager
    loc          ChunkXz
    blocks       []byte
    blockData    []byte
    blockLight   []byte
    skyLight     []byte
    heightMap    []byte
    items        map[EntityId]IItem
    rand         *rand.Rand
    neighbours   neighboursCache
    cachedPacket []byte                 // Cached packet data for this block.
    subscribers  map[IPacketSender]bool // Subscribers getting updates from the chunk
}

func newChunkFromReader(reader chunkstore.ChunkReader, mgr *ChunkManager) (chunk *Chunk) {
    chunk = &Chunk{
        mainQueue:   make(chan func(IChunk), 256),
        mgr:         mgr,
        loc:         *reader.ChunkLoc(),
        blocks:      reader.Blocks(),
        blockData:   reader.BlockData(),
        skyLight:    reader.SkyLight(),
        blockLight:  reader.BlockLight(),
        heightMap:   reader.HeightMap(),
        items:       make(map[EntityId]IItem),
        rand:        rand.New(rand.NewSource(time.UTC().Seconds())),
        subscribers: make(map[IPacketSender]bool),
    }
    chunk.neighbours.init()
    go chunk.mainLoop()
    return
}

func blockIndex(subLoc *SubChunkXyz) (index int32, shift byte, ok bool) {
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
func (chunk *Chunk) setBlock(blockLoc *BlockXyz, subLoc *SubChunkXyz, index int32, shift byte, blockType BlockId, blockMetadata byte) {

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
    chunk.multicastSubscribers(packet.Bytes())

    // Update neighbour caches of this change
    chunk.neighbours.setBlock(subLoc, blockType)

    return
}

func (chunk *Chunk) GetLoc() *ChunkXz {
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
        chunk.items[entity.EntityId] = item
        wg.Done()
    })
    wg.Wait()

    // Spawn new item for players
    buf := &bytes.Buffer{}
    item.SendSpawn(buf)
    chunk.multicastSubscribers(buf.Bytes())
}

func (chunk *Chunk) TransferItem(item IItem) {
    chunk.items[item.GetEntity().EntityId] = item
}

func (chunk *Chunk) GetBlock(subLoc *SubChunkXyz) (blockType BlockId, ok bool) {
    index, _, ok := blockIndex(subLoc)
    if !ok {
        return
    }

    blockType = BlockId(chunk.blocks[index])

    return
}

func (chunk *Chunk) DestroyBlock(subLoc *SubChunkXyz) (ok bool) {
    index, shift, ok := blockIndex(subLoc)
    if !ok {
        return
    }

    blockTypeId := BlockId(chunk.blocks[index])
    blockLoc := chunk.loc.ToBlockXyz(subLoc)

    if blockType, ok := chunk.mgr.blockTypes[blockTypeId]; ok {
        if blockType.Destroy(chunk, blockLoc) {
            chunk.setBlock(blockLoc, subLoc, index, shift, block.BlockIdAir, 0)
        }
    } else {
        log.Printf("Attempted to destroy unknown block Id %d", blockTypeId)
        ok = false
    }

    return
}

func (chunk *Chunk) Tick() {
    // Update neighbouring chunks of block changes in this chunk
    chunk.neighbours.flush()

    blockQuery := func(blockLoc *BlockXyz) (isSolid bool, isWithinChunk bool) {
        chunkLoc, subLoc := blockLoc.ToChunkLocal()

        // If we are in doubt, we assume that the block asked about is solid
        // (this way items don't fly off the side of the map needlessly)
        isSolid = true

        var blockTypeId BlockId
        if chunkLoc.X == chunk.loc.X && chunkLoc.Z == chunk.loc.Z {
            // The item is asking about this chunk.
            blockTypeId, _ = chunk.GetBlock(subLoc)
            isWithinChunk = true
        } else {
            // The item is asking about a separate chunk.
            isWithinChunk = false

            var ok bool
            ok, blockTypeId = chunk.neighbours.GetCachedBlock(
                chunk.loc.X-chunkLoc.X,
                chunk.loc.Z-chunkLoc.Z,
                subLoc)

            if !ok {
                return
            }
        }

        blockType, ok := chunk.mgr.blockTypes[blockTypeId]
        if !ok {
            log.Printf(
                "game.physicsTick/blockQuery found unknown block type Id %d at %+v",
                blockTypeId, blockLoc)
        } else {
            isSolid = blockType.IsSolid()
        }
        return
    }

    destroyedEntityIds := []EntityId{}
    leftItems := []IItem{}

    for _, item := range chunk.items {
        if item.Tick(blockQuery) {
            if item.GetPosition().Y <= 0 {
                // Item fell out of the world
                destroyedEntityIds = append(
                    destroyedEntityIds, item.GetEntity().EntityId)
            } else {
                leftItems = append(leftItems, item)
            }
        }
    }

    if len(leftItems) > 0 {
        // Remove items from this chunk
        for _, item := range leftItems {
            chunk.items[item.GetEntity().EntityId] = nil, false
        }

        // Send items to new chunk
        chunk.mgr.game.Enqueue(func(game IGame) {
            mgr := game.GetChunkManager()
            for _, item := range leftItems {
                chunkLoc := item.GetPosition().ToChunkXz()
                blockChunk := mgr.Get(chunkLoc)
                blockChunk.Enqueue(func(blockChunk IChunk) {
                    blockChunk.AddItem(item)
                })
            }
        })
    }

    if len(destroyedEntityIds) > 0 {
        buf := &bytes.Buffer{}
        for _, entityId := range destroyedEntityIds {
            proto.WriteEntityDestroy(buf, entityId)
            chunk.items[entityId] = nil, false
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

func (chunk *Chunk) sideCacheSetNeighbour(side ChunkSideDir, neighbour *Chunk) {
    chunk.neighbours.sideCacheSetNeighbour(side, neighbour, chunk.blocks)
}

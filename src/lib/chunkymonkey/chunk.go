// Map chunks

package chunkymonkey

import (
    "bytes"
    "io"
    "log"
    "os"
    "path"
    "rand"
    "time"

    "chunkymonkey/proto"
    .   "chunkymonkey/interfaces"
    .   "chunkymonkey/types"
    "nbt/nbt"
)

// A chunk is slice of the world map
type Chunk struct {
    mainQueue  chan func(IChunk)
    mgr        *ChunkManager
    loc        ChunkXZ
    blocks     []byte
    blockData  []byte
    blockLight []byte
    skyLight   []byte
    heightMap  []byte
    items      map[EntityID]IItem
    rand       *rand.Rand
}

func newChunk(loc *ChunkXZ, mgr *ChunkManager, blocks, blockData, skyLight, blockLight, heightMap []byte) (chunk *Chunk) {
    chunk = &Chunk{
        mainQueue:  make(chan func(IChunk), 256),
        mgr:        mgr,
        loc:        *loc,
        blocks:     blocks,
        blockData:  blockData,
        skyLight:   skyLight,
        blockLight: blockLight,
        heightMap:  heightMap,
        rand:       rand.New(rand.NewSource(time.UTC().Seconds())),
        items:      make(map[EntityID]IItem),
    }
    go chunk.mainLoop()
    return
}

func blockIndex(subLoc *SubChunkXYZ) (index int32, shift byte, ok bool) {
    if subLoc.X < 0 || subLoc.Y < 0 || subLoc.Z < 0 || subLoc.X >= ChunkSizeX || subLoc.Y >= ChunkSizeY || subLoc.Z >= ChunkSizeZ {
        ok = false
        index = 0
    } else {
        ok = true

        index = int32(subLoc.Y) + (int32(subLoc.Z) * ChunkSizeY) + (int32(subLoc.X) * ChunkSizeY * ChunkSizeZ)

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
func (chunk *Chunk) setBlock(blockLoc *BlockXYZ, index int32, shift byte, blockType BlockID, blockMetadata byte) {
    chunk.blocks[index] = byte(blockType)

    mask := byte(0x0f) << shift
    twoBlockData := chunk.blockData[index/2]
    twoBlockData = ((blockMetadata << shift) & mask) | (twoBlockData & ^mask)
    chunk.blockData[index/2] = twoBlockData

    // Tell players that the block changed
    packet := &bytes.Buffer{}
    proto.WriteBlockChange(packet, blockLoc, blockType, blockMetadata)
    chunk.mgr.game.MulticastChunkPacket(packet.Bytes(), &chunk.loc)

    return
}

func (chunk *Chunk) GetLoc() *ChunkXZ {
    return &chunk.loc
}

// This is the only method that's safe to call from outside the chunk's own
// goroutine. Everything else must be called via this function.
func (chunk *Chunk) Enqueue(f func(IChunk)) {
    chunk.mainQueue <- f
}

func (chunk *Chunk) mainLoop() {
    for {
        f := <-chunk.mainQueue
        f(chunk)
    }
}

// Tells the chunk to take posession of the item.
func (chunk *Chunk) TransferItem(item IItem) {
    chunk.items[item.GetEntity().EntityID] = item
}

func (chunk *Chunk) SendUpdate() {
    // TODO send only to players in range
    buf := &bytes.Buffer{}
    for _, item := range chunk.items {
        item.SendUpdate(buf)
    }
    chunk.mgr.game.Enqueue(func(game IGame) {
        game.MulticastPacket(buf.Bytes(), nil)
    })
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
            chunk.setBlock(blockLoc, index, shift, BlockIDAir, 0)
        }
    } else {
        log.Printf("Attempted to destroy unknown block ID %d", blockTypeID)
        ok = false
    }

    return
}

func (chunk *Chunk) AddItem(item IItem) {
    chunk.mgr.game.Enqueue(func(game IGame) {
        entity := item.GetEntity()
        game.AddEntity(entity)
        chunk.items[entity.EntityID] = item

        // Spawn new item for players
        buf := &bytes.Buffer{}
        err := item.SendSpawn(buf)
        if err != nil {
            log.Print("AddItem", err.String())
            return
        }
        game.MulticastChunkPacket(buf.Bytes(), &chunk.loc)
    })
}

func (chunk *Chunk) PhysicsTick() {
    blockQuery := func(blockLoc *BlockXYZ) (isSolid bool, isWithinChunk bool) {
        chunkLoc, subLoc := blockLoc.ToChunkLocal()
        var blockTypeID BlockID
        if chunkLoc.X == chunk.loc.X && chunkLoc.Z == chunk.loc.Z {
            // The item is asking about this chunk
            blockTypeID, _ = chunk.GetBlock(subLoc)
        } else {
            // The item is asking about a seperate chunk
            isWithinChunk = false

            blockTypeIDChan := make(chan int)
            chunk.mgr.game.Enqueue(func(game IGame) {
                blockChunk := chunk.mgr.Get(chunkLoc)

                if chunk == nil {
                    // Object fell off the side of the world
                    blockTypeIDChan <- -1
                    return
                }

                blockChunk.Enqueue(func(blockChunk IChunk) {
                    blockTypeID, _ = blockChunk.GetBlock(subLoc)
                    blockTypeIDChan <- int(blockTypeID)
                })
            })
            blockTypeIDRaw := <-blockTypeIDChan
            if blockTypeIDRaw < 0 {
                // Continuation of object falling off the side of the world
                isSolid = false
                return
            }
            blockTypeID = BlockID(blockTypeIDRaw)
        }

        blockType, ok := chunk.mgr.blockTypes[blockTypeID]
        if !ok {
            log.Print("game.physicsTick/blockQuery found unknown block type ID", blockTypeID)
            // Assume this unknown block is solid
            isSolid = true
        } else {
            isSolid = blockType.IsSolid
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
        // TODO send only to players in range
        chunk.mgr.game.MulticastPacket(buf.Bytes(), nil)
    }
}

// Send chunk data down network connection
func (chunk *Chunk) SendChunkData(writer io.Writer) (err os.Error) {
    return proto.WriteMapChunk(writer, &chunk.loc, chunk.blocks, chunk.blockData, chunk.blockLight, chunk.skyLight)
}

// ChunkManager contains all chunks and can look them up
type ChunkManager struct {
    game       IGame
    blockTypes map[BlockID]*BlockType
    worldPath  string
    chunks     map[uint64]*Chunk
}

func NewChunkManager(worldPath string) *ChunkManager {
    return &ChunkManager{
        worldPath: worldPath,
        chunks:    make(map[uint64]*Chunk),
    }
}

func base36Encode(n int32) (s string) {
    alphabet := "0123456789abcdefghijklmnopqrstuvwxyz"
    negative := false

    if n < 0 {
        n = -n
        negative = true
    }
    if n == 0 {
        return "0"
    }

    for n != 0 {
        i := n % int32(len(alphabet))
        n /= int32(len(alphabet))
        s = string(alphabet[i:i+1]) + s
    }
    if negative {
        s = "-" + s
    }
    return
}

func (mgr *ChunkManager) chunkPath(loc *ChunkXZ) string {
    return path.Join(mgr.worldPath, base36Encode(int32(loc.X&63)), base36Encode(int32(loc.Z&63)),
        "c."+base36Encode(int32(loc.X))+"."+base36Encode(int32(loc.Z))+".dat")
}

// Load a chunk from its NBT representation
func (mgr *ChunkManager) loadChunk(reader io.Reader) (chunk *Chunk, err os.Error) {
    level, err := nbt.Read(reader)
    if err != nil {
        return
    }

    chunk = newChunk(
        &ChunkXZ{
            X: ChunkCoord(level.Lookup("/Level/xPos").(*nbt.Int).Value),
            Z: ChunkCoord(level.Lookup("/Level/zPos").(*nbt.Int).Value),
        },
        mgr,
        level.Lookup("/Level/Blocks").(*nbt.ByteArray).Value,
        level.Lookup("/Level/Data").(*nbt.ByteArray).Value,
        level.Lookup("/Level/SkyLight").(*nbt.ByteArray).Value,
        level.Lookup("/Level/BlockLight").(*nbt.ByteArray).Value,
        level.Lookup("/Level/HeightMap").(*nbt.ByteArray).Value,
    )

    return
}

// Get a chunk at given coordinates
func (mgr *ChunkManager) Get(loc *ChunkXZ) (chunk IChunk) {
    // FIXME this function looks subject to race conditions with itself
    key := uint64(loc.X)<<32 | uint64(uint32(loc.Z))
    chunk, ok := mgr.chunks[key]
    if ok {
        return
    }

    file, err := os.Open(mgr.chunkPath(loc), os.O_RDONLY, 0)
    if err != nil {
        log.Fatalf("ChunkManager.Get: %s", err.String())
    }
    defer file.Close()

    loaded_chunk, err := mgr.loadChunk(file)

    if err != nil {
        log.Fatalf("ChunkManager.loadChunk: %s", err.String())
    }

    mgr.chunks[key] = loaded_chunk
    chunk = loaded_chunk
    return
}

// Return a channel to iterate over all chunks within a chunk's radius
func (mgr *ChunkManager) ChunksInRadius(loc *ChunkXZ) (c chan IChunk) {
    c = make(chan IChunk)
    go func() {
        curChunkXZ := ChunkXZ{0, 0}
        for z := loc.Z - ChunkRadius; z <= loc.Z+ChunkRadius; z++ {
            for x := loc.X - ChunkRadius; x <= loc.X+ChunkRadius; x++ {
                curChunkXZ.X, curChunkXZ.Z = x, z
                c <- mgr.Get(&curChunkXZ)
            }
        }
        close(c)
    }()
    return
}

// Return a channel to iterate over all chunks within a player's radius
func (mgr *ChunkManager) ChunksInPlayerRadius(player IPlayer) chan IChunk {
    locChan := make(chan *ChunkXZ)
    player.Enqueue(func(player IPlayer) {
        locChan<-player.GetChunkPosition()
    })
    return mgr.ChunksInRadius(<-locChan)
}

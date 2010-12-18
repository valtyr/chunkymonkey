// Map chunks

package main

import (
    "io"
    "os"
    "log"
    "path"
    "nbt/nbt"
)

// A chunk is slice of the world map
type Chunk struct {
    X, Z       ChunkCoord
    Blocks     []byte
    BlockData  []byte
    SkyLight   []byte
    BlockLight []byte
    HeightMap  []byte
}

func blockIndex(x, y, z SubChunkCoord) (index int32, shift byte, err bool) {
    if x < 0 || y < 0 || z < 0 || x >= ChunkSizeX || y >= ChunkSizeY || z >= ChunkSizeZ {
        err = true
        index = 0
    } else {
        err = false

        index = int32(y + (z * ChunkSizeY) + (x * ChunkSizeY * ChunkSizeZ))

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
func (chunk *Chunk) SetBlock(x, y, z SubChunkCoord, blockType BlockID, blockMetadata byte) (err bool) {
    index, shift, err := blockIndex(x, y, z)
    if err {
        return
    }

    chunk.Blocks[index] = byte(blockType)

    mask := byte(0x0f) << shift
    twoBlockData := chunk.BlockData[index/2]
    twoBlockData = ((blockMetadata << shift) & mask) | (twoBlockData & ^mask)
    chunk.BlockData[index/2] = twoBlockData

    return
}

// Load a chunk from its NBT representation
func loadChunk(reader io.Reader) (chunk *Chunk, err os.Error) {
    level, err := nbt.Read(reader)
    if err != nil {
        return
    }

    chunk = &Chunk{
        X:          ChunkCoord(level.Lookup("/Level/xPos").(*nbt.Int).Value),
        Z:          ChunkCoord(level.Lookup("/Level/zPos").(*nbt.Int).Value),
        Blocks:     level.Lookup("/Level/Blocks").(*nbt.ByteArray).Value,
        BlockData:  level.Lookup("/Level/Data").(*nbt.ByteArray).Value,
        SkyLight:   level.Lookup("/Level/SkyLight").(*nbt.ByteArray).Value,
        BlockLight: level.Lookup("/Level/BlockLight").(*nbt.ByteArray).Value,
        HeightMap:  level.Lookup("/Level/HeightMap").(*nbt.ByteArray).Value,
    }
    return
}

// ChunkManager contains all chunks and can look them up
type ChunkManager struct {
    worldPath string
    chunks    map[uint64]*Chunk
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

func (mgr *ChunkManager) chunkPath(x ChunkCoord, z ChunkCoord) string {
    return path.Join(mgr.worldPath, base36Encode(int32(x&63)), base36Encode(int32(z&63)),
        "c."+base36Encode(int32(x))+"."+base36Encode(int32(z))+".dat")
}

// Get a chunk at given coordinates
func (mgr *ChunkManager) Get(x ChunkCoord, z ChunkCoord) (chunk *Chunk) {
    key := uint64(x)<<32 | uint64(uint32(z))
    chunk, ok := mgr.chunks[key]
    if ok {
        return
    }

    file, err := os.Open(mgr.chunkPath(x, z), os.O_RDONLY, 0)
    if err != nil {
        log.Exit("ChunkManager.Get: ", err.String())
    }

    chunk, err = loadChunk(file)
    file.Close()
    if err != nil {
        log.Exit("ChunkManager.loadChunk: ", err.String())
    }

    mgr.chunks[key] = chunk
    return
}

// Return a channel to iterate over all chunks within a chunk's radius
func (mgr *ChunkManager) ChunksInRadius(chunkX ChunkCoord, chunkZ ChunkCoord) (c chan *Chunk) {
    c = make(chan *Chunk)
    go func() {
        for z := chunkZ - ChunkRadius; z <= chunkZ+ChunkRadius; z++ {
            for x := chunkX - ChunkRadius; x <= chunkX+ChunkRadius; x++ {
                c <- mgr.Get(x, z)
            }
        }
        close(c)
    }()
    return
}

// Return a channel to iterate over all chunks within a player's radius
func (mgr *ChunkManager) ChunksInPlayerRadius(player *Player) chan *Chunk {
    playerX, playerZ := AbsoluteToChunkCoords(player.position.x, player.position.z)
    return mgr.ChunksInRadius(playerX, playerZ)
}

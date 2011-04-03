package chunk

import (
    "io"
    "log"
    "os"
    "path"

    .   "chunkymonkey/interfaces"
    .   "chunkymonkey/types"
    "nbt"
)

// ChunkManager contains all chunks and can look them up
type ChunkManager struct {
    game       IGame
    blockTypes map[BlockID]IBlockType
    worldPath  string
    chunks     map[uint64]*Chunk
}

func NewChunkManager(worldPath string, game IGame) *ChunkManager {
    return &ChunkManager{
        worldPath:  worldPath,
        chunks:     make(map[uint64]*Chunk),
        game:       game,
        blockTypes: game.GetBlockTypes(),
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
            X:  ChunkCoord(level.Lookup("/Level/xPos").(*nbt.Int).Value),
            Z:  ChunkCoord(level.Lookup("/Level/zPos").(*nbt.Int).Value),
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
func (mgr *ChunkManager) Get(loc *ChunkXZ) (c IChunk) {
    var chunk *Chunk
    key := loc.ChunkKey()
    chunk, ok := mgr.chunks[key]
    if ok {
        c = chunk
        return
    }

    file, err := os.Open(mgr.chunkPath(loc), os.O_RDONLY, 0)
    if err != nil {
        log.Printf("ChunkManager.Get: %s", err.String())
        return nil
    }
    defer file.Close()

    chunk, err = mgr.loadChunk(file)

    if err != nil {
        log.Printf("ChunkManager.loadChunk: %s", err.String())
        return nil
    }

    c = chunk

    // Notify neighbouring chunk(s) (if any) that this chunk is now active, and
    // notify this chunk of its active neighbours
    linkNeighbours := func(from ChunkSideDir) {
        dx, dz := from.GetDXz()
        loc := ChunkXZ{
            X:  loc.X + dx,
            Z:  loc.Z + dz,
        }
        neighbour, exists := mgr.chunks[loc.ChunkKey()]
        if exists {
            to := from.GetOpposite()
            chunk.Enqueue(func(_ IChunk) {
                chunk.sideCacheSetNeighbour(from, neighbour)
            })
            neighbour.Enqueue(func(_ IChunk) {
                neighbour.sideCacheSetNeighbour(to, chunk)
            })
        }
    }
    // TODO corresponding unlinking when a chunk is unloaded
    linkNeighbours(ChunkSideEast)
    linkNeighbours(ChunkSideSouth)
    linkNeighbours(ChunkSideWest)
    linkNeighbours(ChunkSideNorth)

    mgr.chunks[key] = chunk
    return
}

func (mgr *ChunkManager) ChunksActive() <-chan IChunk {
    c := make(chan IChunk)
    go func() {
        for _, chunk := range mgr.chunks {
            c <- chunk
        }
        close(c)
    }()
    return c
}

// Return a channel to iterate over all chunks within a chunk's radius
func (mgr *ChunkManager) ChunksInRadius(loc *ChunkXZ) <-chan IChunk {
    c := make(chan IChunk)
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
    return c
}

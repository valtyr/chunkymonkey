// Reads the Minecraft Alpha world format.
package store_alpha

import (
    "fmt"
    "io"
    "os"
    "path"

    "chunkymonkey/chunk/store"
    .   "chunkymonkey/types"
    "nbt"
)

type ChunkStoreAlpha struct {
    worldPath string
}

func NewChunkStoreAlpha(worldPath string) store.ChunkStore {
    return &ChunkStoreAlpha{
        worldPath: worldPath,
    }
}

func (s *ChunkStoreAlpha) chunkPath(chunkLoc *ChunkXZ) string {
    return path.Join(
        s.worldPath,
        base36Encode(int32(chunkLoc.X&63)),
        base36Encode(int32(chunkLoc.Z&63)),
        "c."+base36Encode(int32(chunkLoc.X))+"."+base36Encode(int32(chunkLoc.Z))+".dat")
}

// Load a chunk from its NBT representation
func (s *ChunkStoreAlpha) LoadChunk(chunkLoc *ChunkXZ) (reader store.ChunkReader, err os.Error) {
    if err != nil {
        return
    }

    file, err := os.Open(s.chunkPath(chunkLoc), os.O_RDONLY, 0)
    if err != nil {
        return
    }
    defer file.Close()

    reader, err = newChunkReader(file)
    if err != nil {
        return
    }

    loadedLoc := reader.ChunkLoc()
    if loadedLoc.X != chunkLoc.X || loadedLoc.Z != chunkLoc.Z {
        err = os.NewError(fmt.Sprintf(
            "Attempted to load chunk for %+v, but got chunk identified as %+v",
            chunkLoc,
            loadedLoc,
        ))
    }

    return
}

// Returned to chunks to pull their data from.
type chunkReader struct {
    chunkTag *nbt.NamedTag
}

func newChunkReader(reader io.Reader) (r *chunkReader, err os.Error) {
    chunkTag, err := nbt.Read(reader)
    if err != nil {
        return
    }

    r = &chunkReader{
        chunkTag: chunkTag,
    }

    return
}

// Returns the chunk location.
func (r *chunkReader) ChunkLoc() *ChunkXZ {
    return &ChunkXZ{
        X:  ChunkCoord(r.chunkTag.Lookup("/Level/xPos").(*nbt.Int).Value),
        Z:  ChunkCoord(r.chunkTag.Lookup("/Level/zPos").(*nbt.Int).Value),
    }
}

// Returns the block IDs in the chunk.
func (r *chunkReader) Blocks() []byte {
    return r.chunkTag.Lookup("/Level/Blocks").(*nbt.ByteArray).Value
}

// Returns the block data in the chunk.
func (r *chunkReader) BlockData() []byte {
    return r.chunkTag.Lookup("/Level/Data").(*nbt.ByteArray).Value
}

// Returns the block light data in the chunk.
func (r *chunkReader) BlockLight() []byte {
    return r.chunkTag.Lookup("/Level/BlockLight").(*nbt.ByteArray).Value
}

// Returns the sky light data in the chunk.
func (r *chunkReader) SkyLight() []byte {
    return r.chunkTag.Lookup("/Level/SkyLight").(*nbt.ByteArray).Value
}

// Returns the height map data in the chunk.
func (r *chunkReader) HeightMap() []byte {
    return r.chunkTag.Lookup("/Level/HeightMap").(*nbt.ByteArray).Value
}

// Utility functions:

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

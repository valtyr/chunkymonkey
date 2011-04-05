package store

import (
    "os"

    .   "chunkymonkey/types"
)

type ChunkStore interface {
    LoadChunk(chunkLoc *ChunkXZ) (reader ChunkReader, err os.Error)
}

type ChunkReader interface {
    // Returns the chunk location.
    ChunkLoc() *ChunkXZ

    // Returns the block IDs in the chunk.
    Blocks() []byte

    // Returns the block data in the chunk.
    BlockData() []byte

    // Returns the block light data in the chunk.
    BlockLight() []byte

    // Returns the sky light data in the chunk.
    SkyLight() []byte

    // Returns the height map data in the chunk.
    HeightMap() []byte
}

package chunkstore

import (
    "fmt"
    "os"

    .   "chunkymonkey/types"
    "nbt"
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

// Given the NamedTag for a level.dat, returns an appropriate ChunkStore.
// TODO consolidate chunk store with level.dat data into an appropriate type.
func ChunkStoreForLevel(worldPath string, level *nbt.NamedTag) (store ChunkStore, err os.Error) {
    versionTag, ok := level.Lookup("/Data/version").(*nbt.Int)

    if !ok {
        store = NewChunkStoreAlpha(worldPath)
    } else {
        switch version := versionTag.Value; version {
        case 19132:
            store = NewChunkStoreBeta(worldPath)
        default:
            err = UnknownLevelVersion(version)
        }
    }

    return
}

type UnknownLevelVersion int32

func (err UnknownLevelVersion) String() string {
    return fmt.Sprintf("Unknown level version %d", err)
}

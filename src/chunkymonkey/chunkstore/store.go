package chunkstore

import (
	"fmt"
	"os"

	"chunkymonkey/gamerules"
	. "chunkymonkey/types"
	"nbt"
)

type ChunkReadResult struct {
	Reader IChunkReader
	Err    os.Error
}

type IChunkStore interface {
	// Serve() serves requests in the foreground.
	Serve()

	ReadChunk(chunkLoc ChunkXz) (result <-chan ChunkReadResult)
	SupportsWrite() bool
	Writer() IChunkWriter
	WriteChunk(writer IChunkWriter)
}

type IChunkReader interface {
	// Returns the chunk location.
	ChunkLoc() ChunkXz

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

	// Return a list of the entities (items, mobs) within the chunk.
	Entities() []gamerules.INonPlayerEntity

	// For low-level NBT access. Not for regular use. It's possible that this
	// might return nil if the underlying system doesn't use NBT.
	RootTag() nbt.ITag
}

// IChunkWriter is the interface for objects that accept chunk data and write
// it. These are created by IChunkWriteableStore for use by a chunk to store a
// snapshot of its current state into. The Set* functions make copies of the
// data passed in, so that the original data structures passed in can be
// modified upon return.
type IChunkWriter interface {
	// Returns the chunk location.
	ChunkLoc() ChunkXz

	// Sets the chunk location.
	SetChunkLoc(loc ChunkXz)

	// Sets the block IDs in the chunk.
	SetBlocks(blocks []byte)

	// Sets the block data in the chunk.
	SetBlockData(blockData []byte)

	// Sets the block light data in the chunk.
	BlockLight(blockLight []byte)

	// Sets the sky light data in the chunk.
	SkyLight(skyLight []byte)

	// Sets the height map data in the chunk.
	HeightMap(heightMap []byte)

	// Sets a list of the entities (items, mobs) within the chunk.
	Entities(entities []gamerules.INonPlayerEntity)

	// Submits the set chunk data for writing. The chunk writer must not be
	// altered any further after calling this (i.e it is passed to another
	// goroutine for processing).
	Commit()
}

// Given the NamedTag for a level.dat, returns an appropriate
// IChunkStoreForeground.
func ChunkStoreForLevel(worldPath string, levelData nbt.ITag, dimension DimensionId) (store IChunkStoreForeground, err os.Error) {
	versionTag, ok := levelData.Lookup("Data/version").(*nbt.Int)

	if !ok {
		store = newChunkStoreAlpha(worldPath, dimension)
	} else {
		switch version := versionTag.Value; version {
		case 19132:
			store = newChunkStoreBeta(worldPath, dimension)
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

type NoSuchChunkError bool

func (err NoSuchChunkError) String() string {
	return "Chunk does not exist."
}

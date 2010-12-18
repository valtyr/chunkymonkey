package main

const (
    // Chunk coordinates can be converted to block coordinates
    ChunkSizeX = 16
    ChunkSizeY = 128
    ChunkSizeZ = 16

    // The area within which a client receives updates
    ChunkRadius = 10
)

// Block face (0-5)
type Face byte

// Specifies exact world location in pixels
type AbsoluteCoord float64

// Specifies approximate world coordinate in pixels (absolute / 32 ?)
// TODO verify the physical size of values of this type
type AbsoluteCoordInteger int32

// Coordinate of a block within the world (integer version of AbsoluteCoord)
type BlockCoord int32

// Coordinate of a chunk in the world (block / 16)
type ChunkCoord int32

// An angle in radians
type AngleRadians float32

// Convert an (x, z) absolute coordinate pair to chunk coordinates
func AbsoluteToChunkCoords(absX AbsoluteCoord, absZ AbsoluteCoord) (chunkX ChunkCoord, chunkZ ChunkCoord) {
    return ChunkCoord(absX / ChunkSizeX), ChunkCoord(absZ / ChunkSizeZ)
}

// Convert an (x, z) block coordinate pair to chunk coordinates
func BlockToChunkCoords(blockX AbsoluteCoord, blockZ AbsoluteCoord) (chunkX ChunkCoord, chunkZ ChunkCoord) {
    return ChunkCoord(blockX / ChunkSizeX), ChunkCoord(blockZ / ChunkSizeZ)
}

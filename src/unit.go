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

// Coordinate of a block within a chunk
type SubChunkCoord int32

// An angle in radians
type AngleRadians float32

// Convert an (x, z) absolute coordinate pair to chunk coordinates
func AbsoluteToChunkCoords(absX, absZ AbsoluteCoord) (chunkX, chunkZ ChunkCoord) {
    return ChunkCoord(absX / ChunkSizeX), ChunkCoord(absZ / ChunkSizeZ)
}

// Convert an (x, z) block coordinate pair to chunk coordinates and the
// coordinates of the block within the chunk
func BlockToChunkCoords(blockX, blockZ BlockCoord) (chunkX, chunkZ ChunkCoord, subX, subZ SubChunkCoord) {
    chunkX = ChunkCoord(blockX / ChunkSizeX)
    subX = SubChunkCoord(blockX % ChunkSizeX)
    if subX < 0 {
        subX += ChunkSizeX
    }
    chunkZ = ChunkCoord(blockZ / ChunkSizeZ)
    subZ = SubChunkCoord(blockZ % ChunkSizeZ)
    if subZ < 0 {
        subZ += ChunkSizeZ
    }
    return
}

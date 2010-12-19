package chunkymonkey

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

type XYZ struct {
    x, y, z AbsoluteCoord
}

type Orientation struct {
    rotation AngleRadians
    pitch    AngleRadians
}

type ChunkXZ struct {
    x, z ChunkCoord
}

type BlockXYZ struct {
    x, y, z BlockCoord
}

type SubChunkXYZ struct {
    x, y, z SubChunkCoord
}

// Convert an (x, z) absolute coordinate pair to chunk coordinates
func (abs XYZ) ToChunkXZ() (chunkXz ChunkXZ) {
    return ChunkXZ{
        ChunkCoord(abs.x / ChunkSizeX),
        ChunkCoord(abs.z / ChunkSizeZ),
    }
}

// Convert an (x, z) block coordinate pair to chunk coordinates and the
// coordinates of the block within the chunk
func (blockLoc BlockXYZ) ToChunkLocal() (chunkLoc ChunkXZ, subLoc SubChunkXYZ) {
    chunkLoc = ChunkXZ{
        ChunkCoord(blockLoc.x / ChunkSizeX),
        ChunkCoord(blockLoc.z / ChunkSizeZ),
    }

    subX := SubChunkCoord(blockLoc.x % ChunkSizeX)
    if subX < 0 {
        subX += ChunkSizeX
    }
    subZ := SubChunkCoord(blockLoc.z % ChunkSizeZ)
    if subZ < 0 {
        subZ += ChunkSizeZ
    }

    subLoc = SubChunkXYZ{subX, SubChunkCoord(blockLoc.y), subZ}
    return
}

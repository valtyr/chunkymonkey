package chunkymonkey

const (
    // Chunk coordinates can be converted to block coordinates
    ChunkSizeX = 16
    ChunkSizeY = 128
    ChunkSizeZ = 16

    // The area within which a client receives updates
    ChunkRadius = 10

    // Sometimes it is useful to convert block coordinates to pixels
    PixelsPerBlock = 32
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

type XYZInteger struct {
    x, y, z AbsoluteCoordInteger
}

type Orientation struct {
    rotation AngleRadians
    pitch    AngleRadians
}

type OrientationPacked struct {
    rotation, pitch, roll byte
}

type ChunkXZ struct {
    x, z ChunkCoord
}

// Convert a position within a chunk to a block position within the world
func (chunkLoc *ChunkXZ) ToBlockXY(subLoc *SubChunkXYZ) *BlockXYZ {
    return &BlockXYZ{
        BlockCoord(chunkLoc.x)*ChunkSizeX + BlockCoord(subLoc.x),
        BlockCoord(subLoc.y),
        BlockCoord(chunkLoc.z)*ChunkSizeZ + BlockCoord(subLoc.z),
    }
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

// Convert (x, z) absolute integer coordinates to chunk coordinates
func (abs XYZInteger) ToChunkXZ() ChunkXZ {
    // TODO check this conversion
    return ChunkXZ{
        ChunkCoord(abs.x / ChunkSizeX),
        ChunkCoord(abs.z / ChunkSizeZ),
    }
}

func coordDivMod(num, denom int32) (div, mod int32) {
    div = num / denom
    mod = num % denom
    if mod < 0 {
        mod += denom
        div -= 1
    }
    return
}

// Convert an (x, z) block coordinate pair to chunk coordinates and the
// coordinates of the block within the chunk
func (blockLoc BlockXYZ) ToChunkLocal() (chunkLoc ChunkXZ, subLoc SubChunkXYZ) {
    chunkX, subX := coordDivMod(int32(blockLoc.x), ChunkSizeX)
    chunkZ, subZ := coordDivMod(int32(blockLoc.z), ChunkSizeZ)

    chunkLoc = ChunkXZ{ChunkCoord(chunkX), ChunkCoord(chunkZ)}
    subLoc = SubChunkXYZ{SubChunkCoord(subX), SubChunkCoord(blockLoc.y), SubChunkCoord(subZ)}
    return
}

func (blockLoc BlockXYZ) ToXYZInteger() XYZInteger {
    // TODO check this conversion
    return XYZInteger{
        AbsoluteCoordInteger(blockLoc.x * PixelsPerBlock),
        AbsoluteCoordInteger(blockLoc.y * PixelsPerBlock),
        AbsoluteCoordInteger(blockLoc.z * PixelsPerBlock),
    }
}

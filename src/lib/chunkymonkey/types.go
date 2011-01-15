package types

// Defines the basic types such as ID types, and world units.

type ItemCount byte

type ItemID int16

type EntityID int32

type BlockID byte

type DigStatus byte

const (
    DigStarted    = DigStatus(0)
    DigDigging    = DigStatus(1)
    DigStopped    = DigStatus(2)
    DigBlockBroke = DigStatus(3)
)

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

type RelMoveCoord byte

// Specifies exact world location in pixels
type AbsoluteCoord float64

// Specifies approximate world coordinate in pixels (absolute / 32 ?)
// TODO verify the physical size of values of this type
type AbsoluteCoordInteger int32

// Coordinate of a block within the world (integer version of AbsoluteCoord)
type BlockCoord int32
type BlockYCoord byte

// Coordinate of a chunk in the world (block / 16)
type ChunkCoord int32

// Coordinate of a block within a chunk
type SubChunkCoord byte

// An angle, where there are 256 units in a circle.
type AngleByte byte

// An angle in radians
type AngleRadians float32

type RelMove struct {
    X, Y, Z RelMoveCoord
}

type XYZ struct {
    X, Y, Z AbsoluteCoord
}

type XYZInteger struct {
    X, Y, Z AbsoluteCoordInteger
}

// FIXME go through code and seperate the uses into degrees and radian types of
// Orientation, rename Rotation to Yaw
type Orientation struct {
    Rotation AngleRadians
    Pitch    AngleRadians
}

type OrientationPacked struct {
    Rotation, Pitch, Roll AngleByte
}

type ChunkXZ struct {
    X, Z ChunkCoord
}

// Convert a position within a chunk to a block position within the world
func (chunkLoc *ChunkXZ) ToBlockXY(subLoc *SubChunkXYZ) *BlockXYZ {
    return &BlockXYZ{
        BlockCoord(chunkLoc.X)*ChunkSizeX + BlockCoord(subLoc.X),
        BlockYCoord(subLoc.Y),
        BlockCoord(chunkLoc.Z)*ChunkSizeZ + BlockCoord(subLoc.Z),
    }
}

type BlockXYZ struct {
    X   BlockCoord
    Y   BlockYCoord
    Z   BlockCoord
}

type SubChunkXYZ struct {
    X, Y, Z SubChunkCoord
}

// Convert an (x, z) absolute coordinate pair to chunk coordinates
func (abs XYZ) ToChunkXZ() (chunkXz ChunkXZ) {
    return ChunkXZ{
        ChunkCoord(abs.X / ChunkSizeX),
        ChunkCoord(abs.Z / ChunkSizeZ),
    }
}

// Convert (x, z) absolute integer coordinates to chunk coordinates
func (abs XYZInteger) ToChunkXZ() ChunkXZ {
    // TODO check this conversion
    return ChunkXZ{
        ChunkCoord(abs.X / ChunkSizeX),
        ChunkCoord(abs.Z / ChunkSizeZ),
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
    chunkX, subX := coordDivMod(int32(blockLoc.X), ChunkSizeX)
    chunkZ, subZ := coordDivMod(int32(blockLoc.Z), ChunkSizeZ)

    chunkLoc = ChunkXZ{ChunkCoord(chunkX), ChunkCoord(chunkZ)}
    subLoc = SubChunkXYZ{SubChunkCoord(subX), SubChunkCoord(blockLoc.Y), SubChunkCoord(subZ)}
    return
}

func (blockLoc BlockXYZ) ToXYZInteger() XYZInteger {
    // TODO check this conversion
    return XYZInteger{
        AbsoluteCoordInteger(blockLoc.X * PixelsPerBlock),
        AbsoluteCoordInteger(blockLoc.Y * PixelsPerBlock),
        AbsoluteCoordInteger(blockLoc.Z * PixelsPerBlock),
    }
}

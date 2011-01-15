package types

// Defines the basic types such as ID types, and world units.

type TimeOfDay int64

type RandomSeed int64

// Which 'world'?
type DimensionID int8

const (
    DimensionNether = DimensionID(-1)
    DimensionNormal = DimensionID(0)
)

// Item-related types

// Item type ID
type ItemID int16

// Number of times that an item has been used
type ItemUses int16

// How many items are in a stack/slot etc.
type ItemCount byte

// Entity-related types

type EntityID int32

// The type of mob
type EntityMobType byte

type EntityStatus byte

type PlayerAnimation byte

// Block-related types

type BlockID byte

// Block face (0-5)
type Face byte

// Action-related types and constants

type DigStatus byte

const (
    DigStarted    = DigStatus(0)
    DigDigging    = DigStatus(1)
    DigStopped    = DigStatus(2)
    DigBlockBroke = DigStatus(3)
)

// Window/inventory-related types and constants

// ID specifying which slotted window, such as inventory
// TODO get numeric constants for these
type WindowID byte

// ID of the slow in inventory or other item-slotted window element
type SlotID int16

// Transaction ID
type TxID int16

// Movement-related types and constants

type VelocityComponent int16

type Velocity struct {
    X, Y, Z VelocityComponent
}

type RelMoveCoord byte

type RelMove struct {
    X, Y, Z RelMoveCoord
}

// Angle-related types and constants

const (
    DegreesToBytes = 256 / 360
)

// An angle, where there are 256 units in a circle.
type AngleBytes byte

// An angle in degrees
type AngleDegrees float32

type LookDegrees struct {
    Yaw, Pitch AngleDegrees
}

func (l *LookDegrees) ToLookBytes() *LookBytes {
    return &LookBytes{
        AngleBytes(l.Yaw * DegreesToBytes),
        AngleBytes(l.Pitch * DegreesToBytes),
    }
}

type LookBytes struct {
    Yaw, Pitch AngleBytes
}

type OrientationDegrees struct {
    Yaw, Pitch, Roll AngleDegrees
}

type OrientationBytes struct {
    Yaw, Pitch, Roll AngleBytes
}

// Location-related types and constants

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

// Specifies exact world distance in blocks (floating point)
type AbsCoord float64

type AbsXYZ struct {
    X, Y, Z AbsCoord
}

func (p *AbsXYZ) ToAbsIntXYZ() *AbsIntXYZ {
    return &AbsIntXYZ{
        AbsIntCoord(p.X * PixelsPerBlock),
        AbsIntCoord(p.Y * PixelsPerBlock),
        AbsIntCoord(p.Z * PixelsPerBlock),
    }
}

func (p *AbsXYZ) ToBlockXYZ() *BlockXYZ {
    return &BlockXYZ{
        BlockCoord(p.X),
        BlockYCoord(p.Y),
        BlockCoord(p.Z),
    }
}

// Specifies approximate world distance in pixels (absolute / PixelsPerBlock)
type AbsIntCoord int32

type AbsIntXYZ struct {
    X, Y, Z AbsIntCoord
}

// Coordinate of a chunk in the world (block / 16)
type ChunkCoord int32

type ChunkXZ struct {
    X, Z ChunkCoord
}

// Returns the world BlockXYZ position of the (0, 0, 0) block in the chunk
func (chunkLoc *ChunkXZ) GetChunkCornerBlockXY() *BlockXYZ {
    return &BlockXYZ{
        BlockCoord(chunkLoc.X) * ChunkSizeX,
        0,
        BlockCoord(chunkLoc.Z) * ChunkSizeZ,
    }
}

// Convert a position within a chunk to a block position within the world
func (chunkLoc *ChunkXZ) ToBlockXY(subLoc *SubChunkXYZ) *BlockXYZ {
    return &BlockXYZ{
        BlockCoord(chunkLoc.X)*ChunkSizeX + BlockCoord(subLoc.X),
        BlockYCoord(subLoc.Y),
        BlockCoord(chunkLoc.Z)*ChunkSizeZ + BlockCoord(subLoc.Z),
    }
}

// Size of a sub-chunk
type SubChunkSizeCoord byte

// Coordinate of a block within a chunk
type SubChunkCoord byte

type SubChunkSize struct {
    X, Y, Z SubChunkSizeCoord
}

type SubChunkXYZ struct {
    X, Y, Z SubChunkCoord
}

// Coordinate of a block within the world
type BlockCoord int32
type BlockYCoord byte

type BlockXYZ struct {
    X   BlockCoord
    Y   BlockYCoord
    Z   BlockCoord
}

// Convert an (x, z) absolute coordinate pair to chunk coordinates
func (abs *AbsXYZ) ToChunkXZ() (chunkXz ChunkXZ) {
    return ChunkXZ{
        ChunkCoord(abs.X / ChunkSizeX),
        ChunkCoord(abs.Z / ChunkSizeZ),
    }
}

// Convert (x, z) absolute integer coordinates to chunk coordinates
func (abs *AbsIntXYZ) ToChunkXZ() ChunkXZ {
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
func (blockLoc *BlockXYZ) ToChunkLocal() (chunkLoc ChunkXZ, subLoc SubChunkXYZ) {
    chunkX, subX := coordDivMod(int32(blockLoc.X), ChunkSizeX)
    chunkZ, subZ := coordDivMod(int32(blockLoc.Z), ChunkSizeZ)

    chunkLoc = ChunkXZ{ChunkCoord(chunkX), ChunkCoord(chunkZ)}
    subLoc = SubChunkXYZ{SubChunkCoord(subX), SubChunkCoord(blockLoc.Y), SubChunkCoord(subZ)}
    return
}

func (blockLoc *BlockXYZ) ToXYZInteger() AbsIntXYZ {
    // TODO check this conversion
    return AbsIntXYZ{
        AbsIntCoord(blockLoc.X * PixelsPerBlock),
        AbsIntCoord(blockLoc.Y * PixelsPerBlock),
        AbsIntCoord(blockLoc.Z * PixelsPerBlock),
    }
}

// Misc. types and constants

type ChunkLoadMode byte

const (
    // Client should unload the chunk
    ChunkUnload = ChunkLoadMode(0)

    // Client should initialise the chunk
    ChunkInit = ChunkLoadMode(1)
)

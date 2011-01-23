package types

import (
    "math"
)

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

const ItemIDNull = ItemID(-1)

// Number of times that an item has been used
type ItemUses int16

// How many items are in a stack/slot etc.
type ItemCount byte

// Entity-related types

type EntityID int32

// The type of mob
type EntityMobType byte

type EntityStatus byte

type EntityAnimation byte

const (
    EntityAnimationNone     = EntityAnimation(0)
    EntityAnimationSwingArm = EntityAnimation(1)
    EntityAnimationDamage   = EntityAnimation(2)
    EntityAnimationUnknown1 = EntityAnimation(102)
    EntityAnimationCrouch   = EntityAnimation(104)
    EntityAnimationUncrouch = EntityAnimation(105)
)

type EntityAction byte

const (
    EntityActionCrouch   = EntityAction(1)
    EntityActionUncrouch = EntityAction(2)
)

type ObjTypeID int8

const (
    ObjTypeIDBoat           = ObjTypeID(1)
    ObjTypeIDMinecart       = ObjTypeID(10)
    ObjTypeIDStorageCart    = ObjTypeID(11)
    ObjTypeIDPoweredCart    = ObjTypeID(12)
    ObjTypeIDActivatedTNT   = ObjTypeID(50)
    ObjTypeIDArrow          = ObjTypeID(60)
    ObjTypeIDThrownSnowball = ObjTypeID(61)
    ObjTypeIDThrownEgg      = ObjTypeID(62)
    ObjTypeIDFallingSand    = ObjTypeID(70)
    ObjTypeIDFallingGravel  = ObjTypeID(71)
    ObjTypeIDFishingFloat   = ObjTypeID(90)
)

type PaintingTypeID int32

type InstrumentID byte

const (
    InstrumentIDDoubleBass = InstrumentID(1)
    InstrumentIDSnareDrum  = InstrumentID(2)
    InstrumentIDSticks     = InstrumentID(3)
    InstrumentIDBassDrum   = InstrumentID(4)
    InstrumentIDHarp       = InstrumentID(5)
)

type NotePitch byte

const (
    NotePitchMin = NotePitch(0)
    NotePitchMax = NotePitch(24)
)

// Block-related types

type BlockID byte

// Block face (0-5)
type Face int8

// Used when a block face is not appropriate to the situation, but block
// location data passed (such as using an item not on a block).
const FaceNull = Face(-1)

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
type WindowID int8

type InvTypeID byte

const (
    InvTypeIDChest     = InvTypeID(0)
    InvTypeIDWorkbench = InvTypeID(1)
    InvTypeIDFurnace   = InvTypeID(2)
)

// ID of the slow in inventory or other item-slotted window element
type SlotID int16

type PrgBarID int16

const (
    PrgBarIDFurnaceProgress = PrgBarID(0)
    PrgBarIDFurnaceFire     = PrgBarID(1)
)

type PrgBarValue int16

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
    DegreesToBytes = 256.0 / 360.0
)

// An angle, where there are 256 units in a circle.
type AngleBytes byte

// An angle in degrees
type AngleDegrees float32

func (d *AngleDegrees) ToAngleBytes() AngleBytes {
    norm := math.Fmod(float64(*d), 360)
    if norm < 0 {
        norm = 360 + norm
    }
    return AngleBytes(norm * DegreesToBytes)
}

type LookDegrees struct {
    Yaw, Pitch AngleDegrees
}

func (l *LookDegrees) ToLookBytes() *LookBytes {
    return &LookBytes{
        l.Yaw.ToAngleBytes(),
        l.Pitch.ToAngleBytes(),
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
func (chunkLoc *ChunkXZ) ToBlockXYZ(subLoc *SubChunkXYZ) *BlockXYZ {
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
type BlockYCoord int8

type BlockXYZ struct {
    X   BlockCoord
    Y   BlockYCoord
    Z   BlockCoord
}

// Test if a block location is not appropriate to the situation, but block
// location data passed (such as using an item not on a block).
func (b *BlockXYZ) IsNull() bool {
    return b.Y == -1 && b.X == -1 && b.Z == -1
}

// Convert an (x, z) absolute coordinate pair to chunk coordinates
func (abs *AbsXYZ) ToChunkXZ() (chunkXz *ChunkXZ) {
    return &ChunkXZ{
        ChunkCoord(abs.X / ChunkSizeX),
        ChunkCoord(abs.Z / ChunkSizeZ),
    }
}

// Convert (x, z) absolute integer coordinates to chunk coordinates
func (abs *AbsIntXYZ) ToChunkXZ() *ChunkXZ {
    chunkX, _ := coordDivMod(int32(abs.X), ChunkSizeX*PixelsPerBlock)
    chunkZ, _ := coordDivMod(int32(abs.Z), ChunkSizeZ*PixelsPerBlock)

    return &ChunkXZ{
        ChunkCoord(chunkX),
        ChunkCoord(chunkZ),
    }
}

func (abs *AbsIntXYZ) IAdd(dx, dy, dz AbsIntCoord) {
    abs.X += dx
    abs.Y += dy
    abs.Z += dz
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
func (blockLoc *BlockXYZ) ToChunkLocal() (chunkLoc *ChunkXZ, subLoc *SubChunkXYZ) {
    chunkX, subX := coordDivMod(int32(blockLoc.X), ChunkSizeX)
    chunkZ, subZ := coordDivMod(int32(blockLoc.Z), ChunkSizeZ)

    chunkLoc = &ChunkXZ{ChunkCoord(chunkX), ChunkCoord(chunkZ)}
    subLoc = &SubChunkXYZ{SubChunkCoord(subX), SubChunkCoord(blockLoc.Y), SubChunkCoord(subZ)}
    return
}

func (blockLoc *BlockXYZ) ToAbsIntXYZ() *AbsIntXYZ {
    return &AbsIntXYZ{
        AbsIntCoord(blockLoc.X) * PixelsPerBlock,
        AbsIntCoord(blockLoc.Y) * PixelsPerBlock,
        AbsIntCoord(blockLoc.Z) * PixelsPerBlock,
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

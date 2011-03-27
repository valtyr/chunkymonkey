package types

import (
    "math"
)

// Defines the basic types such as ID types, and world units.

type TimeOfDay int64

const (
    DayTicksPerDay    = TimeOfDay(24000)
    DayTicksPerSecond = TimeOfDay(20)
)

// 1 "TickTime" is the duration of a server "tick". This value is intended for
// use in sub-tick physics calculations.
type TickTime float64

const (
    TicksPerSecond      = 5
    DayTicksPerTick     = DayTicksPerSecond / TicksPerSecond
    NanosecondsInSecond = 1e9
)

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
    DigBlockBroke = DigStatus(2)
    // TODO Investigate what this value means:
    DigDropItem = DigStatus(4)
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

// VelocityComponent in millipixels / tick
type VelocityComponent int16

const (
    VelocityComponentMax = 28800
    VelocityComponentMin = -28800

    MaxVelocityBlocksPerTick = VelocityComponentMax / AbsVelocityCoord(MilliPixelsPerBlock)
    MinVelocityBlocksPerTick = VelocityComponentMin / AbsVelocityCoord(MilliPixelsPerBlock)
)

type Velocity struct {
    X, Y, Z VelocityComponent
}

type AbsVelocityCoord AbsCoord

func (v AbsVelocityCoord) ToVelocityComponent() VelocityComponent {
    return VelocityComponent(v * MilliPixelsPerBlock)
}

type AbsVelocity struct {
    X, Y, Z AbsVelocityCoord
}

func (v *AbsVelocity) ToVelocity() *Velocity {
    return &Velocity{
        v.X.ToVelocityComponent(),
        v.Y.ToVelocityComponent(),
        v.Z.ToVelocityComponent(),
    }
}

func (v *AbsVelocityCoord) Constrain() {
    if *v > MaxVelocityBlocksPerTick {
        *v = MaxVelocityBlocksPerTick
    } else if *v < MinVelocityBlocksPerTick {
        *v = MinVelocityBlocksPerTick
    }
}

// Relative movement, using same units as AbsIntCoord, but in byte form so
// constrained
type RelMoveCoord int8

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
    // Pitch is -ve when looking above the horizontal, and +ve below
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

// Cardinal directions
type ChunkSideDir int

const (
    ChunkSideEast  = 0
    ChunkSideSouth = 1
    ChunkSideWest  = 2
    ChunkSideNorth = 3
)

func (d ChunkSideDir) GetDXz() (dx, dz ChunkCoord) {
    switch d {
    case ChunkSideEast:
        dx = 0
        dz = 1
    case ChunkSideSouth:
        dx = 1
        dz = 0
    case ChunkSideWest:
        dx = 0
        dz = -1
    case ChunkSideNorth:
        dx = -1
        dz = 0
    }
    return
}

func (d ChunkSideDir) GetOpposite() ChunkSideDir {
    switch d {
    case ChunkSideEast:
        return ChunkSideWest
    case ChunkSideSouth:
        return ChunkSideNorth
    case ChunkSideWest:
        return ChunkSideEast
    case ChunkSideNorth:
        return ChunkSideSouth
    }
    // Should not happen (should we panic on this?)
    return ChunkSideNorth
}

// Returns the direction that (dx,dz) is in. Exactly one of dx and dz must be
// -1 or 1, and the other must be 0, otherwide ok will return as false.
func DXzToDir(dx, dz int32) (dir ChunkSideDir, ok bool) {
    ok = true
    if dz == 0 {
        if dx == -1 {
            dir = ChunkSideNorth
        } else if dx == 1 {
            dir = ChunkSideSouth
        } else {
            ok = false
        }
    } else if dx == 0 {
        if dz == -1 {
            dir = ChunkSideWest
        } else if dz == 1 {
            dir = ChunkSideEast
        } else {
            ok = false
        }
    } else {
        ok = false
    }
    return
}

// Location-related types and constants

const (
    // Chunk coordinates can be converted to block coordinates
    ChunkSizeH = 16
    ChunkSizeY = 128

    // The area within which a client receives updates
    ChunkRadius = 10

    // Sometimes it is useful to convert block coordinates to pixels
    PixelsPerBlock = 32

    // Millipixels are used in velocity values
    MilliPixelsPerPixel = 1000
    MilliPixelsPerBlock = PixelsPerBlock * MilliPixelsPerPixel
)

// Specifies exact world distance in blocks (floating point)
type AbsCoord float64

type AbsXYZ struct {
    X, Y, Z AbsCoord
}

func (p *AbsXYZ) ApplyVelocity(dt TickTime, v *AbsVelocity) {
    p.X += AbsCoord(float64(v.X) * float64(dt))
    p.Y += AbsCoord(float64(v.Y) * float64(dt))
    p.Z += AbsCoord(float64(v.Z) * float64(dt))
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
        BlockCoord(math.Floor(float64(p.X))),
        BlockYCoord(math.Floor(float64(p.Y))),
        BlockCoord(math.Floor(float64(p.Z))),
    }
}

// Specifies approximate world distance in pixels (absolute / PixelsPerBlock)
type AbsIntCoord int32

type AbsIntXYZ struct {
    X, Y, Z AbsIntCoord
}

func (p *AbsIntXYZ) ToBlockXYZ() *BlockXYZ {
    return &BlockXYZ{
        BlockCoord(p.X / PixelsPerBlock),
        BlockYCoord(p.Y / PixelsPerBlock),
        BlockCoord(p.Z / PixelsPerBlock),
    }
}

// Coordinate of a chunk in the world (block / 16)
type ChunkCoord int32

type ChunkXZ struct {
    X, Z ChunkCoord
}

// Returns the world BlockXYZ position of the (0, 0, 0) block in the chunk
func (chunkLoc *ChunkXZ) GetChunkCornerBlockXY() *BlockXYZ {
    return &BlockXYZ{
        BlockCoord(chunkLoc.X) * ChunkSizeH,
        0,
        BlockCoord(chunkLoc.Z) * ChunkSizeH,
    }
}

// Convert a position within a chunk to a block position within the world
func (chunkLoc *ChunkXZ) ToBlockXYZ(subLoc *SubChunkXYZ) *BlockXYZ {
    return &BlockXYZ{
        BlockCoord(chunkLoc.X)*ChunkSizeH + BlockCoord(subLoc.X),
        BlockYCoord(subLoc.Y),
        BlockCoord(chunkLoc.Z)*ChunkSizeH + BlockCoord(subLoc.Z),
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
        ChunkCoord(abs.X / ChunkSizeH),
        ChunkCoord(abs.Z / ChunkSizeH),
    }
}

// Convert (x, z) absolute integer coordinates to chunk coordinates
func (abs *AbsIntXYZ) ToChunkXZ() *ChunkXZ {
    chunkX, _ := coordDivMod(int32(abs.X), ChunkSizeH*PixelsPerBlock)
    chunkZ, _ := coordDivMod(int32(abs.Z), ChunkSizeH*PixelsPerBlock)

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
    chunkX, subX := coordDivMod(int32(blockLoc.X), ChunkSizeH)
    chunkZ, subZ := coordDivMod(int32(blockLoc.Z), ChunkSizeH)

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

func (blockLoc *BlockXYZ) ToAbsXYZ() *AbsXYZ {
    return &AbsXYZ{
        AbsCoord(blockLoc.X),
        AbsCoord(blockLoc.Y),
        AbsCoord(blockLoc.Z),
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

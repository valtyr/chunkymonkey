package chunkymonkey

import (
    "io"
    "os"
    "fmt"
    "bytes"
    "encoding/binary"
    "compress/zlib"

    . "chunkymonkey/types"
)

const (
    // Currently only this protocol version is supported
    protocolVersion = 8

    // Packet type IDs
    packetIDKeepAlive            = 0x00
    packetIDLogin                = 0x01
    packetIDHandshake            = 0x02
    packetIDChatMessage          = 0x03
    packetIDTimeUpdate           = 0x04
    packetIDPlayerInventory      = 0x05
    packetIDSpawnPosition        = 0x06
    packetIDUseEntity            = 0x07
    packetIDUpdateHealth         = 0x08
    packetIDFlying               = 0x0a
    packetIDPlayerPosition       = 0x0b
    packetIDPlayerLook           = 0x0c
    packetIDPlayerPositionLook   = 0x0d
    packetIDPlayerDigging        = 0x0e
    packetIDPlayerBlockPlacement = 0x0f
    packetIDHoldingChange        = 0x10
    packetIDArmAnimation         = 0x12
    packetIDNamedEntitySpawn     = 0x14
    packetIDPickupSpawn          = 0x15
    packetIDMobSpawn             = 0x18
    packetIDUnknownX19           = 0x19
    packetIDEntityVelocity       = 0x1c
    packetIDEntity               = 0x1e
    packetIDEntityDestroy        = 0x1d
    packetIDEntityRelMove        = 0x1f
    packetIDEntityLook           = 0x20
    packetIDEntityLookAndRelMove = 0x21
    packetIDEntityTeleport       = 0x22
    packetIDEntityStatus         = 0x26
    packetIDUnknownX28           = 0x28
    packetIDPreChunk             = 0x32
    packetIDMapChunk             = 0x33
    packetIDBlockChangeMulti     = 0x34
    packetIDBlockChange          = 0x35
    packetIDUnknownX36           = 0x36
    packetIDWindowClick          = 0x66
    packetIDSetSlot              = 0x67
    packetIDWindowItems          = 0x68
    packetIDDisconnect           = 0xff

    // Inventory types
    inventoryTypeMain     = -1
    inventoryTypeArmor    = -2
    inventoryTypeCrafting = -3
)

// Callers must implement this interface to receive packets
type PacketHandler interface {
    PacketKeepAlive()
    PacketChatMessage(message string)
    PacketFlying(onGround bool)
    PacketPlayerPosition(position *XYZ, stance AbsoluteCoord, onGround bool)
    PacketPlayerLook(orientation *Orientation, onGround bool)
    PacketPlayerDigging(status DigStatus, blockLoc *BlockXYZ, face Face)
    PacketPlayerBlockPlacement(blockItemID int16, blockLoc *BlockXYZ, direction Face)
    PacketHoldingChange(blockItemID int16)
    PacketArmAnimation(forward bool)
    PacketDisconnect(reason string)
}

type CSPacketHandler interface {
    PacketHandler
}

type SCPacketHandler interface {
    PacketHandler
    SCPacketLogin(entityID EntityID, str1 string, str2 string, mapSeed int64, dimension byte)
    SCPacketTimeUpdate(time int64)
    SCPacketSpawnPosition(position *BlockXYZ)
    SCPacketUpdateHealth(health int16)
    SCPacketMobSpawn(entityID EntityID, mobType byte, position *XYZInteger, yaw byte, pitch byte)
    SCPacketPreChunk(position *ChunkXZ, mode bool)
    SCPacketMapChunk(position *BlockXYZ, sizeX, sizeY, sizeZ byte, data []byte)
}

// Common protocol helper functions

func boolToByte(b bool) byte {
    if b {
        return 1
    }
    return 0
}

func byteToBool(b byte) bool {
    return b != 0
}

func ReadString(reader io.Reader) (s string, err os.Error) {
    var length int16
    err = binary.Read(reader, binary.BigEndian, &length)
    if err != nil {
        return
    }

    bs := make([]byte, uint16(length))
    _, err = io.ReadFull(reader, bs)
    return string(bs), err
}

func WriteString(writer io.Writer, s string) (err os.Error) {
    bs := []byte(s)

    err = binary.Write(writer, binary.BigEndian, int16(len(bs)))
    if err != nil {
        return
    }

    _, err = writer.Write(bs)
    return
}

// Reads extra data from the end of certain packets, whose meaning isn't known
// yet. Currently all this code does is read and discard bytes.
// TODO update to pull useful data out as it becomes understood
// http://pastebin.com/HHW52Awn
func readUnknownExtra(reader io.Reader) (err os.Error) {
    var entryType byte

    var byteVal byte
    var int16Val int16
    var int32Val int32
    var floatVal float32
    var position struct {
        X   int16
        Y   byte
        Z   int16
    }

    for {
        err = binary.Read(reader, binary.BigEndian, &entryType)
        if err != nil {
            return
        }
        if entryType == 127 {
            break
        }

        switch entryTypeEnum := (entryType & 0xe0) >> 5; entryTypeEnum {
        case 0:
            err = binary.Read(reader, binary.BigEndian, &byteVal)
        case 1:
            err = binary.Read(reader, binary.BigEndian, &int16Val)
        case 2:
            err = binary.Read(reader, binary.BigEndian, &int32Val)
        case 3:
            err = binary.Read(reader, binary.BigEndian, &floatVal)
        case 4:
            _, err = ReadString(reader)
        case 5:
            err = binary.Read(reader, binary.BigEndian, &position)
        }

        if err != nil {
            return
        }
    }
    return
}

// Start of packet reader/writer functions

// packetIDKeepAlive

func WriteKeepAlive(writer io.Writer) os.Error {
    return binary.Write(writer, binary.BigEndian, byte(packetIDKeepAlive))
}

func ReadKeepAlive(reader io.Reader, handler PacketHandler) (err os.Error) {
    handler.PacketKeepAlive()
    return
}

// packetIDLogin

func CSReadLogin(reader io.Reader) (username, password string, err os.Error) {
    var packetStart struct {
        PacketID byte
        Version  int32
    }

    err = binary.Read(reader, binary.BigEndian, &packetStart)
    if err != nil {
        return
    }
    if packetStart.PacketID != packetIDLogin {
        err = os.NewError(fmt.Sprintf("CSReadLogin: invalid packet ID %#x", packetStart.PacketID))
        return
    }
    if packetStart.Version != protocolVersion {
        err = os.NewError(fmt.Sprintf("CSReadLogin: unsupported protocol version %#x", packetStart.Version))
        return
    }

    username, err = ReadString(reader)
    if err != nil {
        return
    }

    password, err = ReadString(reader)
    if err != nil {
        return
    }

    var packetEnd struct {
        MapSeed   int64
        Dimension byte
    }

    err = binary.Read(reader, binary.BigEndian, &packetEnd)

    return
}

func SCReadLogin(reader io.Reader, handler SCPacketHandler) (err os.Error) {
    var entityID int32

    err = binary.Read(reader, binary.BigEndian, &entityID)
    if err != nil {
        return
    }

    str1, err := ReadString(reader)
    if err != nil {
        return
    }

    str2, err := ReadString(reader)
    if err != nil {
        return
    }

    var packetEnd struct {
        MapSeed   int64
        Dimension byte
    }

    err = binary.Read(reader, binary.BigEndian, &packetEnd)
    if err != nil {
        return
    }

    handler.SCPacketLogin(
        EntityID(entityID),
        str1,
        str2,
        packetEnd.MapSeed,
        packetEnd.Dimension)

    return
}

func SCWriteLogin(writer io.Writer, entityID EntityID) (err os.Error) {
    var packetStart = struct {
        PacketID byte
        EntityID int32
    }{
        packetIDLogin,
        int32(entityID),
    }
    err = binary.Write(writer, binary.BigEndian, &packetStart)
    if err != nil {
        return
    }

    // TODO unknown string
    err = WriteString(writer, "")
    if err != nil {
        return
    }

    // TODO unknown string
    err = WriteString(writer, "")
    if err != nil {
        return
    }

    var packetEnd = struct {
        MapSeed   int64
        Dimension byte
    }{
        // TODO proper map seed as a parameter
        0,
        // TODO proper dimension as a parameter
        0,
    }
    return binary.Write(writer, binary.BigEndian, &packetEnd)
}

// packetIDHandshake

func CSReadHandshake(reader io.Reader) (username string, err os.Error) {
    var packetID byte
    err = binary.Read(reader, binary.BigEndian, &packetID)
    if err != nil {
        return
    }
    if packetID != packetIDHandshake {
        err = os.NewError(fmt.Sprintf("CSReadHandshake: invalid packet ID %#x", packetID))
        return
    }

    return ReadString(reader)
}

func SCReadHandshake(reader io.Reader) (connectionHash string, err os.Error) {
    var packetID byte
    err = binary.Read(reader, binary.BigEndian, &packetID)
    if err != nil {
        return
    }
    if packetID != packetIDHandshake {
        err = os.NewError(fmt.Sprintf("SCReadHandshake: invalid packet ID %#x", packetID))
        return
    }

    return ReadString(reader)
}

func SCWriteHandshake(writer io.Writer, reply string) (err os.Error) {
    err = binary.Write(writer, binary.BigEndian, byte(packetIDHandshake))
    if err != nil {
        return
    }

    return WriteString(writer, reply)
}

// packetIDChatMessage

func WriteChatMessage(writer io.Writer, message string) (err os.Error) {
    err = binary.Write(writer, binary.BigEndian, byte(packetIDChatMessage))
    if err != nil {
        return
    }

    err = WriteString(writer, message)
    return
}

func ReadChatMessage(reader io.Reader, handler PacketHandler) (err os.Error) {
    message, err := ReadString(reader)
    if err != nil {
        return
    }

    // TODO sanitize chat message

    handler.PacketChatMessage(message)
    return
}

// packetIDTimeUpdate

func WriteTimeUpdate(writer io.Writer, time int64) os.Error {
    var packet = struct {
        PacketID byte
        Time     int64
    }{
        packetIDTimeUpdate,
        time,
    }
    return binary.Write(writer, binary.BigEndian, &packet)
}

func SCReadTimeUpdate(reader io.Reader, handler SCPacketHandler) (err os.Error) {
    var time int64

    err = binary.Read(reader, binary.BigEndian, &time)
    if err != nil {
        return
    }

    handler.SCPacketTimeUpdate(time)
    return
}

// packetIDPlayerInventory

func WritePlayerInventory(writer io.Writer) (err os.Error) {
    type InventoryType struct {
        inventoryType int32
        count         int16
        itemID        int16
        // TODO confirm what this field really is
        damage int16
    }
    // TODO pass actual values
    var inventories = []InventoryType{
        InventoryType{inventoryTypeMain, 36, 0, 0},
        InventoryType{inventoryTypeArmor, 4, 0, 0},
        InventoryType{inventoryTypeCrafting, 4, 0, 0},
    }

    for _, inventory := range inventories {
        var packet = struct {
            PacketID      byte
            InventoryType int32
            Count         int16
        }{
            packetIDPlayerInventory,
            inventory.inventoryType,
            inventory.count,
        }
        err = binary.Write(writer, binary.BigEndian, &packet)
        if err != nil {
            return
        }

        for i := int16(0); i < inventory.count; i++ {
            err = binary.Write(writer, binary.BigEndian, int16(-1))
            if err != nil {
                return
            }
        }
    }
    return
}

// packetIDSpawnPosition

func WriteSpawnPosition(writer io.Writer, position *XYZ) os.Error {
    var packet = struct {
        PacketID byte
        X        int32
        Y        int32
        Z        int32
    }{
        packetIDSpawnPosition,
        int32(position.X),
        int32(position.Y),
        int32(position.Z),
    }
    return binary.Write(writer, binary.BigEndian, &packet)
}

func SCReadSpawnPosition(reader io.Reader, handler SCPacketHandler) (err os.Error) {
    var packet struct {
        X   int32
        Y   int32
        Z   int32
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.SCPacketSpawnPosition(&BlockXYZ{
        BlockCoord(packet.X),
        BlockCoord(packet.Y),
        BlockCoord(packet.Z),
    })
    return
}

// packetIDUseEntity

func ReadUseEntity(reader io.Reader, handler PacketHandler) (err os.Error) {
    var packet struct {
        User      EntityID
        Target    EntityID
        LeftClick byte
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    // TODO pass values to handler

    return
}

// packetIDUpdateHealth

func SCReadUpdateHealth(reader io.Reader, handler SCPacketHandler) (err os.Error) {
    var health int16

    err = binary.Read(reader, binary.BigEndian, &health)
    if err != nil {
        return
    }

    handler.SCPacketUpdateHealth(health)
    return
}

// packetIDFlying

func ReadFlying(reader io.Reader, handler PacketHandler) (err os.Error) {
    var packet struct {
        OnGround byte
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.PacketFlying(byteToBool(packet.OnGround))
    return
}

// packetIDPlayerPosition

func WritePlayerPosition(writer io.Writer, position *XYZ, stance float64, onGround bool) os.Error {
    var packet = struct {
        PacketID byte
        X        float64
        Y        float64
        Stance   float64
        Z        float64
        OnGround byte
    }{
        packetIDPlayerPosition,
        float64(position.X),
        float64(position.Y),
        float64(stance),
        float64(position.Z),
        boolToByte(onGround),
    }
    return binary.Write(writer, binary.BigEndian, &packet)
}

func ReadPlayerPosition(reader io.Reader, handler PacketHandler) (err os.Error) {
    var packet struct {
        X      float64
        Y      float64
        Stance float64
        Z      float64
        Flying byte
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.PacketPlayerPosition(&XYZ{
        AbsoluteCoord(packet.X),
        AbsoluteCoord(packet.Y),
        AbsoluteCoord(packet.Z),
    },
        AbsoluteCoord(packet.Stance), byteToBool(packet.Flying))
    return
}

// packetIDPlayerLook

func ReadPlayerLook(reader io.Reader, handler PacketHandler) (err os.Error) {
    var packet struct {
        Yaw      float32
        Pitch    float32
        OnGround byte
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.PacketPlayerLook(&Orientation{
        AngleRadians(packet.Yaw),
        AngleRadians(packet.Pitch),
    },
        byteToBool(packet.OnGround))
    return
}

func SCReadPlayerPositionLook(reader io.Reader, handler SCPacketHandler) (err os.Error) {
    var packet struct {
        X        float64
        Y        float64
        Stance   float64
        Z        float64
        Yaw      float32
        Pitch    float32
        OnGround byte
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.PacketPlayerPosition(&XYZ{
        AbsoluteCoord(packet.X),
        AbsoluteCoord(packet.Y),
        AbsoluteCoord(packet.Z),
    },
        AbsoluteCoord(packet.Stance), byteToBool(packet.OnGround))
    handler.PacketPlayerLook(&Orientation{
        AngleRadians(packet.Yaw),
        AngleRadians(packet.Pitch),
    },
        byteToBool(packet.OnGround))
    return
}

// packetIDPlayerPositionLook

func SCWritePlayerPositionLook(writer io.Writer, position *XYZ, orientation *Orientation, stance AbsoluteCoord, flying bool) os.Error {
    var packet = struct {
        PacketID byte
        X        float64
        Y        float64
        Stance   float64
        Z        float64
        Yaw      float32
        Pitch    float32
        Flying   byte
    }{
        packetIDPlayerPositionLook,
        float64(position.X),
        float64(position.Y),
        float64(stance),
        float64(position.Z),
        float32(orientation.Rotation),
        float32(orientation.Pitch),
        boolToByte(flying),
    }
    return binary.Write(writer, binary.BigEndian, &packet)
}

func CSReadPlayerPositionLook(reader io.Reader, handler CSPacketHandler) (err os.Error) {
    var packet struct {
        X        float64
        Stance   float64
        Y        float64
        Z        float64
        Yaw      float32
        Pitch    float32
        OnGround byte
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.PacketPlayerPosition(&XYZ{
        AbsoluteCoord(packet.X),
        AbsoluteCoord(packet.Y),
        AbsoluteCoord(packet.Z),
    },
        AbsoluteCoord(packet.Stance), byteToBool(packet.OnGround))
    handler.PacketPlayerLook(&Orientation{
        AngleRadians(packet.Yaw),
        AngleRadians(packet.Pitch),
    },
        byteToBool(packet.OnGround))
    return
}

// packetIDPlayerDigging

func ReadPlayerDigging(reader io.Reader, handler PacketHandler) (err os.Error) {
    var packet struct {
        Status byte
        X      int32
        Y      byte
        Z      int32
        Face   byte
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.PacketPlayerDigging(
        DigStatus(packet.Status),
        &BlockXYZ{
            BlockCoord(packet.X),
            BlockCoord(packet.Y),
            BlockCoord(packet.Z),
        },
        Face(packet.Face))
    return
}

// packetIDPlayerBlockPlacement

func ReadPlayerBlockPlacement(reader io.Reader, handler PacketHandler) (err os.Error) {
    var packet struct {
        ID        int16
        X         int32
        Y         byte
        Z         int32
        Direction byte
        ItemID    int16
    }
    var packetExtra struct {
        Amount byte
        Uses   int16
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    if packet.ItemID >= 0 {
        err = binary.Read(reader, binary.BigEndian, &packetExtra)
        if err != nil {
            return
        }
    }

    // TODO pass ItemID, Amount, Uses on to handler
    handler.PacketPlayerBlockPlacement(packet.ID,
        &BlockXYZ{
            BlockCoord(packet.X),
            BlockCoord(packet.Y),
            BlockCoord(packet.Z),
        },
        Face(packet.Direction))
    return
}

// packetIDHoldingChange

func ReadHoldingChange(reader io.Reader, handler PacketHandler) (err os.Error) {
    var packet struct {
        BlockItemID int16
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.PacketHoldingChange(packet.BlockItemID)
    return
}

// packetIDArmAnimation

func ReadArmAnimation(reader io.Reader, handler PacketHandler) (err os.Error) {
    var packet struct {
        EntityID int32
        Forward  byte
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.PacketArmAnimation(byteToBool(packet.Forward))
    return
}

// packetIDNamedEntitySpawn

func WriteNamedEntitySpawn(writer io.Writer, entityID EntityID, name string, position *XYZ, orientation *Orientation, currentItem int16) (err os.Error) {
    var packetStart = struct {
        PacketID byte
        EntityID int32
    }{
        packetIDNamedEntitySpawn,
        int32(entityID),
    }

    err = binary.Write(writer, binary.BigEndian, &packetStart)
    if err != nil {
        return
    }

    err = WriteString(writer, name)
    if err != nil {
        return
    }

    var packetFinish = struct {
        X           int32
        Y           int32
        Z           int32
        Yaw         byte
        Pitch       byte
        CurrentItem int16
    }{
        int32(position.X * PixelsPerBlock),
        int32(position.Y * PixelsPerBlock),
        int32(position.Z * PixelsPerBlock),
        byte(orientation.Rotation),
        byte(orientation.Pitch),
        currentItem,
    }

    err = binary.Write(writer, binary.BigEndian, &packetFinish)
    return
}

// packetIDPickupSpawn

func WritePickupSpawn(writer io.Writer, item *PickupItem) os.Error {
    var packet = struct {
        PacketID byte
        EntityID int32
        ItemID   int16
        Count    byte
        // TODO check this field
        Uses  int16
        X     AbsoluteCoordInteger
        Y     AbsoluteCoordInteger
        Z     AbsoluteCoordInteger
        Yaw   byte
        Pitch byte
        Roll  byte
    }{
        packetIDPickupSpawn,
        int32(item.Entity.EntityID),
        int16(item.itemType),
        byte(item.count),
        // TODO pass proper damage value
        0,
        item.position.X,
        item.position.Y,
        item.position.Z,
        byte(item.orientation.Rotation),
        byte(item.orientation.Pitch),
        byte(item.orientation.Roll),
    }

    return binary.Write(writer, binary.BigEndian, &packet)
}

// packetIDMobSpawn

func SCReadMobSpawn(reader io.Reader, handler SCPacketHandler) (err os.Error) {
    var packet struct {
        EntityID EntityID
        MobType  byte
        X        AbsoluteCoordInteger
        Y        AbsoluteCoordInteger
        Z        AbsoluteCoordInteger
        Yaw      byte
        Pitch    byte
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    err = readUnknownExtra(reader)
    if err != nil {
        return
    }

    handler.SCPacketMobSpawn(
        EntityID(packet.EntityID), packet.MobType,
        &XYZInteger{packet.X, packet.Y, packet.Z},
        packet.Yaw, packet.Pitch)

    return err
}

// packetIDUnknownX19

// TODO determine what this packet is
func ReadUnknownX19(reader io.Reader, handler PacketHandler) (err os.Error) {
    var field1 int32
    err = binary.Read(reader, binary.BigEndian, &field1)
    if err != nil {
        return
    }

    _, err = ReadString(reader)
    if err != nil {
        return
    }

    var packetEnd struct {
        field3, field4, field5, field6 int32
    }

    err = binary.Read(reader, binary.BigEndian, &packetEnd)
    if err != nil {
        return
    }

    // TODO pass this data to handler

    return
}

// packetIDEntityVelocity

func ReadEntityVelocity(reader io.Reader, handler PacketHandler) (err os.Error) {
    var packet struct {
        EntityID EntityID
        X, Y, Z  int16
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    // TODO pass this data to handler

    return
}

// packetIDEntity

func SCReadEntity(reader io.Reader, handler SCPacketHandler) (err os.Error) {
    var entityID EntityID

    err = binary.Read(reader, binary.BigEndian, &entityID)
    if err != nil {
        return
    }

    // TODO pass this data to handler

    return
}

// packetIDEntityDestroy

func WriteEntityDestroy(writer io.Writer, entityID EntityID) os.Error {
    var packet = struct {
        PacketID byte
        EntityID int32
    }{
        packetIDEntityDestroy,
        int32(entityID),
    }
    return binary.Write(writer, binary.BigEndian, &packet)
}

func SCReadEntityDestroy(reader io.Reader, handler SCPacketHandler) (err os.Error) {
    var entityID EntityID

    err = binary.Read(reader, binary.BigEndian, &entityID)
    if err != nil {
        return
    }

    // TODO pass this data to handler

    return
}

// packetIDEntityRelMove

func SCReadEntityRelMove(reader io.Reader, handler SCPacketHandler) (err os.Error) {
    var packet struct {
        EntityID EntityID
        X, Y, Z  byte
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    // TODO pass this data to handler

    return
}

// packetIDEntityLook

func WriteEntityLook(writer io.Writer, entityID EntityID, orientation *Orientation) os.Error {
    var packet = struct {
        PacketID byte
        EntityID int32
        Yaw      byte
        Pitch    byte
    }{
        packetIDEntityLook,
        int32(entityID),
        byte(orientation.Rotation * 256 / 360),
        byte(orientation.Pitch * 64 / 90),
    }
    return binary.Write(writer, binary.BigEndian, &packet)
}

func SCReadEntityLook(reader io.Reader, handler SCPacketHandler) (err os.Error) {
    var packet struct {
        EntityID int32
        Yaw      byte
        Pitch    byte
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    // TODO pass values to handler

    return
}

// packetIDEntityLookAndRelMove

func SCReadEntityLookAndRelMove(reader io.Reader, handler SCPacketHandler) (err os.Error) {
    var packet struct {
        EntityID int32
        X, Y, Z  byte
        Yaw      byte
        Pitch    byte
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    // TODO pass values to handler

    return
}

// packetIDEntityTeleport

func WriteEntityTeleport(writer io.Writer, entityID EntityID, position *XYZ, orientation *Orientation) os.Error {
    var packet = struct {
        PacketID byte
        EntityID int32
        X        int32
        Y        int32
        Z        int32
        Yaw      byte
        Pitch    byte
    }{
        packetIDEntityTeleport,
        int32(entityID),
        int32(position.X * PixelsPerBlock),
        int32(position.Y * PixelsPerBlock),
        int32(position.Z * PixelsPerBlock),
        byte(orientation.Rotation * 256 / 360),
        byte(orientation.Pitch * 64 / 90),
    }
    return binary.Write(writer, binary.BigEndian, &packet)
}

// packetIDEntityStatus

func SCReadEntityStatus(reader io.Reader, handler SCPacketHandler) (err os.Error) {
    var packet struct {
        EntityID EntityID
        Status   byte
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    // TODO pass values to handler

    return
}

// packetIDUnknownX28

func ReadUnknownX28(reader io.Reader, handler PacketHandler) (err os.Error) {
    var field1 int32

    err = binary.Read(reader, binary.BigEndian, &field1)
    if err != nil {
        return
    }

    err = readUnknownExtra(reader)
    if err != nil {
        return
    }

    // TODO pass values to handler

    return
}

// packetIDPreChunk

func WritePreChunk(writer io.Writer, chunkLoc *ChunkXZ, willSend bool) os.Error {
    var packet = struct {
        PacketID byte
        X        int32
        Z        int32
        WillSend byte
    }{
        packetIDPreChunk,
        int32(chunkLoc.X),
        int32(chunkLoc.Z),
        boolToByte(willSend),
    }
    return binary.Write(writer, binary.BigEndian, &packet)
}

func SCReadPreChunk(reader io.Reader, handler SCPacketHandler) (err os.Error) {
    var packet struct {
        X    ChunkCoord
        Z    ChunkCoord
        Mode byte
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.SCPacketPreChunk(&ChunkXZ{packet.X, packet.Z}, packet.Mode != 0)

    return
}

// packetIDMapChunk

func WriteMapChunk(writer io.Writer, chunkLoc *ChunkXZ, blocks, blockData, blockLight, skyLight []byte) (err os.Error) {
    buf := &bytes.Buffer{}
    compressed, err := zlib.NewWriter(buf)
    if err != nil {
        return
    }

    compressed.Write(blocks)
    compressed.Write(blockData)
    compressed.Write(blockLight)
    compressed.Write(skyLight)
    compressed.Close()
    bs := buf.Bytes()

    var packet = struct {
        PacketID         byte
        X                int32
        Y                int16
        Z                int32
        SizeX            byte
        SizeY            byte
        SizeZ            byte
        CompressedLength int32
    }{
        packetIDMapChunk,
        int32(chunkLoc.X * ChunkSizeX),
        0,
        int32(chunkLoc.Z * ChunkSizeZ),
        ChunkSizeX - 1,
        ChunkSizeY - 1,
        ChunkSizeZ - 1,
        int32(len(bs)),
    }

    err = binary.Write(writer, binary.BigEndian, &packet)
    if err != nil {
        return
    }
    err = binary.Write(writer, binary.BigEndian, bs)
    return
}

func SCReadMapChunk(reader io.Reader, handler SCPacketHandler) (err os.Error) {
    var packet struct {
        X                BlockCoord
        Y                int16
        Z                BlockCoord
        SizeX            byte
        SizeY            byte
        SizeZ            byte
        CompressedLength int32
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    data := make([]byte, packet.CompressedLength)
    _, err = io.ReadFull(reader, data)
    if err != nil {
        return
    }

    handler.SCPacketMapChunk(
        &BlockXYZ{packet.X, BlockCoord(packet.Y), packet.Z},
        packet.SizeX, packet.SizeY, packet.SizeZ,
        data)
    return
}

// packetIDBlockChangeMulti

func SCReadBlockChangeMulti(reader io.Reader, handler SCPacketHandler) (err os.Error) {
    var packet struct {
        ChunkX int32
        ChunkZ int32
        Count  int16
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    coordArray := make([]int16, packet.Count)
    blockTypeArray := make([]byte, packet.Count)
    blockMetadataArray := make([]byte, packet.Count)

    err = binary.Read(reader, binary.BigEndian, coordArray)
    err = binary.Read(reader, binary.BigEndian, blockTypeArray)
    err = binary.Read(reader, binary.BigEndian, blockMetadataArray)

    // TODO pass values to handler

    return
}

// packetIDBlockChange

func WriteBlockChange(writer io.Writer, blockLoc *BlockXYZ, blockType BlockID, blockMetaData byte) (err os.Error) {
    var packet = struct {
        PacketID      byte
        X             int32
        Y             byte
        Z             int32
        BlockType     byte
        BlockMetadata byte
    }{
        packetIDBlockChange,
        int32(blockLoc.X),
        byte(blockLoc.Y),
        int32(blockLoc.Z),
        byte(blockType),
        byte(blockMetaData),
    }
    err = binary.Write(writer, binary.BigEndian, &packet)
    return
}

func SCReadBlockChange(reader io.Reader, handler SCPacketHandler) (err os.Error) {
    var packet struct {
        X             int32
        Y             byte
        Z             int32
        BlockType     byte
        BlockMetadata byte
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    // TODO pass values to handler

    return
}

// packetIDUnknownX36

func ReadUnknownX36(reader io.Reader, handler PacketHandler) (err os.Error) {
    var packet struct {
        field1         int32
        field2         int16
        field3         int32
        field4, field5 byte
    }

    err = binary.Read(reader, binary.BigEndian, &packet)

    if err != nil {
        return
    }

    // TODO pass values to handler

    return
}

// packetIDWindowClick

func CSReadWindowClick(reader io.Reader, handler CSPacketHandler) (err os.Error) {
    var packetStart struct {
        WindowId     byte
        Slot         int16
        RightClick   byte
        ActionNumber int16
        ItemID       int16
    }

    err = binary.Read(reader, binary.BigEndian, &packetStart)
    if err != nil {
        return
    }

    if packetStart.ItemID != -1 {
        var packetEnd struct {
            Amount byte
            Uses   int16
        }
        err = binary.Read(reader, binary.BigEndian, &packetEnd)
        if err != nil {
            return
        }
    }

    // TODO pass values to handler

    return
}

// packetIDSetSlot

func SCReadSetSlot(reader io.Reader, handler SCPacketHandler) (err os.Error) {
    var packetStart struct {
        WindowId byte
        Slot     int16
        ItemID   int16
    }

    err = binary.Read(reader, binary.BigEndian, &packetStart)
    if err != nil {
        return
    }

    if packetStart.ItemID != -1 {
        var packetEnd struct {
            Amount byte
            Uses   int16
        }

        err = binary.Read(reader, binary.BigEndian, &packetEnd)
        if err != nil {
            return
        }
    }

    // TODO pass values to handler

    return
}

// packetIDWindowItems

func SCReadWindowItems(reader io.Reader, handler SCPacketHandler) (err os.Error) {
    var packetStart struct {
        WindowId byte
        Count    int16
    }

    err = binary.Read(reader, binary.BigEndian, &packetStart)
    if err != nil {
        return
    }

    var itemID int16
    var itemInfo struct {
        Amount byte
        Uses   int16
    }

    // TODO collect inventory data for handler, data is currently discarded
    for i := int16(0); i < packetStart.Count; i++ {
        err = binary.Read(reader, binary.BigEndian, &itemID)
        if err != nil {
            return
        }

        if itemID != -1 {
            err = binary.Read(reader, binary.BigEndian, &itemInfo)
            if err != nil {
                return
            }
        }
    }

    // TODO pass values to handler

    return
}

// packetIDDisconnect

func ReadDisconnect(reader io.Reader, handler PacketHandler) (err os.Error) {
    reason, err := ReadString(reader)
    if err != nil {
        return
    }

    handler.PacketDisconnect(reason)
    return
}

func WriteDisconnect(writer io.Writer, reason string) (err os.Error) {
    buf := &bytes.Buffer{}
    binary.Write(buf, binary.BigEndian, byte(packetIDDisconnect))
    WriteString(buf, reason)
    _, err = writer.Write(buf.Bytes())
    return
}


// End of packet reader/writer functions


type commonPacketHandler func(io.Reader, PacketHandler) os.Error
type csPacketHandler func(io.Reader, CSPacketHandler) os.Error
type scPacketHandler func(io.Reader, SCPacketHandler) os.Error

type commonPacketReaderMap map[byte]commonPacketHandler
type csPacketReaderMap map[byte]csPacketHandler
type scPacketReaderMap map[byte]scPacketHandler

// Common packet mapping
var commonReadFns = commonPacketReaderMap{
    packetIDKeepAlive:            ReadKeepAlive,
    packetIDChatMessage:          ReadChatMessage,
    packetIDUseEntity:            ReadUseEntity,
    packetIDFlying:               ReadFlying,
    packetIDPlayerPosition:       ReadPlayerPosition,
    packetIDPlayerLook:           ReadPlayerLook,
    packetIDPlayerDigging:        ReadPlayerDigging,
    packetIDPlayerBlockPlacement: ReadPlayerBlockPlacement,
    packetIDHoldingChange:        ReadHoldingChange,
    packetIDArmAnimation:         ReadArmAnimation,
    packetIDUnknownX19:           ReadUnknownX19,
    packetIDEntityVelocity:       ReadEntityVelocity,
    packetIDUnknownX28:           ReadUnknownX28,
    packetIDUnknownX36:           ReadUnknownX36,
    packetIDDisconnect:           ReadDisconnect,
}

// Client->server specific packet mapping
var csReadFns = csPacketReaderMap{
    packetIDPlayerPositionLook: CSReadPlayerPositionLook,
    packetIDWindowClick:        CSReadWindowClick,
}

// Server->client specific packet mapping
var scReadFns = scPacketReaderMap{
    packetIDLogin:                SCReadLogin,
    packetIDTimeUpdate:           SCReadTimeUpdate,
    packetIDSpawnPosition:        SCReadSpawnPosition,
    packetIDUpdateHealth:         SCReadUpdateHealth,
    packetIDPlayerPositionLook:   SCReadPlayerPositionLook,
    packetIDMobSpawn:             SCReadMobSpawn,
    packetIDEntityDestroy:        SCReadEntityDestroy,
    packetIDEntity:               SCReadEntity,
    packetIDEntityRelMove:        SCReadEntityRelMove,
    packetIDEntityLook:           SCReadEntityLook,
    packetIDEntityLookAndRelMove: SCReadEntityLookAndRelMove,
    packetIDEntityStatus:         SCReadEntityStatus,
    packetIDPreChunk:             SCReadPreChunk,
    packetIDMapChunk:             SCReadMapChunk,
    packetIDBlockChangeMulti:     SCReadBlockChangeMulti,
    packetIDBlockChange:          SCReadBlockChange,
    packetIDSetSlot:              SCReadSetSlot,
    packetIDWindowItems:          SCReadWindowItems,
}

func CSReadPacket(reader io.Reader, handler CSPacketHandler) os.Error {
    var packetID byte

    if err := binary.Read(reader, binary.BigEndian, &packetID); err != nil {
        return err
    }

    if commonFn, ok := commonReadFns[packetID]; ok {
        return commonFn(reader, handler)
    }

    if csFn, ok := csReadFns[packetID]; ok {
        return csFn(reader, handler)
    }

    return os.NewError(fmt.Sprintf("unhandled packet type %#x", packetID))
}

func SCReadPacket(reader io.Reader, handler SCPacketHandler) os.Error {
    var packetID byte

    if err := binary.Read(reader, binary.BigEndian, &packetID); err != nil {
        return err
    }

    if commonFn, ok := commonReadFns[packetID]; ok {
        return commonFn(reader, handler)
    }

    if scFn, ok := scReadFns[packetID]; ok {
        return scFn(reader, handler)
    }

    return os.NewError(fmt.Sprintf("unhandled packet type %#x", packetID))
}

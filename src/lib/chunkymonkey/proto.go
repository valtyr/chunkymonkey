package proto

import (
    "io"
    "os"
    "fmt"
    "bytes"
    "encoding/binary"
    "compress/zlib"

    .   "chunkymonkey/types"
)

const (
    // Currently only this protocol version is supported
    protocolVersion = 9

    // Packet type IDs
    packetIDKeepAlive            = 0x00
    packetIDLogin                = 0x01
    packetIDHandshake            = 0x02
    packetIDChatMessage          = 0x03
    packetIDTimeUpdate           = 0x04
    packetIDEntityEquipment      = 0x05
    packetIDSpawnPosition        = 0x06
    packetIDUseEntity            = 0x07
    packetIDUpdateHealth         = 0x08
    packetIDRespawn              = 0x09
    packetIDPlayer               = 0x0a
    packetIDPlayerPosition       = 0x0b
    packetIDPlayerLook           = 0x0c
    packetIDPlayerPositionLook   = 0x0d
    packetIDPlayerDigging        = 0x0e
    packetIDPlayerBlockPlacement = 0x0f
    packetIDHoldingChange        = 0x10
    packetIDUseBed               = 0x11
    packetIDEntityAnimation      = 0x12
    packetIDEntityAction         = 0x13
    packetIDNamedEntitySpawn     = 0x14
    packetIDItemSpawn            = 0x15
    packetIDItemCollect          = 0x16
    packetIDObjectSpawn          = 0x17
    packetIDEntitySpawn          = 0x18
    packetIDPaintingSpawn        = 0x19
    packetIDUnknown0x1b          = 0x1b
    packetIDEntityVelocity       = 0x1c
    packetIDEntityDestroy        = 0x1d
    packetIDEntity               = 0x1e
    packetIDEntityRelMove        = 0x1f
    packetIDEntityLook           = 0x20
    packetIDEntityLookAndRelMove = 0x21
    packetIDEntityTeleport       = 0x22
    packetIDEntityStatus         = 0x26
    packetIDEntityMetadata       = 0x28
    packetIDPreChunk             = 0x32
    packetIDMapChunk             = 0x33
    packetIDBlockChangeMulti     = 0x34
    packetIDBlockChange          = 0x35
    packetIDNoteBlockPlay        = 0x36
    packetIDExplosion            = 0x3c
    packetIDWindowOpen           = 0x64
    packetIDWindowClose          = 0x65
    packetIDWindowClick          = 0x66
    packetIDWindowSetSlot        = 0x67
    packetIDWindowItems          = 0x68
    packetIDWindowProgressBar    = 0x69
    packetIDWindowTransaction    = 0x6a
    packetIDSignUpdate           = 0x82
    packetIDDisconnect           = 0xff

    // Inventory types
    // FIXME remove these with the WritePlayerInventory
    inventoryTypeMain     = -1
    inventoryTypeArmor    = -2
    inventoryTypeCrafting = -3
)

// Packets commonly received by both client and server
type PacketHandler interface {
    PacketKeepAlive()
    PacketChatMessage(message string)
    PacketEntityAction(entityID EntityID, action EntityAction)
    PacketUseEntity(user EntityID, target EntityID, leftClick bool)
    PacketRespawn()
    PacketPlayerPosition(position *AbsXYZ, stance AbsCoord, onGround bool)
    PacketPlayerLook(look *LookDegrees, onGround bool)
    PacketPlayerDigging(status DigStatus, blockLoc *BlockXYZ, face Face)
    PacketPlayerBlockPlacement(itemID ItemID, blockLoc *BlockXYZ, face Face, amount ItemCount, uses ItemUses)
    PacketEntityAnimation(entityID EntityID, animation EntityAnimation)
    PacketUnknown0x1b(field1, field2, field3, field4 float32, field5, field6 bool)
    PacketSignUpdate(position *BlockXYZ, lines [4]string)
    PacketDisconnect(reason string)
}

// Servers to the protocol must implement this interface to receive packets
type ServerPacketHandler interface {
    PacketHandler
    PacketPlayer(onGround bool)
    PacketHoldingChange(itemID ItemID)
    PacketWindowClose(windowID WindowID)
    PacketWindowClick(windowID WindowID, slot SlotID, rightClick bool, txID TxID, itemID ItemID, amount ItemCount, uses ItemUses)
}

// Clients to the protocol must implement this interface to receive packets
type ClientPacketHandler interface {
    PacketHandler
    ClientPacketLogin(entityID EntityID, mapSeed RandomSeed, dimension DimensionID)
    PacketTimeUpdate(time TimeOfDay)
    PacketUseBed(flag bool, bedLoc *BlockXYZ)
    PacketNamedEntitySpawn(entityID EntityID, name string, position *AbsIntXYZ, look *LookBytes, currentItem ItemID)
    PacketEntityEquipment(entityID EntityID, slot SlotID, itemID ItemID, uses ItemUses)
    PacketSpawnPosition(position *BlockXYZ)
    PacketUpdateHealth(health int16)
    PacketItemSpawn(entityID EntityID, itemID ItemID, count ItemCount, uses ItemUses, location *AbsIntXYZ, orientation *OrientationBytes)
    PacketItemCollect(collectedItem EntityID, collector EntityID)
    PacketObjectSpawn(entityID EntityID, objType ObjTypeID, position *AbsIntXYZ)
    PacketEntitySpawn(entityID EntityID, mobType EntityMobType, position *AbsIntXYZ, look *LookBytes, data []EntityMetadata)
    PacketPaintingSpawn(entityID EntityID, title string, position *BlockXYZ, paintingType PaintingTypeID)
    PacketEntityVelocity(entityID EntityID, velocity *Velocity)
    PacketEntityDestroy(entityID EntityID)
    PacketEntity(entityID EntityID)
    PacketEntityRelMove(entityID EntityID, movement *RelMove)
    PacketEntityLook(entityID EntityID, look *LookBytes)
    PacketEntityTeleport(entityID EntityID, position *AbsIntXYZ, look *LookBytes)
    PacketEntityStatus(entityID EntityID, status EntityStatus)
    PacketEntityMetadata(entityID EntityID, metadata []EntityMetadata)

    PacketPreChunk(position *ChunkXZ, mode ChunkLoadMode)
    PacketMapChunk(position *BlockXYZ, size *SubChunkSize, data []byte)
    PacketBlockChangeMulti(chunkLoc *ChunkXZ, blockCoords []SubChunkXYZ, blockTypes []BlockID, blockMetaData []byte)
    PacketBlockChange(blockLoc *BlockXYZ, blockType BlockID, blockMetaData byte)
    PacketNoteBlockPlay(position *BlockXYZ, instrument InstrumentID, pitch NotePitch)

    // NOTE method signature likely to change
    PacketExplosion(position *AbsXYZ, power float32, blockOffsets []ExplosionOffsetXYZ)

    PacketWindowOpen(windowID WindowID, invTypeID InvTypeID, windowTitle string, numSlots byte)
    PacketWindowSetSlot(windowID WindowID, slot SlotID, itemID ItemID, amount ItemCount, uses ItemUses)
    PacketWindowItems(windowID WindowID, items []WindowSlot)
    PacketWindowProgressBar(windowID WindowID, prgBarID PrgBarID, value PrgBarValue)
    PacketWindowTransaction(windowID WindowID, txID TxID, accepted bool)
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

func readString(reader io.Reader) (s string, err os.Error) {
    var length int16
    err = binary.Read(reader, binary.BigEndian, &length)
    if err != nil {
        return
    }

    bs := make([]byte, uint16(length))
    _, err = io.ReadFull(reader, bs)
    return string(bs), err
}

func writeString(writer io.Writer, s string) (err os.Error) {
    bs := []byte(s)

    err = binary.Write(writer, binary.BigEndian, int16(len(bs)))
    if err != nil {
        return
    }

    _, err = writer.Write(bs)
    return
}

type WindowSlot struct {
    ItemID ItemID
    Amount ItemCount
    Uses   ItemUses
}

type EntityMetadata struct {
    Field1 byte
    Field2 byte
    Field3 interface{}
}

func writeEntityMetadataField(writer io.Writer, data []EntityMetadata) (err os.Error) {
    // NOTE that no checking is done upon the form of the data, so it's
    // possible to form bad data packets with this.
    var entryType byte

    for _, item := range data {
        entryType = (item.Field1 << 5) & 0xe0
        entryType |= (item.Field2 & 0x1f)

        if err = binary.Write(writer, binary.BigEndian, entryType); err != nil {
            return
        }

        if err = binary.Write(writer, binary.BigEndian, &item.Field3); err != nil {
            return
        }
    }

    // Mark end of metadata
    return binary.Write(writer, binary.BigEndian, byte(127))
}

// Reads entity metadata from the end of certain packets. Most of the meaning
// of the packets isn't yet known.
// TODO update to pull useful data out as it becomes understood
func readEntityMetadataField(reader io.Reader) (data []EntityMetadata, err os.Error) {
    var entryType byte

    var field1, field2 byte
    var field3 interface{}

    for {
        err = binary.Read(reader, binary.BigEndian, &entryType)
        if err != nil {
            return
        }
        if entryType == 127 {
            break
        }
        field2 = entryType & 0x1f

        switch field1 := (entryType & 0xe0) >> 5; field1 {
        case 0:
            var byteVal byte
            err = binary.Read(reader, binary.BigEndian, &byteVal)
            field3 = byteVal
        case 1:
            var int16Val int16
            err = binary.Read(reader, binary.BigEndian, &int16Val)
            field3 = int16Val
        case 2:
            var int32Val int32
            err = binary.Read(reader, binary.BigEndian, &int32Val)
            field3 = int32Val
        case 3:
            var floatVal float32
            err = binary.Read(reader, binary.BigEndian, &floatVal)
            field3 = floatVal
        case 4:
            var stringVal string
            stringVal, err = readString(reader)
            field3 = stringVal
        case 5:
            var position struct {
                X   int16
                Y   byte
                Z   int16
            }
            err = binary.Read(reader, binary.BigEndian, &position)
            field3 = position
        }

        data = append(data, EntityMetadata{field1, field2, field3})

        if err != nil {
            return
        }
    }
    return
}

// Start of packet reader/writer functions

// Naming convention:
// * Client* functions are specific to use by clients writing to a server, and
//   reading from it.
// * Server* functions are specific to use by servers writing to clients, and
//   reading from them.
// * Those without a client or server prefix don't differ in content or meaning
//   between client and server.


// packetIDKeepAlive

func WriteKeepAlive(writer io.Writer) os.Error {
    return binary.Write(writer, binary.BigEndian, byte(packetIDKeepAlive))
}

func readKeepAlive(reader io.Reader, handler PacketHandler) (err os.Error) {
    handler.PacketKeepAlive()
    return
}

// packetIDLogin

func commonReadLogin(reader io.Reader) (versionOrEntityID int32, str1, str2 string, mapSeed RandomSeed, dimension DimensionID, err os.Error) {
    if err = binary.Read(reader, binary.BigEndian, &versionOrEntityID); err != nil {
        return
    }
    if str1, err = readString(reader); err != nil {
        return
    }
    if str2, err = readString(reader); err != nil {
        return
    }

    var packetEnd struct {
        MapSeed   RandomSeed
        Dimension DimensionID
    }
    if err = binary.Read(reader, binary.BigEndian, &packetEnd); err != nil {
        return
    }

    mapSeed = packetEnd.MapSeed
    dimension = packetEnd.Dimension

    return
}

func ServerReadLogin(reader io.Reader) (username, password string, err os.Error) {
    var packetID byte
    if err = binary.Read(reader, binary.BigEndian, &packetID); err != nil {
        return
    }
    if packetID != packetIDLogin {
        err = os.NewError(fmt.Sprintf("serverLogin: invalid packet ID %#x", packetID))
        return
    }

    version, username, password, _, _, err := commonReadLogin(reader)
    if err != nil {
        return
    }

    if version != protocolVersion {
        err = os.NewError(fmt.Sprintf("serverLogin: unsupported protocol version %#x", version))
        return
    }
    return
}

func clientReadLogin(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    entityIDInt32, _, _, mapSeed, dimension, err := commonReadLogin(reader)
    if err != nil {
        return
    }

    handler.ClientPacketLogin(EntityID(entityIDInt32), mapSeed, dimension)

    return
}

func commonWriteLogin(writer io.Writer, str1, str2 string, entityID EntityID, mapSeed RandomSeed, dimension DimensionID) (err os.Error) {
    if err = binary.Write(writer, binary.BigEndian, entityID); err != nil {
        return
    }

    // These strings are currently unused
    if err = writeString(writer, str1); err != nil {
        return
    }
    if err = writeString(writer, str2); err != nil {
        return
    }

    var packetEnd = struct {
        MapSeed   RandomSeed
        Dimension DimensionID
    }{
        mapSeed,
        dimension,
    }
    return binary.Write(writer, binary.BigEndian, &packetEnd)
}

func ServerWriteLogin(writer io.Writer, entityID EntityID, mapSeed RandomSeed, dimension DimensionID) (err os.Error) {
    if err = binary.Write(writer, binary.BigEndian, byte(packetIDLogin)); err != nil {
        return
    }

    return commonWriteLogin(writer, "", "", entityID, mapSeed, dimension)
}

func ClientWriteLogin(writer io.Writer, username, password string) (err os.Error) {
    if err = binary.Write(writer, binary.BigEndian, byte(packetIDLogin)); err != nil {
        return
    }

    return commonWriteLogin(writer, username, password, protocolVersion, 0, 0)
}

// packetIDHandshake

func ServerReadHandshake(reader io.Reader) (username string, err os.Error) {
    var packetID byte
    err = binary.Read(reader, binary.BigEndian, &packetID)
    if err != nil {
        return
    }
    if packetID != packetIDHandshake {
        err = os.NewError(fmt.Sprintf("serverHandshake: invalid packet ID %#x", packetID))
        return
    }

    return readString(reader)
}

func ClientReadHandshake(reader io.Reader) (serverId string, err os.Error) {
    var packetID byte
    err = binary.Read(reader, binary.BigEndian, &packetID)
    if err != nil {
        return
    }
    if packetID != packetIDHandshake {
        err = os.NewError(fmt.Sprintf("readHandshake: invalid packet ID %#x", packetID))
        return
    }

    return readString(reader)
}

func ServerWriteHandshake(writer io.Writer, reply string) (err os.Error) {
    err = binary.Write(writer, binary.BigEndian, byte(packetIDHandshake))
    if err != nil {
        return
    }

    return writeString(writer, reply)
}

// packetIDChatMessage

func WriteChatMessage(writer io.Writer, message string) (err os.Error) {
    err = binary.Write(writer, binary.BigEndian, byte(packetIDChatMessage))
    if err != nil {
        return
    }

    err = writeString(writer, message)
    return
}

func readChatMessage(reader io.Reader, handler PacketHandler) (err os.Error) {
    message, err := readString(reader)
    if err != nil {
        return
    }

    // TODO sanitize chat message

    handler.PacketChatMessage(message)
    return
}

// packetIDTimeUpdate

func ServerWriteTimeUpdate(writer io.Writer, time TimeOfDay) os.Error {
    var packet = struct {
        PacketID byte
        Time     TimeOfDay
    }{
        packetIDTimeUpdate,
        time,
    }
    return binary.Write(writer, binary.BigEndian, &packet)
}

func readTimeUpdate(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var time TimeOfDay

    err = binary.Read(reader, binary.BigEndian, &time)
    if err != nil {
        return
    }

    handler.PacketTimeUpdate(time)
    return
}

// packetIDEntityEquipment

func ServerWriteEntityEquipment(writer io.Writer, entityID EntityID, slot SlotID, itemID ItemID, uses ItemUses) (err os.Error) {
    var packet = struct {
        PacketID byte
        EntityID EntityID
        Slot     SlotID
        ItemID   ItemID
        Uses     ItemUses
    }{
        packetIDEntityEquipment,
        entityID,
        slot,
        itemID,
        uses,
    }

    return binary.Write(writer, binary.BigEndian, &packet)
}

func readEntityEquipment(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packet struct {
        EntityID EntityID
        Slot     SlotID
        ItemID   ItemID
        Uses     ItemUses
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.PacketEntityEquipment(
        packet.EntityID, packet.Slot, packet.ItemID, packet.Uses)

    return
}

// packetIDSpawnPosition

func WriteSpawnPosition(writer io.Writer, position *BlockXYZ) os.Error {
    var packet = struct {
        PacketID byte
        X        BlockCoord
        Y        int32
        Z        BlockCoord
    }{
        packetIDSpawnPosition,
        position.X,
        int32(position.Y),
        position.Z,
    }
    return binary.Write(writer, binary.BigEndian, &packet)
}

func readSpawnPosition(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packet struct {
        X   BlockCoord
        Y   int32
        Z   BlockCoord
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.PacketSpawnPosition(&BlockXYZ{
        packet.X,
        BlockYCoord(packet.Y),
        packet.Z,
    })
    return
}

// packetIDUseEntity

func WriteUseEntity(writer io.Writer, user EntityID, target EntityID, leftClick bool) (err os.Error) {
    var packet = struct {
        PacketID  byte
        User      EntityID
        Target    EntityID
        LeftClick byte
    }{
        packetIDUseEntity,
        user,
        target,
        boolToByte(leftClick),
    }

    return binary.Write(writer, binary.BigEndian, &packet)
}

func readUseEntity(reader io.Reader, handler PacketHandler) (err os.Error) {
    var packet struct {
        User      EntityID
        Target    EntityID
        LeftClick byte
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.PacketUseEntity(packet.User, packet.Target, byteToBool(packet.LeftClick))

    return
}

// packetIDUpdateHealth

func WriteUpdateHealth(writer io.Writer, health int16) (err os.Error) {
    var packet = struct {
        PacketID byte
        health   int16
    }{
        packetIDUpdateHealth,
        health,
    }

    return binary.Write(writer, binary.BigEndian, &packet)
}

func readUpdateHealth(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var health int16

    err = binary.Read(reader, binary.BigEndian, &health)
    if err != nil {
        return
    }

    handler.PacketUpdateHealth(health)
    return
}

// packetIDRespawn

func WriteRespawn(writer io.Writer) os.Error {
    var packetID byte

    return binary.Write(writer, binary.BigEndian, &packetID)
}

func readRespawn(reader io.Reader, handler PacketHandler) (err os.Error) {
    handler.PacketRespawn()

    return
}

// packetIDPlayer

func WritePlayer(writer io.Writer, onGround bool) (err os.Error) {
    var packet = struct {
        PacketID byte
        OnGround byte
    }{
        packetIDPlayer,
        boolToByte(onGround),
    }

    return binary.Write(writer, binary.BigEndian, &packet)
}

func readPlayer(reader io.Reader, handler ServerPacketHandler) (err os.Error) {
    var onGround byte

    if err = binary.Read(reader, binary.BigEndian, &onGround); err != nil {
        return
    }

    handler.PacketPlayer(byteToBool(onGround))

    return
}

// packetIDPlayerPosition

func WritePlayerPosition(writer io.Writer, position *AbsXYZ, stance AbsCoord, onGround bool) os.Error {
    var packet = struct {
        PacketID byte
        X        AbsCoord
        Y        AbsCoord
        Stance   AbsCoord
        Z        AbsCoord
        OnGround byte
    }{
        packetIDPlayerPosition,
        position.X,
        position.Y,
        stance,
        position.Z,
        boolToByte(onGround),
    }
    return binary.Write(writer, binary.BigEndian, &packet)
}

func readPlayerPosition(reader io.Reader, handler PacketHandler) (err os.Error) {
    var packet struct {
        X        AbsCoord
        Y        AbsCoord
        Stance   AbsCoord
        Z        AbsCoord
        OnGround byte
    }

    if err = binary.Read(reader, binary.BigEndian, &packet); err != nil {
        return
    }

    handler.PacketPlayerPosition(
        &AbsXYZ{
            AbsCoord(packet.X),
            AbsCoord(packet.Y),
            AbsCoord(packet.Z),
        },
        packet.Stance,
        byteToBool(packet.OnGround))
    return
}

// packetIDPlayerLook

func WritePlayerLook(writer io.Writer, look *LookDegrees, onGround bool) (err os.Error) {
    var packet = struct {
        PacketID byte
        Yaw      AngleDegrees
        Pitch    AngleDegrees
        OnGround byte
    }{
        packetIDPlayerLook,
        look.Yaw, look.Pitch,
        boolToByte(onGround),
    }

    return binary.Write(writer, binary.BigEndian, &packet)
}

func readPlayerLook(reader io.Reader, handler PacketHandler) (err os.Error) {
    var packet struct {
        Yaw      AngleDegrees
        Pitch    AngleDegrees
        OnGround byte
    }

    if err = binary.Read(reader, binary.BigEndian, &packet); err != nil {
        return
    }

    handler.PacketPlayerLook(
        &LookDegrees{
            packet.Yaw,
            packet.Pitch,
        },
        byteToBool(packet.OnGround))
    return
}

// packetIDPlayerPositionLook

func WritePlayerPositionLook(writer io.Writer, position *AbsXYZ, stance AbsCoord, look *LookDegrees, onGround bool) (err os.Error) {
    var packet = struct {
        PacketID byte
        X        AbsCoord
        Y        AbsCoord
        Stance   AbsCoord
        Z        AbsCoord
        Yaw      AngleDegrees
        Pitch    AngleDegrees
        OnGround byte
    }{
        packetIDPlayerPositionLook,
        position.X, position.Y, stance, position.Z,
        look.Yaw, look.Pitch,
        boolToByte(onGround),
    }

    return binary.Write(writer, binary.BigEndian, &packet)
}

func readPlayerPositionLook(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packet struct {
        X        AbsCoord
        Y        AbsCoord
        Stance   AbsCoord
        Z        AbsCoord
        Yaw      AngleDegrees
        Pitch    AngleDegrees
        OnGround byte
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.PacketPlayerPosition(
        &AbsXYZ{
            packet.X,
            packet.Y,
            packet.Z,
        },
        packet.Stance,
        byteToBool(packet.OnGround))

    handler.PacketPlayerLook(
        &LookDegrees{
            packet.Yaw,
            packet.Pitch,
        },
        byteToBool(packet.OnGround))
    return
}

// packetIDPlayerPositionLook

// TODO client versions, factor out common code

func ServerWritePlayerPositionLook(writer io.Writer, position *AbsXYZ, look *LookDegrees, stance AbsCoord, onGround bool) os.Error {
    var packet = struct {
        PacketID byte
        X        AbsCoord
        Y        AbsCoord
        Stance   AbsCoord
        Z        AbsCoord
        Yaw      AngleDegrees
        Pitch    AngleDegrees
        OnGround byte
    }{
        packetIDPlayerPositionLook,
        position.X,
        position.Y,
        stance,
        position.Z,
        look.Yaw,
        look.Pitch,
        boolToByte(onGround),
    }
    return binary.Write(writer, binary.BigEndian, &packet)
}

func serverPlayerPositionLook(reader io.Reader, handler ServerPacketHandler) (err os.Error) {
    var packet struct {
        X        AbsCoord
        Stance   AbsCoord
        Y        AbsCoord
        Z        AbsCoord
        Yaw      AngleDegrees
        Pitch    AngleDegrees
        OnGround byte
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.PacketPlayerPosition(
        &AbsXYZ{
            packet.X,
            packet.Y,
            packet.Z,
        },
        packet.Stance,
        byteToBool(packet.OnGround))

    handler.PacketPlayerLook(
        &LookDegrees{
            packet.Yaw,
            packet.Pitch,
        },
        byteToBool(packet.OnGround))
    return
}

// packetIDPlayerDigging

func WritePlayerDigging(writer io.Writer, status DigStatus, blockLoc *BlockXYZ, face Face) (err os.Error) {
    var packet = struct {
        PacketID byte
        Status   DigStatus
        X        BlockCoord
        Y        BlockYCoord
        Z        BlockCoord
        Face     Face
    }{
        packetIDPlayerDigging,
        status,
        blockLoc.X, blockLoc.Y, blockLoc.Z,
        face,
    }

    return binary.Write(writer, binary.BigEndian, &packet)
}

func readPlayerDigging(reader io.Reader, handler PacketHandler) (err os.Error) {
    var packet struct {
        Status DigStatus
        X      BlockCoord
        Y      BlockYCoord
        Z      BlockCoord
        Face   Face
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.PacketPlayerDigging(
        packet.Status,
        &BlockXYZ{packet.X, packet.Y, packet.Z},
        packet.Face)
    return
}

// packetIDPlayerBlockPlacement

func WritePlayerBlockPlacement(writer io.Writer, itemID ItemID, blockLoc *BlockXYZ, face Face, amount ItemCount, uses ItemUses) (err os.Error) {
    var packet = struct {
        PacketID byte
        X        BlockCoord
        Y        BlockYCoord
        Z        BlockCoord
        Face     Face
        ItemID   ItemID
    }{
        packetIDPlayerBlockPlacement,
        blockLoc.X, blockLoc.Y, blockLoc.Z,
        face,
        itemID,
    }

    if err = binary.Write(writer, binary.BigEndian, &packet); err != nil {
        return
    }

    if itemID != -1 {
        var packetExtra = struct {
            Amount ItemCount
            Uses   ItemUses
        }{
            amount,
            uses,
        }
        err = binary.Write(writer, binary.BigEndian, &packetExtra)
    }

    return
}

func readPlayerBlockPlacement(reader io.Reader, handler PacketHandler) (err os.Error) {
    var packet struct {
        X      BlockCoord
        Y      BlockYCoord
        Z      BlockCoord
        Face   Face
        ItemID ItemID
    }
    var packetExtra struct {
        Amount ItemCount
        Uses   ItemUses
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

    handler.PacketPlayerBlockPlacement(
        packet.ItemID,
        &BlockXYZ{
            packet.X,
            packet.Y,
            packet.Z,
        },
        packet.Face,
        packetExtra.Amount,
        packetExtra.Uses)
    return
}

// packetIDHoldingChange

func WriteHoldingChange(writer io.Writer, itemID ItemID) (err os.Error) {
    var packet = struct {
        PacketID byte
        ItemID   ItemID
    }{
        packetIDHoldingChange,
        itemID,
    }

    return binary.Write(writer, binary.BigEndian, &packet)
}

func readHoldingChange(reader io.Reader, handler ServerPacketHandler) (err os.Error) {
    var itemID ItemID

    if err = binary.Read(reader, binary.BigEndian, &itemID); err != nil {
        return
    }

    handler.PacketHoldingChange(itemID)

    return
}

// packetIDUseBed

func WriteUseBed(writer io.Writer, flag bool, bedLoc *BlockXYZ) (err os.Error) {
    var packet = struct {
        PacketID byte
        Flag     byte
        X        BlockCoord
        Y        BlockYCoord
        Z        BlockCoord
    }{
        packetIDUseBed,
        boolToByte(flag),
        bedLoc.X,
        bedLoc.Y,
        bedLoc.Z,
    }

    return binary.Write(writer, binary.BigEndian, &packet)
}

func readUseBed(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packet struct {
        Flag     byte
        X        BlockCoord
        Y        BlockYCoord
        Z        BlockCoord
    }

    if err = binary.Read(reader, binary.BigEndian, &packet); err != nil {
        handler.PacketUseBed(
                byteToBool(packet.Flag),
                &BlockXYZ{packet.X, packet.Y, packet.Z})
    }

    return
}

// packetIDEntityAnimation

func WriteEntityAnimation(writer io.Writer, entityID EntityID, animation EntityAnimation) (err os.Error) {
    var packet = struct {
        PacketID  byte
        EntityID  EntityID
        Animation EntityAnimation
    }{
        packetIDEntityAnimation,
        entityID,
        animation,
    }

    return binary.Write(writer, binary.BigEndian, &packet)
}

func readEntityAnimation(reader io.Reader, handler PacketHandler) (err os.Error) {
    var packet struct {
        EntityID  EntityID
        Animation EntityAnimation
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.PacketEntityAnimation(packet.EntityID, packet.Animation)
    return
}

// packetIDEntityAction

func WriteEntityAction(writer io.Writer, entityID EntityID, action EntityAction) (err os.Error) {
    var packet = struct {
        PacketID byte
        EntityID EntityID
        Action   EntityAction
    }{
        packetIDEntityAction,
        entityID,
        action,
    }

    return binary.Write(writer, binary.BigEndian, &packet)
}


func readEntityAction(reader io.Reader, handler PacketHandler) (err os.Error) {
    var packet struct {
        EntityID EntityID
        Action   EntityAction
    }

    if err = binary.Read(reader, binary.BigEndian, &packet); err != nil {
        return
    }

    handler.PacketEntityAction(packet.EntityID, packet.Action)

    return
}

// packetIDNamedEntitySpawn

func WriteNamedEntitySpawn(writer io.Writer, entityID EntityID, name string, position *AbsIntXYZ, look *LookBytes, currentItem ItemID) (err os.Error) {
    var packetStart = struct {
        PacketID byte
        EntityID EntityID
    }{
        packetIDNamedEntitySpawn,
        entityID,
    }

    err = binary.Write(writer, binary.BigEndian, &packetStart)
    if err != nil {
        return
    }

    err = writeString(writer, name)
    if err != nil {
        return
    }

    var packetFinish = struct {
        X           AbsIntCoord
        Y           AbsIntCoord
        Z           AbsIntCoord
        Yaw         AngleBytes
        Pitch       AngleBytes
        CurrentItem ItemID
    }{
        position.X,
        position.Y,
        position.Z,
        look.Yaw,
        look.Pitch,
        currentItem,
    }

    err = binary.Write(writer, binary.BigEndian, &packetFinish)
    return
}

func readNamedEntitySpawn(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var entityID EntityID

    if err = binary.Read(reader, binary.BigEndian, &entityID); err != nil {
        return
    }

    var name string
    if name, err = readString(reader); err != nil {
        return
    }

    var packetEnd struct {
        X, Y, Z     AbsIntCoord
        Yaw, Pitch  AngleBytes
        CurrentItem ItemID
    }

    handler.PacketNamedEntitySpawn(
        entityID,
        name,
        &AbsIntXYZ{packetEnd.X, packetEnd.Y, packetEnd.Z},
        &LookBytes{packetEnd.Yaw, packetEnd.Pitch},
        packetEnd.CurrentItem)

    return
}

// packetIDItemSpawn

func WriteItemSpawn(writer io.Writer, entityID EntityID, itemType ItemID, amount ItemCount, uses ItemUses, position *AbsIntXYZ, orientation *OrientationBytes) os.Error {
    var packet = struct {
        PacketID byte
        EntityID EntityID
        ItemID   ItemID
        Count    ItemCount
        Uses     ItemUses
        X        AbsIntCoord
        Y        AbsIntCoord
        Z        AbsIntCoord
        Yaw      AngleBytes
        Pitch    AngleBytes
        Roll     AngleBytes
    }{
        packetIDItemSpawn,
        entityID,
        itemType,
        amount,
        uses,
        position.X,
        position.Y,
        position.Z,
        orientation.Yaw,
        orientation.Pitch,
        orientation.Roll,
    }

    return binary.Write(writer, binary.BigEndian, &packet)
}

func readItemSpawn(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packet struct {
        EntityID EntityID
        ItemID   ItemID
        Count    ItemCount
        Uses     ItemUses
        X        AbsIntCoord
        Y        AbsIntCoord
        Z        AbsIntCoord
        Yaw      AngleBytes
        Pitch    AngleBytes
        Roll     AngleBytes
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.PacketItemSpawn(
        packet.EntityID,
        packet.ItemID,
        packet.Count,
        packet.Uses,
        &AbsIntXYZ{packet.X, packet.Y, packet.Z},
        &OrientationBytes{packet.Yaw, packet.Pitch, packet.Roll})

    return
}

// packetIDItemCollect

func WriteItemCollect(writer io.Writer, collectedItem EntityID, collector EntityID) (err os.Error) {
    var packet = struct {
        PacketID      byte
        CollectedItem EntityID
        Collector     EntityID
    }{
        packetIDItemCollect,
        collectedItem,
        collector,
    }

    return binary.Write(writer, binary.BigEndian, &packet)
}

func readItemCollect(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packet struct {
        CollectedItem EntityID
        Collector     EntityID
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.PacketItemCollect(packet.CollectedItem, packet.Collector)

    return
}

// packetIDObjectSpawn

func WriteObjectSpawn(writer io.Writer, entityID EntityID, objType ObjTypeID, position *AbsIntXYZ) (err os.Error) {
    var packet = struct {
        PacketID byte
        EntityID EntityID
        ObjType  ObjTypeID
        X        AbsIntCoord
        Y        AbsIntCoord
        Z        AbsIntCoord
    }{
        packetIDObjectSpawn,
        entityID,
        objType,
        position.X,
        position.Y,
        position.Z,
    }

    return binary.Write(writer, binary.BigEndian, &packet)
}

func readObjectSpawn(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packet struct {
        EntityID EntityID
        ObjType  ObjTypeID
        X        AbsIntCoord
        Y        AbsIntCoord
        Z        AbsIntCoord
    }

    if err = binary.Read(reader, binary.BigEndian, &packet); err != nil {
        return
    }

    handler.PacketObjectSpawn(
        packet.EntityID,
        packet.ObjType,
        &AbsIntXYZ{packet.X, packet.Y, packet.Z})

    return
}

// packetIDEntitySpawn

func WriteEntitySpawn(writer io.Writer, entityID EntityID, mobType EntityMobType, position *AbsIntXYZ, look *LookBytes, data []EntityMetadata) (err os.Error) {
    var packet = struct {
        PacketID byte
        EntityID EntityID
        MobType  EntityMobType
        X        AbsIntCoord
        Y        AbsIntCoord
        Z        AbsIntCoord
        Yaw      AngleBytes
        Pitch    AngleBytes
    }{
        packetIDEntitySpawn,
        entityID,
        mobType,
        position.X, position.Y, position.Z,
        look.Yaw, look.Pitch,
    }

    if err = binary.Write(writer, binary.BigEndian, &packet); err != nil {
        return
    }

    return writeEntityMetadataField(writer, data)
}

func readEntitySpawn(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packet struct {
        EntityID EntityID
        MobType  EntityMobType
        X        AbsIntCoord
        Y        AbsIntCoord
        Z        AbsIntCoord
        Yaw      AngleBytes
        Pitch    AngleBytes
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    metadata, err := readEntityMetadataField(reader)
    if err != nil {
        return
    }

    handler.PacketEntitySpawn(
        EntityID(packet.EntityID), packet.MobType,
        &AbsIntXYZ{packet.X, packet.Y, packet.Z},
        &LookBytes{packet.Yaw, packet.Pitch},
        metadata)

    return err
}

// packetIDPaintingSpawn

func WritePaintingSpawn(writer io.Writer, entityID EntityID, title string, position *BlockXYZ, paintingType PaintingTypeID) (err os.Error) {
    if err = binary.Write(writer, binary.BigEndian, &entityID); err != nil {
        return
    }

    if err = writeString(writer, title); err != nil {
        return
    }

    var packetEnd = struct {
        X, Y, Z      BlockCoord
        PaintingType PaintingTypeID
    }{
        position.X, BlockCoord(position.Y), position.Z,
        paintingType,
    }

    return binary.Write(writer, binary.BigEndian, &packetEnd)
}

func readPaintingSpawn(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var entityID EntityID

    if err = binary.Read(reader, binary.BigEndian, &entityID); err != nil {
        return
    }

    title, err := readString(reader)
    if err != nil {
        return
    }

    var packetEnd struct {
        X, Y, Z      BlockCoord
        PaintingType PaintingTypeID
    }

    err = binary.Read(reader, binary.BigEndian, &packetEnd)
    if err != nil {
        return
    }

    handler.PacketPaintingSpawn(
        entityID,
        title,
        &BlockXYZ{packetEnd.X, BlockYCoord(packetEnd.Y), packetEnd.Z},
        packetEnd.PaintingType)

    return
}

// packetIDUnknown0x1b

func readUnknown0x1b(reader io.Reader, handler PacketHandler) (err os.Error) {
    var packet struct {
        field1, field2, field3, field4 float32
        field5, field6 byte
    }

    if err = binary.Read(reader, binary.BigEndian, &packet); err != nil {
        return
    }

    handler.PacketUnknown0x1b(
            packet.field1, packet.field2, packet.field3, packet.field4,
            byteToBool(packet.field5), byteToBool(packet.field6))

    return
}

// packetIDEntityVelocity

func WriteEntityVelocity(writer io.Writer, entityID EntityID, velocity *Velocity) (err os.Error) {
    var packet = struct {
        packetID byte
        EntityID EntityID
        X, Y, Z  VelocityComponent
    }{
        packetIDEntityVelocity,
        entityID,
        velocity.X,
        velocity.Y,
        velocity.Z,
    }

    return binary.Write(writer, binary.BigEndian, &packet)
}

func readEntityVelocity(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packet struct {
        EntityID EntityID
        X, Y, Z  VelocityComponent
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.PacketEntityVelocity(
        packet.EntityID,
        &Velocity{packet.X, packet.Y, packet.Z})

    return
}

// packetIDEntityDestroy

func WriteEntityDestroy(writer io.Writer, entityID EntityID) os.Error {
    var packet = struct {
        PacketID byte
        EntityID EntityID
    }{
        packetIDEntityDestroy,
        entityID,
    }
    return binary.Write(writer, binary.BigEndian, &packet)
}

func readEntityDestroy(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var entityID EntityID

    err = binary.Read(reader, binary.BigEndian, &entityID)
    if err != nil {
        return
    }

    handler.PacketEntityDestroy(entityID)

    return
}

// packetIDEntity

func WriteEntity(writer io.Writer, entityID EntityID) (err os.Error) {
    var packet = struct {
        PacketID byte
        EntityID EntityID
    }{
        packetIDEntity,
        entityID,
    }

    return binary.Write(writer, binary.BigEndian, &packet)
}

func readEntity(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var entityID EntityID

    err = binary.Read(reader, binary.BigEndian, &entityID)
    if err != nil {
        return
    }

    handler.PacketEntity(entityID)

    return
}

// packetIDEntityRelMove

func WriteEntityRelMove(writer io.Writer, entityID EntityID, movement *RelMove) (err os.Error) {
    var packet = struct {
        PacketID byte
        EntityID EntityID
        X, Y, Z  RelMoveCoord
    }{
        packetIDEntityRelMove,
        entityID,
        movement.X,
        movement.Y,
        movement.Z,
    }

    if err = binary.Write(writer, binary.BigEndian, &packet); err != nil {
        return
    }

    return
}

func readEntityRelMove(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packet struct {
        EntityID EntityID
        X, Y, Z  RelMoveCoord
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.PacketEntityRelMove(
        packet.EntityID,
        &RelMove{packet.X, packet.Y, packet.Z})

    return
}

// packetIDEntityLook

func WriteEntityLook(writer io.Writer, entityID EntityID, look *LookBytes) os.Error {
    var packet = struct {
        PacketID byte
        EntityID EntityID
        Yaw      AngleBytes
        Pitch    AngleBytes
    }{
        packetIDEntityLook,
        entityID,
        look.Yaw,
        look.Pitch,
    }
    return binary.Write(writer, binary.BigEndian, &packet)
}

func readEntityLook(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packet struct {
        EntityID EntityID
        Yaw      AngleBytes
        Pitch    AngleBytes
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.PacketEntityLook(
        packet.EntityID,
        &LookBytes{packet.Yaw, packet.Pitch})

    return
}

// packetIDEntityLookAndRelMove

func WriteEntityLookAndRelMove(writer io.Writer, entityID EntityID, movement *RelMove, look *LookBytes) (err os.Error) {
    var packet = struct {
        PacketID byte
        EntityID EntityID
        X, Y, Z  RelMoveCoord
        Yaw      AngleBytes
        Pitch    AngleBytes
    }{
        packetIDEntityLookAndRelMove,
        entityID,
        movement.X, movement.Y, movement.Z,
        look.Yaw, look.Pitch,
    }

    return binary.Write(writer, binary.BigEndian, &packet)
}

func readEntityLookAndRelMove(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packet struct {
        EntityID EntityID
        X, Y, Z  RelMoveCoord
        Yaw      AngleBytes
        Pitch    AngleBytes
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.PacketEntityRelMove(
        packet.EntityID,
        &RelMove{packet.X, packet.Y, packet.Z})

    handler.PacketEntityLook(
        packet.EntityID,
        &LookBytes{packet.Yaw, packet.Pitch})

    return
}

// packetIDEntityTeleport

func WriteEntityTeleport(writer io.Writer, entityID EntityID, position *AbsIntXYZ, look *LookBytes) os.Error {
    var packet = struct {
        PacketID byte
        EntityID EntityID
        X        AbsIntCoord
        Y        AbsIntCoord
        Z        AbsIntCoord
        Yaw      AngleBytes
        Pitch    AngleBytes
    }{
        packetIDEntityTeleport,
        entityID,
        position.X,
        position.Y,
        position.Z,
        look.Yaw,
        look.Pitch,
    }
    return binary.Write(writer, binary.BigEndian, &packet)
}

func readEntityTeleport(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packet struct {
        EntityID EntityID
        X        AbsIntCoord
        Y        AbsIntCoord
        Z        AbsIntCoord
        Yaw      AngleBytes
        Pitch    AngleBytes
    }

    if err = binary.Read(reader, binary.BigEndian, &packet); err != nil {
        return
    }

    handler.PacketEntityTeleport(
        packet.EntityID,
        &AbsIntXYZ{
            packet.X,
            packet.Y,
            packet.Z,
        },
        &LookBytes{
            packet.Yaw,
            packet.Pitch,
        })

    return
}

// packetIDEntityStatus

func WriteEntityStatus(writer io.Writer, entityID EntityID, status EntityStatus) (err os.Error) {
    var packet = struct {
        PacketID byte
        EntityID EntityID
        Status   EntityStatus
    }{
        packetIDEntityStatus,
        entityID,
        status,
    }

    return binary.Write(writer, binary.BigEndian, &packet)
}

func readEntityStatus(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packet struct {
        EntityID EntityID
        Status   EntityStatus
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.PacketEntityStatus(packet.EntityID, packet.Status)

    return
}

// packetIDEntityMetadata

func WriteEntityMetadata(writer io.Writer, entityID EntityID, data []EntityMetadata) (err os.Error) {
    var packet = struct {
        PacketID byte
        EntityID EntityID
    }{
        packetIDEntityMetadata,
        entityID,
    }

    if err = binary.Write(writer, binary.BigEndian, &packet); err != nil {
        return
    }

    return writeEntityMetadataField(writer, data)
}

func readEntityMetadata(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var entityID EntityID

    if err = binary.Read(reader, binary.BigEndian, &entityID); err != nil {
        return
    }

    metadata, err := readEntityMetadataField(reader)
    if err != nil {
        return
    }

    handler.PacketEntityMetadata(entityID, metadata)

    return
}

// packetIDPreChunk

func WritePreChunk(writer io.Writer, chunkLoc *ChunkXZ, mode ChunkLoadMode) os.Error {
    var packet = struct {
        PacketID byte
        X        ChunkCoord
        Z        ChunkCoord
        Mode     ChunkLoadMode
    }{
        packetIDPreChunk,
        chunkLoc.X,
        chunkLoc.Z,
        mode,
    }
    return binary.Write(writer, binary.BigEndian, &packet)
}

func readPreChunk(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packet struct {
        X    ChunkCoord
        Z    ChunkCoord
        Mode ChunkLoadMode
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.PacketPreChunk(&ChunkXZ{packet.X, packet.Z}, packet.Mode)

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

    chunkCornerLoc := chunkLoc.GetChunkCornerBlockXY()

    var packet = struct {
        PacketID         byte
        X                BlockCoord
        Y                int16
        Z                BlockCoord
        SizeX            SubChunkSizeCoord
        SizeY            SubChunkSizeCoord
        SizeZ            SubChunkSizeCoord
        CompressedLength int32
    }{
        packetIDMapChunk,
        chunkCornerLoc.X,
        int16(chunkCornerLoc.Y),
        chunkCornerLoc.Z,
        ChunkSizeH - 1,
        ChunkSizeY - 1,
        ChunkSizeH - 1,
        int32(len(bs)),
    }

    err = binary.Write(writer, binary.BigEndian, &packet)
    if err != nil {
        return
    }
    err = binary.Write(writer, binary.BigEndian, bs)
    return
}

func readMapChunk(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packet struct {
        X                BlockCoord
        Y                int16
        Z                BlockCoord
        SizeX            SubChunkSizeCoord
        SizeY            SubChunkSizeCoord
        SizeZ            SubChunkSizeCoord
        CompressedLength int32
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    // TODO extract block data from raw data field, and pass on to handler
    data := make([]byte, packet.CompressedLength)
    _, err = io.ReadFull(reader, data)
    if err != nil {
        return
    }

    handler.PacketMapChunk(
        &BlockXYZ{packet.X, BlockYCoord(packet.Y), packet.Z},
        &SubChunkSize{packet.SizeX, packet.SizeY, packet.SizeZ},
        data)
    return
}

// packetIDBlockChangeMulti

func WriteBlockChangeMulti(writer io.Writer, chunkLoc *ChunkXZ, blockCoords []SubChunkXYZ, blockTypes []BlockID, blockMetaData []byte) (err os.Error) {
    // NOTE that we don't yet check that blockCoords, blockTypes and
    // blockMetaData are of the same length.

    var packet = struct {
        PacketID byte
        ChunkX   ChunkCoord
        ChunkZ   ChunkCoord
        Count    int16
    }{
        packetIDBlockChangeMulti,
        chunkLoc.X, chunkLoc.Z,
        int16(len(blockCoords)),
    }

    if err = binary.Write(writer, binary.BigEndian, &packet); err != nil {
        return
    }

    rawBlockLocs := make([]int16, packet.Count)
    for index, blockCoord := range blockCoords {
        rawBlockCoord := int16(0)
        rawBlockCoord |= int16((blockCoord.X & 0x0f) << 12)
        rawBlockCoord |= int16((blockCoord.Y & 0xff))
        rawBlockCoord |= int16((blockCoord.Z & 0x0f) << 8)
        rawBlockLocs[index] = rawBlockCoord
    }

    binary.Write(writer, binary.BigEndian, rawBlockLocs)

    return
}

func readBlockChangeMulti(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packet struct {
        ChunkX ChunkCoord
        ChunkZ ChunkCoord
        Count  int16
    }

    if err = binary.Read(reader, binary.BigEndian, &packet); err != nil {
        return
    }

    rawBlockLocs := make([]int16, packet.Count)
    blockTypes := make([]BlockID, packet.Count)
    // blockMetadata array appears to represent one block per byte
    blockMetadata := make([]byte, packet.Count)

    err = binary.Read(reader, binary.BigEndian, rawBlockLocs)
    err = binary.Read(reader, binary.BigEndian, blockTypes)
    err = binary.Read(reader, binary.BigEndian, blockMetadata)

    blockLocs := make([]SubChunkXYZ, packet.Count)
    for index, rawLoc := range rawBlockLocs {
        blockLocs[index] = SubChunkXYZ{
            X:  SubChunkCoord(rawLoc >> 12),
            Y:  SubChunkCoord(rawLoc & 0xff),
            Z:  SubChunkCoord((rawLoc >> 8) & 0x0f),
        }
    }

    handler.PacketBlockChangeMulti(
        &ChunkXZ{packet.ChunkX, packet.ChunkZ},
        blockLocs,
        blockTypes,
        blockMetadata)

    return
}

// packetIDBlockChange

func WriteBlockChange(writer io.Writer, blockLoc *BlockXYZ, blockType BlockID, blockMetaData byte) (err os.Error) {
    var packet = struct {
        PacketID      byte
        X             BlockCoord
        Y             BlockYCoord
        Z             BlockCoord
        BlockType     BlockID
        BlockMetadata byte
    }{
        packetIDBlockChange,
        blockLoc.X,
        blockLoc.Y,
        blockLoc.Z,
        blockType,
        blockMetaData,
    }
    err = binary.Write(writer, binary.BigEndian, &packet)
    return
}

func readBlockChange(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packet struct {
        X             BlockCoord
        Y             BlockYCoord
        Z             BlockCoord
        BlockType     BlockID
        BlockMetadata byte
    }

    err = binary.Read(reader, binary.BigEndian, &packet)
    if err != nil {
        return
    }

    handler.PacketBlockChange(
        &BlockXYZ{packet.X, packet.Y, packet.Z},
        packet.BlockType,
        packet.BlockMetadata)

    return
}

// packetIDNoteBlockPlay

func WriteNoteBlockPlay(writer io.Writer, position *BlockXYZ, instrument InstrumentID, pitch NotePitch) (err os.Error) {
    var packet = struct {
        PacketID   byte
        X          BlockCoord
        Y          BlockYCoord
        Z          BlockCoord
        Instrument InstrumentID
        Pitch      NotePitch
    }{
        packetIDNoteBlockPlay,
        position.X, position.Y, position.Z,
        instrument,
        pitch,
    }

    return binary.Write(writer, binary.BigEndian, &packet)
}

func readNoteBlockPlay(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packet struct {
        X          BlockCoord
        Y          BlockYCoord
        Z          BlockCoord
        Instrument InstrumentID
        Pitch      NotePitch
    }

    if err = binary.Read(reader, binary.BigEndian, &packet); err != nil {
        return
    }

    handler.PacketNoteBlockPlay(
        &BlockXYZ{packet.X, packet.Y, packet.Z},
        packet.Instrument,
        packet.Pitch)

    return
}

// packetIDExplosion

// TODO introduce better types for ExplosionOffsetXYZ and the floats in the
// packet structure when the packet is better understood.

type ExplosionOffsetXYZ struct {
    X, Y, Z int8
}

func WriteExplosion(writer io.Writer, position *AbsXYZ, power float32, blockOffsets []ExplosionOffsetXYZ) (err os.Error) {
    var packet = struct {
        PacketID byte
        // NOTE AbsCoord is just a guess for now
        X, Y, Z AbsCoord
        // NOTE Power isn't known to be a good name for this field
        Power     float32
        NumBlocks int32
    }{
        packetIDExplosion,
        position.X, position.Y, position.Z,
        power,
        int32(len(blockOffsets)),
    }

    if err = binary.Write(writer, binary.BigEndian, &packet); err != nil {
        return
    }

    return binary.Write(writer, binary.BigEndian, blockOffsets)
}

func readExplosion(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packet struct {
        // NOTE AbsCoord is just a guess for now
        X, Y, Z AbsCoord
        // NOTE Power isn't known to be a good name for this field
        Power     float32
        NumBlocks int32
    }

    if err = binary.Read(reader, binary.BigEndian, &packet); err != nil {
        return
    }

    // TODO put sensible size limits on how big arrays read from network could
    // be, both here and in other places
    blockOffsets := make([]ExplosionOffsetXYZ, packet.NumBlocks)

    if err = binary.Read(reader, binary.BigEndian, blockOffsets); err != nil {
        return
    }

    handler.PacketExplosion(
        &AbsXYZ{packet.X, packet.Y, packet.Z},
        packet.Power,
        blockOffsets)

    return
}

// packetIDWindowOpen

func WriteWindowOpen(writer io.Writer, windowID WindowID, invTypeID InvTypeID, windowTitle string, numSlots byte) (err os.Error) {
    var packet = struct {
        PacketID  byte
        WindowID  WindowID
        InvTypeID InvTypeID
    }{
        packetIDWindowOpen,
        windowID,
        invTypeID,
    }

    if err = binary.Write(writer, binary.BigEndian, &packet); err != nil {
        return
    }

    if err = writeString(writer, windowTitle); err != nil {
        return
    }

    return binary.Write(writer, binary.BigEndian, numSlots)
}

func readWindowOpen(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packet struct {
        WindowID  WindowID
        InvTypeID InvTypeID
    }

    if err = binary.Read(reader, binary.BigEndian, &packet); err != nil {
        return
    }

    windowTitle, err := readString(reader)
    if err != nil {
        return
    }

    var numSlots byte
    if err = binary.Read(reader, binary.BigEndian, &numSlots); err != nil {
        return
    }

    handler.PacketWindowOpen(packet.WindowID, packet.InvTypeID, windowTitle, numSlots)

    return
}

// packetIDWindowClose

func WriteWindowClose(writer io.Writer, windowID WindowID) (err os.Error) {
    var packet = struct {
        PacketID byte
        WindowID WindowID
    }{
        packetIDWindowClose,
        windowID,
    }

    if err = binary.Write(writer, binary.BigEndian, &packet); err != nil {
        return
    }

    return
}

func readWindowClose(reader io.Reader, handler ServerPacketHandler) (err os.Error) {
    var windowID WindowID

    if err = binary.Read(reader, binary.BigEndian, &windowID); err != nil {
        return
    }

    handler.PacketWindowClose(windowID)

    return
}

// packetIDWindowClick

func WriteWindowClick(writer io.Writer, windowID WindowID, slot SlotID, rightClick bool, txID TxID, itemID ItemID, amount ItemCount, uses ItemUses) (err os.Error) {
    var packet = struct {
        PacketID   byte
        WindowID   WindowID
        Slot       SlotID
        RightClick byte
        TxID       TxID
        ItemID     ItemID
    }{
        packetIDWindowClick,
        windowID,
        slot,
        boolToByte(rightClick),
        txID,
        itemID,
    }

    if err = binary.Write(writer, binary.BigEndian, &packet); err != nil {
        return
    }

    if itemID != -1 {
        var packetEnd = struct {
            Amount ItemCount
            Uses   ItemUses
        }{
            amount,
            uses,
        }
        err = binary.Write(writer, binary.BigEndian, &packetEnd)
    }

    return
}

func readWindowClick(reader io.Reader, handler ServerPacketHandler) (err os.Error) {
    var packetStart struct {
        WindowID   WindowID
        Slot       SlotID
        RightClick byte
        TxID       TxID
        ItemID     ItemID
    }

    err = binary.Read(reader, binary.BigEndian, &packetStart)
    if err != nil {
        return
    }

    var packetEnd struct {
        Amount ItemCount
        Uses   ItemUses
    }

    if packetStart.ItemID != -1 {
        err = binary.Read(reader, binary.BigEndian, &packetEnd)
        if err != nil {
            return
        }
    }

    handler.PacketWindowClick(
        packetStart.WindowID,
        packetStart.Slot,
        byteToBool(packetStart.RightClick),
        packetStart.TxID,
        packetStart.ItemID,
        packetEnd.Amount,
        packetEnd.Uses)

    return
}

// packetIDWindowSetSlot

func WriteWindowSetSlot(writer io.Writer, windowID WindowID, slot SlotID, itemID ItemID, amount ItemCount, uses ItemUses) (err os.Error) {
    var packet = struct {
        PacketID byte
        WindowID WindowID
        Slot     SlotID
        ItemID   ItemID
    }{
        packetIDWindowSetSlot,
        windowID,
        slot,
        itemID,
    }

    if err = binary.Write(writer, binary.BigEndian, &packet); err != nil {
        return
    }

    if itemID != -1 {
        var packetEnd = struct {
            Amount ItemCount
            Uses   ItemUses
        }{
            amount,
            uses,
        }
        err = binary.Write(writer, binary.BigEndian, &packetEnd)
    }

    return err
}

func readWindowSetSlot(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packetStart struct {
        WindowID WindowID
        Slot     SlotID
        ItemID   ItemID
    }

    err = binary.Read(reader, binary.BigEndian, &packetStart)
    if err != nil {
        return
    }

    var packetEnd struct {
        Amount ItemCount
        Uses   ItemUses
    }

    if packetStart.ItemID != -1 {
        err = binary.Read(reader, binary.BigEndian, &packetEnd)
        if err != nil {
            return
        }
    }

    handler.PacketWindowSetSlot(
        packetStart.WindowID,
        packetStart.Slot,
        packetStart.ItemID,
        packetEnd.Amount,
        packetEnd.Uses)

    return
}

// packetIDWindowItems

func WriteWindowItems(writer io.Writer, windowID WindowID, items []WindowSlot) (err os.Error) {
    var packet = struct {
        PacketID byte
        WindowID WindowID
        Count    int16
    }{
        packetIDWindowItems,
        windowID,
        int16(len(items)),
    }

    if err = binary.Write(writer, binary.BigEndian, &packet); err != nil {
        return
    }

    for _, slot := range items {
        if err = binary.Write(writer, binary.BigEndian, slot.ItemID); err != nil {
            return
        }

        if slot.ItemID != -1 {
            var itemInfo struct {
                Amount ItemCount
                Uses   ItemUses
            }
            if err = binary.Write(writer, binary.BigEndian, &itemInfo); err != nil {
                return
            }
        }
    }

    return
}

func readWindowItems(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packetStart struct {
        WindowID WindowID
        Count    int16
    }

    err = binary.Read(reader, binary.BigEndian, &packetStart)
    if err != nil {
        return
    }

    var itemID ItemID

    items := make([]WindowSlot, packetStart.Count)

    for i := int16(0); i < packetStart.Count; i++ {
        err = binary.Read(reader, binary.BigEndian, &itemID)
        if err != nil {
            return
        }

        var itemInfo struct {
            Amount ItemCount
            Uses   ItemUses
        }
        if itemID != -1 {
            err = binary.Read(reader, binary.BigEndian, &itemInfo)
            if err != nil {
                return
            }
        }

        items = append(items, WindowSlot{
            ItemID: itemID,
            Amount: itemInfo.Amount,
            Uses:   itemInfo.Uses,
        })
    }

    handler.PacketWindowItems(
        packetStart.WindowID,
        items)

    return
}

// packetIDWindowProgressBar

func WriteWindowProgressBar(writer io.Writer, windowID WindowID, prgBarID PrgBarID, value PrgBarValue) os.Error {
    var packet = struct {
        PacketID byte
        WindowID WindowID
        PrgBarID PrgBarID
        Value    PrgBarValue
    }{
        packetIDWindowProgressBar,
        windowID,
        prgBarID,
        value,
    }

    return binary.Write(writer, binary.BigEndian, &packet)
}

func readWindowProgressBar(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packet struct {
        WindowID WindowID
        PrgBarID PrgBarID
        Value    PrgBarValue
    }

    if err = binary.Read(reader, binary.BigEndian, &packet); err != nil {
        return
    }

    handler.PacketWindowProgressBar(packet.WindowID, packet.PrgBarID, packet.Value)

    return
}

// packetIDWindowTransaction

func WriteWindowTransaction(writer io.Writer, windowID WindowID, txID TxID, accepted bool) (err os.Error) {
    var packet = struct {
        PacketID byte
        WindowID WindowID
        TxID     TxID
        Accepted byte
    }{
        packetIDWindowTransaction,
        windowID,
        txID,
        boolToByte(accepted),
    }

    return binary.Write(writer, binary.BigEndian, &packet)
}

func readWindowTransaction(reader io.Reader, handler ClientPacketHandler) (err os.Error) {
    var packet struct {
        WindowID WindowID
        TxID     TxID
        Accepted byte
    }

    if err = binary.Read(reader, binary.BigEndian, &packet); err != nil {
        return
    }

    handler.PacketWindowTransaction(packet.WindowID, packet.TxID, byteToBool(packet.Accepted))

    return
}

// packetIDSignUpdate

func WriteSignUpdate(writer io.Writer, position *BlockXYZ, lines [4]string) (err os.Error) {
    var packet = struct {
        PacketID byte
        X        BlockCoord
        Y        BlockYCoord
        Z        BlockCoord
    }{
        packetIDSignUpdate,
        position.X, position.Y, position.Z,
    }

    if err = binary.Write(writer, binary.BigEndian, &packet); err != nil {
        return
    }

    for _, line := range lines {
        if err = writeString(writer, line); err != nil {
            return
        }
    }

    return
}

func readSignUpdate(reader io.Reader, handler PacketHandler) (err os.Error) {
    var packet struct {
        X   BlockCoord
        Y   BlockYCoord
        Z   BlockCoord
    }

    if err = binary.Read(reader, binary.BigEndian, &packet); err != nil {
        return
    }

    var lines [4]string

    for i := 0; i < len(lines); i++ {
        if lines[i], err = readString(reader); err != nil {
            return
        }
    }

    handler.PacketSignUpdate(
        &BlockXYZ{packet.X, packet.Y, packet.Z},
        lines)

    return
}

// packetIDDisconnect

func WriteDisconnect(writer io.Writer, reason string) (err os.Error) {
    buf := &bytes.Buffer{}
    binary.Write(buf, binary.BigEndian, byte(packetIDDisconnect))
    writeString(buf, reason)
    _, err = writer.Write(buf.Bytes())
    return
}

func readDisconnect(reader io.Reader, handler PacketHandler) (err os.Error) {
    reason, err := readString(reader)
    if err != nil {
        return
    }

    handler.PacketDisconnect(reason)
    return
}


// End of packet reader/writer functions


type commonPacketHandler func(io.Reader, PacketHandler) os.Error
type serverPacketHandler func(io.Reader, ServerPacketHandler) os.Error
type clientPacketHandler func(io.Reader, ClientPacketHandler) os.Error

type commonPacketReaderMap map[byte]commonPacketHandler
type serverPacketReaderMap map[byte]serverPacketHandler
type clientPacketReaderMap map[byte]clientPacketHandler

// Common packet mapping
var commonReadFns = commonPacketReaderMap{
    packetIDKeepAlive:            readKeepAlive,
    packetIDChatMessage:          readChatMessage,
    packetIDEntityAction:         readEntityAction,
    packetIDUseEntity:            readUseEntity,
    packetIDRespawn:              readRespawn,
    packetIDPlayerPosition:       readPlayerPosition,
    packetIDPlayerLook:           readPlayerLook,
    packetIDPlayerDigging:        readPlayerDigging,
    packetIDPlayerBlockPlacement: readPlayerBlockPlacement,
    packetIDEntityAnimation:      readEntityAnimation,
    packetIDSignUpdate:           readSignUpdate,
    packetIDDisconnect:           readDisconnect,
}

// Client->server specific packet mapping
var serverReadFns = serverPacketReaderMap{
    packetIDPlayer:             readPlayer,
    packetIDPlayerPositionLook: serverPlayerPositionLook,
    packetIDWindowClick:        readWindowClick,
    packetIDHoldingChange:      readHoldingChange,
    packetIDWindowClose:        readWindowClose,
}

// Server->client specific packet mapping
var clientReadFns = clientPacketReaderMap{
    packetIDLogin:                clientReadLogin,
    packetIDTimeUpdate:           readTimeUpdate,
    packetIDEntityEquipment:      readEntityEquipment,
    packetIDSpawnPosition:        readSpawnPosition,
    packetIDUpdateHealth:         readUpdateHealth,
    packetIDPlayerPositionLook:   readPlayerPositionLook,
    packetIDUseBed:               readUseBed,
    packetIDNamedEntitySpawn:     readNamedEntitySpawn,
    packetIDItemSpawn:            readItemSpawn,
    packetIDItemCollect:          readItemCollect,
    packetIDObjectSpawn:          readObjectSpawn,
    packetIDEntitySpawn:          readEntitySpawn,
    packetIDPaintingSpawn:        readPaintingSpawn,
    packetIDEntityVelocity:       readEntityVelocity,
    packetIDEntityDestroy:        readEntityDestroy,
    packetIDEntity:               readEntity,
    packetIDEntityRelMove:        readEntityRelMove,
    packetIDEntityLook:           readEntityLook,
    packetIDEntityLookAndRelMove: readEntityLookAndRelMove,
    packetIDEntityTeleport:       readEntityTeleport,
    packetIDEntityStatus:         readEntityStatus,
    packetIDEntityMetadata:       readEntityMetadata,
    packetIDPreChunk:             readPreChunk,
    packetIDMapChunk:             readMapChunk,
    packetIDBlockChangeMulti:     readBlockChangeMulti,
    packetIDBlockChange:          readBlockChange,
    packetIDNoteBlockPlay:        readNoteBlockPlay,
    packetIDExplosion:            readExplosion,
    packetIDWindowOpen:           readWindowOpen,
    packetIDWindowSetSlot:        readWindowSetSlot,
    packetIDWindowItems:          readWindowItems,
    packetIDWindowProgressBar:    readWindowProgressBar,
    packetIDWindowTransaction:    readWindowTransaction,
}

// A server should call this to receive a single packet from a client. It will
// block until a packet was successfully handled, or there was an error.
func ServerReadPacket(reader io.Reader, handler ServerPacketHandler) os.Error {
    var packetID byte

    if err := binary.Read(reader, binary.BigEndian, &packetID); err != nil {
        return err
    }

    if commonFn, ok := commonReadFns[packetID]; ok {
        return commonFn(reader, handler)
    }

    if serverFn, ok := serverReadFns[packetID]; ok {
        return serverFn(reader, handler)
    }

    return os.NewError(fmt.Sprintf("unhandled packet type %#x", packetID))
}

// A client should call this to receive a single packet from a client. It will
// block until a packet was successfully handled, or there was an error.
func ClientReadPacket(reader io.Reader, handler ClientPacketHandler) os.Error {
    var packetID byte

    if err := binary.Read(reader, binary.BigEndian, &packetID); err != nil {
        return err
    }

    if commonFn, ok := commonReadFns[packetID]; ok {
        return commonFn(reader, handler)
    }

    if clientFn, ok := clientReadFns[packetID]; ok {
        return clientFn(reader, handler)
    }

    return os.NewError(fmt.Sprintf("unhandled packet type %#x", packetID))
}

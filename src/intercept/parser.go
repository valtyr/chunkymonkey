package intercept_parse

import (
    "encoding/hex"
    "io"
    "log"
    "os"

    "chunkymonkey/proto"
    .   "chunkymonkey/types"
)

// Hex dumps the input to the log
func (p *MessageParser) dumpInput(logPrefix string, reader io.Reader) {
    buf := make([]byte, 16, 16)
    for {
        _, err := io.ReadAtLeast(reader, buf, 1)
        if err != nil {
            return
        }

        hexData := hex.EncodeToString(buf)
        p.printf("Unparsed data: %s", hexData)
    }
}

// Consumes data from reader until an error occurs
func (p *MessageParser) consumeUnrecognizedInput(reader io.Reader) {
    p.printf("Lost packet sync. Ignoring further data.")
    buf := make([]byte, 4096)
    for {
        _, err := io.ReadAtLeast(reader, buf, 1)
        if err != nil {
            return
        }
    }
}

type MessageParser struct {
    logPrefix string
}

func (p *MessageParser) printf(format string, v ...interface{}) {
    log.Printf(p.logPrefix+format, v...)
}

func (p *MessageParser) PacketKeepAlive() {
}

func (p *MessageParser) PacketChatMessage(message string) {
    p.printf("PacketChatMessage(%s)", message)
}

func (p *MessageParser) PacketOnGround(onGround bool) {
    p.printf("PacketOnGround(%v)", onGround)
}

func (p *MessageParser) PacketPlayerPosition(position *AbsXYZ, stance AbsCoord, onGround bool) {
    p.printf("PacketPlayerPosition(position=%v, stance=%v, onGround=%v)", position, stance, onGround)
}

func (p *MessageParser) PacketPlayerLook(look *LookDegrees, onGround bool) {
    p.printf("PacketPlayerLook(look=%v, onGround=%v)", look, onGround)
}

func (p *MessageParser) PacketPlayerDigging(status DigStatus, blockLoc *BlockXYZ, face Face) {
    p.printf("PacketPlayerDigging(status=%v, blockLoc=%v, face=%v)", status, blockLoc, face)
}

func (p *MessageParser) PacketPlayerBlockPlacement(itemID ItemID, blockLoc *BlockXYZ, direction Face, amount ItemCount, uses ItemUses) {
    p.printf("PacketPlayerBlockPlacement(itemId=%d, blockLoc=%v, direction=%d, amount=%d, uses=%d)",
        itemID, blockLoc, direction, amount, uses)
}

func (p *MessageParser) PacketHoldingChange(itemID ItemID) {
    p.printf("PacketHoldingChange(%d)", itemID)
}

func (p *MessageParser) PacketPlayerAnimation(animation PlayerAnimation) {
    p.printf("PacketPlayerAnimation(%v)", animation)
}

func (p *MessageParser) PacketDisconnect(reason string) {
    p.printf("PacketDisconnect(%s)", reason)
}

func (p *MessageParser) ClientPacketLogin(entityID EntityID, str1 string, str2 string, mapSeed RandomSeed, dimension DimensionID) {
    p.printf("PacketLogin(entityID=%d, str1=%v, str2=%v, mapSeed=%d, dimension=%d)",
        entityID, str1, str2, mapSeed, dimension)
}

func (p *MessageParser) PacketTimeUpdate(time TimeOfDay) {
    p.printf("PacketTime(time=%d)", time)
}

func (p *MessageParser) PacketEntityEquipment(entityID EntityID, slot SlotID, itemID ItemID, uses ItemUses) {
    p.printf("PacketEntityEquipment(entityID=%d, slot=%d, itemID=%d, uses=%d)",
        entityID, slot, itemID, uses)
}

func (p *MessageParser) PacketSpawnPosition(position *BlockXYZ) {
    p.printf("PacketSpawnPosition(position=%v)", position)
}

func (p *MessageParser) PacketUseEntity(user EntityID, target EntityID, leftClick bool) {
    p.printf("PacketUseEntity(user=%d, target=%d, leftClick=%v)", user, target, leftClick)
}

func (p *MessageParser) PacketUpdateHealth(health int16) {
    p.printf("PacketUpdateHealth(health=%d)", health)
}

func (p *MessageParser) PacketItemSpawn(entityID EntityID, itemID ItemID, count ItemCount, uses ItemUses, location *AbsIntXYZ, yaw, pitch, roll AngleBytes) {
    p.printf("PacketItemSpawn(entityID=%d, itemID=%d, count=%d, uses=%d, location=%v, yaw=%d, pitch=%d, roll=%d)",
        entityID, itemID, count, uses, location, yaw, pitch, roll)
}

func (p *MessageParser) PacketItemCollect(collectedItem EntityID, collector EntityID) {
    p.printf("PacketItemCollect(collectedItem=%d, collector=%d)",
        collectedItem, collector)
}

func (p *MessageParser) PacketEntitySpawn(entityID EntityID, mobType EntityMobType, position *AbsIntXYZ, yaw AngleBytes, pitch AngleBytes, data []proto.UnknownEntityExtra) {
    p.printf("PacketEntitySpawn(entityID=%d, mobType=%d, position=%v, yaw=%d, pitch=%d, data=%v)",
        entityID, mobType, position, yaw, pitch, data)
}

func (p *MessageParser) PacketUnknownX19(field1 int32, field2 string, field3, field4, field5, field6 int32) {
    p.printf("PacketUnknownX19(field1=%d, field2=%v, field3=%d, field4=%d, field5=%d, field6=%d)",
        field1, field2, field3, field4, field5, field6)
}

func (p *MessageParser) PacketEntityVelocity(entityID EntityID, velocity *Velocity) {
    p.printf("PacketEntityVelocity(entityID=%d, velocity=%v)",
        entityID, velocity)
}

func (p *MessageParser) PacketEntityDestroy(entityID EntityID) {
    p.printf("PacketEntityDestroy(entityID=%d)", entityID)
}

func (p *MessageParser) PacketEntity(entityID EntityID) {
    p.printf("PacketEntity(entityID=%d)", entityID)
}

func (p *MessageParser) PacketEntityRelMove(entityID EntityID, movement *RelMove) {
    p.printf("PacketEntityRelMove(entityID=%d, movement=%v)",
        entityID, movement)
}

func (p *MessageParser) PacketEntityLook(entityID EntityID, yaw, pitch AngleBytes) {
    p.printf("PacketEntityLook(entityID=%d, yaw=%d, pitch=%d)",
        entityID, yaw, pitch)
}

func (p *MessageParser) PacketEntityStatus(entityID EntityID, status EntityStatus) {
    p.printf("PacketEntityStatus(entityID=%d, status=%d",
        entityID, status)
}

func (p *MessageParser) PacketUnknownX28(field1 int32, data []proto.UnknownEntityExtra) {
    p.printf("PacketUnknownX28(field1=%d, data=%v)", field1, data)
}

func (p *MessageParser) PacketPreChunk(position *ChunkXZ, mode ChunkLoadMode) {
    p.printf("PacketPreChunk(position=%v, mode=%d)", position, mode)
}

func (p *MessageParser) PacketMapChunk(position *BlockXYZ, size *SubChunkSize, data []byte) {
    p.printf("PacketMapChunk(position=%v, size=%v, len(data)=%d)",
        position, size, len(data))
}

func (p *MessageParser) PacketBlockChangeMulti(chunkLoc *ChunkXZ, blockCoords []SubChunkXYZ, blockTypes []BlockID, blockMetaData []byte) {
    p.printf("PacketBlockChangeMulti(chunkLoc=%v, blockCoords=%v, blockTypes=%v, blockMetaData=%v)",
        chunkLoc, blockCoords, blockTypes, blockMetaData)
}

func (p *MessageParser) PacketBlockChange(blockLoc *BlockXYZ, blockType BlockID, blockMetaData byte) {
    p.printf("PacketBlockChange(blockLoc=%v, blockType=%d, blockMetaData=%d)",
        blockLoc, blockType, blockMetaData)
}

func (p *MessageParser) PacketUnknownX36(field1 int32, field2 int16, field3 int32, field4, field5 byte) {
    p.printf("PacketUnknownX36(field1=%d, field2=%d, field3=%d, field4=%d, field5=%d)",
        field1, field2, field3, field4, field5)
}

func (p *MessageParser) PacketWindowClick(windowID WindowID, slot SlotID, rightClick bool, txID TxID, itemID ItemID, amount ItemCount, uses ItemUses) {
    p.printf("PacketWindowClick(windowID=%d, slot=%d, rightClick=%v, txID=%d, itemID=%d, amount=%d, uses=%d)",
        windowID, slot, rightClick, txID, itemID, amount, uses)
}

func (p *MessageParser) PacketSetSlot(windowID WindowID, slot SlotID, itemID ItemID, amount ItemCount, uses ItemUses) {
    p.printf("PacketSetSlot(windowID=%d, slot=%d, itemID=%d, amount=%d, uses=%d)",
        windowID, slot, itemID, amount, uses)
}

func (p *MessageParser) PacketWindowItems(windowID WindowID, items []proto.WindowSlot) {
    p.printf("PacketWindowItems(windowID=%d, items=%v)",
        windowID, items)
}

// Parses messages from the client
func (p *MessageParser) CSParse(reader io.Reader) {
    // If we return, we should consume all input to avoid blocking the pipe
    // we're listening on. TODO Maybe we could just close it?
    defer p.consumeUnrecognizedInput(reader)

    defer func() {
        if err := recover(); err != nil {
            p.printf("Parsing failed: %v", err)
        }
    }()

    p.logPrefix = "(C->S) "

    username, err := proto.ServerReadHandshake(reader)
    if err != nil {
        p.printf("ServerReadHandshake error: %v", err)
        return
    }
    p.printf("ServerReadHandshake(username=%v)", username)

    loginUsername, _, err := proto.ServerReadLogin(reader)
    if err != nil {
        p.printf("ServerReadLogin error: %v", err)
        return
    }
    p.printf("ServerReadLogin(username=%v)", loginUsername)

    for {
        err := proto.ServerReadPacket(reader, p)
        if err != nil {
            if err != os.EOF {
                p.printf("ReceiveLoop failed: %v", err)
            } else {
                p.printf("ReceiveLoop hit EOF")
            }
            return
        }
    }
}

// Parses messages from the server
func (p *MessageParser) SCParse(reader io.Reader) {
    // If we return, we should consume all input to avoid blocking the pipe
    // we're listening on. TODO Maybe we could just close it?
    defer p.consumeUnrecognizedInput(reader)

    defer func() {
        if err := recover(); err != nil {
            p.printf("Parsing failed: %v", err)
        }
    }()

    p.logPrefix = "(S->C) "

    connectionHash, err := proto.ClientReadHandshake(reader)
    if err != nil {
        p.printf("ClientReadHandshake error: %v", err)
        return
    }
    p.printf("ClientReadHandshake(connectionHash=%v)", connectionHash)

    for {
        err := proto.ClientReadPacket(reader, p)
        if err != nil {
            if err != os.EOF {
                p.printf("ReceiveLoop failed: %v", err)
            } else {
                p.printf("ReceiveLoop hit EOF")
            }
            return
        }
    }
}

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
    // Not logging this packet as it's a bit spammy
}

func (p *MessageParser) PacketChatMessage(message string) {
    p.printf("PacketChatMessage(%s)", message)
}

func (p *MessageParser) PacketRespawn() {
    p.printf("PacketRespawn()")
}

func (p *MessageParser) PacketPlayer(onGround bool) {
    // Not logging this packet as it's a bit spammy
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

func (p *MessageParser) PacketPlayerBlockPlacement(itemID ItemID, blockLoc *BlockXYZ, face Face, amount ItemCount, uses ItemUses) {
    p.printf("PacketPlayerBlockPlacement(itemId=%d, blockLoc=%v, face=%d, amount=%d, uses=%d)",
        itemID, blockLoc, face, amount, uses)
}

func (p *MessageParser) PacketHoldingChange(itemID ItemID) {
    p.printf("PacketHoldingChange(itemID=%d)", itemID)
}

func (p *MessageParser) PacketBedUse(flag bool, bedLoc *BlockXYZ) {
    p.printf("PacketBedUse(flag=%v, bedLoc=%v)", flag, bedLoc)
}

func (p *MessageParser) PacketEntityAnimation(entityID EntityID, animation EntityAnimation) {
    p.printf("PacketEntityAnimation(entityID=%d, animation=%v)", entityID, animation)
}

func (p *MessageParser) PacketEntityAction(entityID EntityID, action EntityAction) {
    p.printf("PacketEntityAction(entityID=%d, action=%d)",
        entityID, action)
}

func (p *MessageParser) PacketSignUpdate(position *BlockXYZ, lines [4]string) {
    p.printf("PacketSignUpdate(position=%v, lines=[%q, %q, %q, %q])",
        position,
        lines[0], lines[1], lines[2], lines[3])
}

func (p *MessageParser) PacketDisconnect(reason string) {
    p.printf("PacketDisconnect(%s)", reason)
}

func (p *MessageParser) ClientPacketLogin(entityID EntityID, mapSeed RandomSeed, dimension DimensionID) {
    p.printf("PacketLogin(entityID=%d, mapSeed=%d, dimension=%d)",
        entityID, mapSeed, dimension)
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

func (p *MessageParser) PacketNamedEntitySpawn(entityID EntityID, name string, position *AbsIntXYZ, look *LookBytes, currentItem ItemID) {
    p.printf("PacketNamedEntitySpawn(entityID=%d, name=%q, position=%v, look=%v, currentItem=%d)",
        entityID, name, position, look, currentItem)
}

func (p *MessageParser) PacketItemSpawn(entityID EntityID, itemID ItemID, count ItemCount, uses ItemUses, location *AbsIntXYZ, orientation *OrientationBytes) {
    p.printf("PacketItemSpawn(entityID=%d, itemID=%d, count=%d, uses=%d, location=%v, orientation=%v)",
        entityID, itemID, count, uses, location, orientation)
}

func (p *MessageParser) PacketItemCollect(collectedItem EntityID, collector EntityID) {
    p.printf("PacketItemCollect(collectedItem=%d, collector=%d)",
        collectedItem, collector)
}

func (p *MessageParser) PacketObjectSpawn(entityID EntityID, objType ObjTypeID, position *AbsIntXYZ) {
    p.printf("PacketObjectSpawn(entityID=%d, objType=%d, position=%v)",
        entityID, objType, position)
}

func (p *MessageParser) PacketEntitySpawn(entityID EntityID, mobType EntityMobType, position *AbsIntXYZ, look *LookBytes, metadata []proto.EntityMetadata) {
    p.printf("PacketEntitySpawn(entityID=%d, mobType=%d, position=%v, look=%v, metadata=%v)",
        entityID, mobType, position, look, metadata)
}

func (p *MessageParser) PacketPaintingSpawn(entityID EntityID, title string, position *BlockXYZ, paintingType PaintingTypeID) {
    p.printf("PacketPaintingSpawn(entityID=%d, title=%s, position=%v, paintingType=%d)",
        entityID, title, position, position, paintingType)
}

func (p *MessageParser) PacketUnknown0x1b(field1, field2, field3, field4 float32, field5, field6 bool) {
    p.printf(
        "PacketUnknown0x1b(field1=%v, field2=%v, field3=%v, field4=%v, field5=%v, field6=%v)",
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

func (p *MessageParser) PacketEntityLook(entityID EntityID, look *LookBytes) {
    p.printf("PacketEntityLook(entityID=%d, look=%v)",
        entityID, look)
}

func (p *MessageParser) PacketEntityTeleport(entityID EntityID, position *AbsIntXYZ, look *LookBytes) {
    p.printf("PacketEntityTeleport(entityID=%d, position=%v, look=%v",
        entityID, position, look)
}

func (p *MessageParser) PacketEntityStatus(entityID EntityID, status EntityStatus) {
    p.printf("PacketEntityStatus(entityID=%d, status=%d",
        entityID, status)
}

func (p *MessageParser) PacketEntityMetadata(entityID EntityID, metadata []proto.EntityMetadata) {
    p.printf("PacketEntityMetadata(entityID=%d, metadata=%v)", entityID, metadata)
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

func (p *MessageParser) PacketNoteBlockPlay(position *BlockXYZ, instrument InstrumentID, pitch NotePitch) {
    p.printf("PacketNoteBlockPlay(position=%v, instrument=%d, pitch=%d)",
        position, instrument, pitch)
}

func (p *MessageParser) PacketExplosion(position *AbsXYZ, power float32, blockOffsets []proto.ExplosionOffsetXYZ) {
    p.printf("PacketExplosion(position=%v, power=%f, blockOffsets=%v)",
        position, power, blockOffsets)
}

func (p *MessageParser) PacketBedInvalid(field1 byte) {
    p.printf("PacketBedInvalid(field1=%t)", field1)
}

func (p *MessageParser) PacketWindowOpen(windowID WindowID, invTypeID InvTypeID, windowTitle string, numSlots byte) {
    p.printf("PacketWindowOpen(windowID=%d, invTypeID=%d, windowTitle=%q, numSlots=%d)",
        windowID, invTypeID, windowTitle, numSlots)
}

func (p *MessageParser) PacketWindowClose(windowID WindowID) {
    p.printf("PacketWindowClose(windowID=%d)", windowID)
}

func (p *MessageParser) PacketWindowProgressBar(windowID WindowID, prgBarID PrgBarID, value PrgBarValue) {
    p.printf("PacketWindowProgressBar(windowID=%d, prgBarID=%d, value=%d)",
        windowID, prgBarID, value)
}

func (p *MessageParser) PacketWindowTransaction(windowID WindowID, txID TxID, accepted bool) {
    p.printf("PacketWindowTransaction(windowID=%d, txID=%d, accepted=%v)")
}

func (p *MessageParser) PacketWindowClick(windowID WindowID, slot SlotID, rightClick bool, txID TxID, itemID ItemID, amount ItemCount, uses ItemUses) {
    p.printf("PacketWindowClick(windowID=%d, slot=%d, rightClick=%v, txID=%d, itemID=%d, amount=%d, uses=%d)",
        windowID, slot, rightClick, txID, itemID, amount, uses)
}

func (p *MessageParser) PacketWindowSetSlot(windowID WindowID, slot SlotID, itemID ItemID, amount ItemCount, uses ItemUses) {
    p.printf("PacketWindowSetSlot(windowID=%d, slot=%d, itemID=%d, amount=%d, uses=%d)",
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

    serverId, err := proto.ClientReadHandshake(reader)
    if err != nil {
        p.printf("ClientReadHandshake error: %v", err)
        return
    }
    p.printf("ClientReadHandshake(serverId=%v)", serverId)

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

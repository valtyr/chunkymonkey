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

func (p *MessageParser) PacketFlying(flying bool) {
    p.printf("PacketFlying(%v)", flying)
}

func (p *MessageParser) PacketPlayerPosition(position *XYZ, stance AbsoluteCoord, flying bool) {
    p.printf("PacketPlayerPosition(%v, %v, %v)", position, stance, flying)
}

func (p *MessageParser) PacketPlayerLook(orientation *Orientation, flying bool) {
    p.printf("PacketPlayerLook(%v, %v)", orientation, flying)
}

func (p *MessageParser) PacketPlayerDigging(status DigStatus, blockLoc *BlockXYZ, face Face) {
    p.printf("PacketPlayerDigging(%v, %v, %v)", status, blockLoc, face)
}

func (p *MessageParser) PacketPlayerBlockPlacement(blockItemID int16, blockLoc *BlockXYZ, direction Face) {
    p.printf("PacketPlayerBlockPlacement(%d, %v, %v)", blockItemID, blockLoc, direction)
}

func (p *MessageParser) PacketHoldingChange(blockItemID int16) {
    p.printf("PacketHoldingChange(%d)", blockItemID)
}

func (p *MessageParser) PacketArmAnimation(forward bool) {
    p.printf("PacketArmAnimation(%v)", forward)
}

func (p *MessageParser) PacketDisconnect(reason string) {
    p.printf("PacketDisconnect(%s)", reason)
}

func (p *MessageParser) SCPacketLogin(entityID EntityID, str1 string, str2 string, mapSeed int64, dimension byte) {
    p.printf("SCPacketLogin(entityID=%d, str1=%v, str2=%v, mapSeed=%d, dimension=%d)",
        entityID, str1, str2, mapSeed, dimension)
}

func (p *MessageParser) SCPacketTimeUpdate(time int64) {
    p.printf("SCPacketTime(time=%d)", time)
}

func (p *MessageParser) SCPacketSpawnPosition(position *BlockXYZ) {
    p.printf("SCPacketSpawnPosition(position=%v)", position)
}

func (p *MessageParser) SCPacketUseEntity(user EntityID, target EntityID, leftClick bool) {
    p.printf("PacketUseEntity(user=%d, target=%d, leftClick=%v)", user, target, leftClick)
}

func (p *MessageParser) SCPacketUpdateHealth(health int16) {
    p.printf("SCPacketUpdateHealth(health=%d)", health)
}

func (p *MessageParser) SCPacketMobSpawn(entityID EntityID, mobType byte, position *XYZInteger, yaw byte, pitch byte, data []proto.UnknownEntityExtra) {
    p.printf("SCPacketMobSpawn(entityID=%d, mobType=%d, position=%v, yaw=%d, pitch=%d, data=%v)",
        entityID, mobType, position, yaw, pitch, data)
}

func (p *MessageParser) SCPacketUnknownX19(field1 int32, field2 string, field3, field4, field5, field6 int32) {
    p.printf("SCPacketUnknownX19(field1=%d, field2=%v, field3=%d, field4=%d, field5=%d, field6=%d)",
        field1, field2, field3, field4, field5, field6)
}

func (p *MessageParser) SCPacketEntityVelocity(entityID EntityID, x, y, z int16) {
    p.printf("SCPacketEntityVelocity(entityID=%d, x=%d, y=%d, z=%d)",
        entityID, x, y, z)
}

func (p *MessageParser) SCPacketUnknownX28(field1 int32, data []proto.UnknownEntityExtra) {
    p.printf("SCPacketUnknownX28(field1=%d, data=%v)", field1, data)
}

func (p *MessageParser) SCPacketPreChunk(position *ChunkXZ, mode bool) {
    p.printf("SCPacketPreChunk(position=%v, mode=%v)", position, mode)
}

func (p *MessageParser) SCPacketMapChunk(position *BlockXYZ, sizeX, sizeY, sizeZ byte, data []byte) {
    p.printf("SCPacketMapChunk(position=%v, sizeX=%d, sizeY=%d, sizeZ=%d, len(data)=%d)",
        position, sizeX, sizeY, sizeZ, len(data))
}

// Parses messages from the client
func (p *MessageParser) CSParse(reader io.Reader) {
    // If we return, we should consume all input to avoid blocking the pipe
    // we're listening on. TODO Maybe we could just close it?
    defer p.consumeUnrecognizedInput(reader)

    p.logPrefix = "(C->S) "

    defer func() {
        if err := recover(); err != nil {
            p.printf("Parsing failed: %v", err)
        }
    }()

    username, err := proto.CSReadHandshake(reader)
    if err != nil {
        p.printf("CSReadHandshake error: %v", err)
        return
    }
    p.printf("CSReadHandshake(username=%v)", username)

    loginUsername, _, err := proto.CSReadLogin(reader)
    if err != nil {
        p.printf("CSReadLogin error: %v", err)
        return
    }
    p.printf("CSReadLogin(username=%v)", loginUsername)

    for {
        err := proto.CSReadPacket(reader, p)
        if err != nil {
            if err != os.EOF {
                p.printf("ReceiveLoop failed: %v", err)
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

    p.logPrefix = "(S->C) "

    connectionHash, err := proto.SCReadHandshake(reader)
    if err != nil {
        p.printf("SCReadHandshake error: %v", err)
        return
    }
    p.printf("SCReadHandshake connectionHash=%v", connectionHash)

    for {
        err := proto.SCReadPacket(reader, p)
        if err != nil {
            if err != os.EOF {
                p.printf("ReceiveLoop failed: %v", err)
            }
            return
        }
    }
}

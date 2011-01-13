package intercept_parse

import (
    cm  "chunkymonkey/chunkymonkey"
    "encoding/hex"
    "io"
    "log"
    "os"
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

func (p *MessageParser) PacketPlayerPosition(position *cm.XYZ, stance cm.AbsoluteCoord, flying bool) {
    p.printf("PacketPlayerPosition(%v, %v, %v)", position, stance, flying)
}

func (p *MessageParser) PacketPlayerLook(orientation *cm.Orientation, flying bool) {
    p.printf("PacketPlayerLook(%v, %v)", orientation, flying)
}

func (p *MessageParser) PacketPlayerDigging(status cm.DigStatus, blockLoc *cm.BlockXYZ, face cm.Face) {
    p.printf("PacketPlayerDigging(%v, %v, %v)", status, blockLoc, face)
}

func (p *MessageParser) PacketPlayerBlockPlacement(blockItemID int16, blockLoc *cm.BlockXYZ, direction cm.Face) {
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

func (p *MessageParser) SCPacketLogin(entityID cm.EntityID, str1 string, str2 string, mapSeed int64, dimension byte) {
    p.printf("SCPacketLogin(entityID=%d, str1=%v, str2=%v, mapSeed=%d, dimension=%d)",
        entityID, str1, str2, mapSeed, dimension)
}

func (p *MessageParser) SCPacketTimeUpdate(time int64) {
    p.printf("SCPacketTime(time=%d)", time)
}

func (p *MessageParser) SCPacketSpawnPosition(position *cm.BlockXYZ) {
    p.printf("SCPacketSpawnPosition(position=%v)", position)
}

func (p *MessageParser) SCPacketUpdateHealth(health int16) {
    p.printf("SCPacketUpdateHealth(health=%d)", health)
}

func (p *MessageParser) SCPacketMobSpawn(entityID cm.EntityID, mobType byte, position *cm.XYZInteger, yaw byte, pitch byte) {
    p.printf("SCPacketMobSpawn(entityID=%d, mobType=%d, position=%v, yaw=%d, pitch=%d)",
        entityID, mobType, position, yaw, pitch)
}

func (p *MessageParser) SCPacketPreChunk(position *cm.ChunkXZ, mode bool) {
    p.printf("SCPacketPreChunk(position=%v, mode=%v)", position, mode)
}

func (p *MessageParser) SCPacketMapChunk(position *cm.BlockXYZ, sizeX, sizeY, sizeZ byte, data []byte) {
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

    username, err := cm.CSReadHandshake(reader)
    if err != nil {
        p.printf("CSReadHandshake error: %v", err)
        return
    }
    p.printf("CSReadHandshake(username=%v)", username)

    loginUsername, _, err := cm.CSReadLogin(reader)
    if err != nil {
        p.printf("CSReadLogin error: %v", err)
        return
    }
    p.printf("CSReadLogin(username=%v)", loginUsername)

    for {
        err := cm.CSReadPacket(reader, p)
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

    connectionHash, err := cm.SCReadHandshake(reader)
    if err != nil {
        p.printf("SCReadHandshake error: %v", err)
        return
    }
    p.printf("SCReadHandshake connectionHash=%v", connectionHash)

    for {
        err := cm.SCReadPacket(reader, p)
        if err != nil {
            if err != os.EOF {
                p.printf("ReceiveLoop failed: %v", err)
            }
            return
        }
    }
}

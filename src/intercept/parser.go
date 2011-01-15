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

func (p *MessageParser) RecvKeepAlive() {
}

func (p *MessageParser) RecvChatMessage(message string) {
    p.printf("RecvChatMessage(%s)", message)
}

func (p *MessageParser) RecvOnGround(onGround bool) {
    p.printf("RecvOnGround(%v)", onGround)
}

func (p *MessageParser) RecvPlayerPosition(position *XYZ, stance AbsoluteCoord, onGround bool) {
    p.printf("RecvPlayerPosition(%v, %v, %v)", position, stance, onGround)
}

func (p *MessageParser) RecvPlayerLook(orientation *Orientation, onGround bool) {
    p.printf("RecvPlayerLook(%v, %v)", orientation, onGround)
}

func (p *MessageParser) RecvPlayerDigging(status DigStatus, blockLoc *BlockXYZ, face Face) {
    p.printf("RecvPlayerDigging(%v, %v, %v)", status, blockLoc, face)
}

func (p *MessageParser) RecvPlayerBlockPlacement(blockItemID int16, blockLoc *BlockXYZ, direction Face) {
    p.printf("RecvPlayerBlockPlacement(%d, %v, %v)", blockItemID, blockLoc, direction)
}

func (p *MessageParser) RecvHoldingChange(blockItemID int16) {
    p.printf("RecvHoldingChange(%d)", blockItemID)
}

func (p *MessageParser) RecvArmAnimation(forward bool) {
    p.printf("RecvArmAnimation(%v)", forward)
}

func (p *MessageParser) RecvDisconnect(reason string) {
    p.printf("RecvDisconnect(%s)", reason)
}

func (p *MessageParser) ClientRecvLogin(entityID EntityID, str1 string, str2 string, mapSeed int64, dimension byte) {
    p.printf("ClientRecvLogin(entityID=%d, str1=%v, str2=%v, mapSeed=%d, dimension=%d)",
        entityID, str1, str2, mapSeed, dimension)
}

func (p *MessageParser) ClientRecvTimeUpdate(time int64) {
    p.printf("ClientRecvTime(time=%d)", time)
}

func (p *MessageParser) ClientRecvSpawnPosition(position *BlockXYZ) {
    p.printf("ClientRecvSpawnPosition(position=%v)", position)
}

func (p *MessageParser) ClientRecvUseEntity(user EntityID, target EntityID, leftClick bool) {
    p.printf("ClientRecvUseEntity(user=%d, target=%d, leftClick=%v)", user, target, leftClick)
}

func (p *MessageParser) ClientRecvUpdateHealth(health int16) {
    p.printf("ClientRecvUpdateHealth(health=%d)", health)
}

func (p *MessageParser) ClientRecvMobSpawn(entityID EntityID, mobType byte, position *XYZInteger, yaw byte, pitch byte, data []proto.UnknownEntityExtra) {
    p.printf("ClientRecvMobSpawn(entityID=%d, mobType=%d, position=%v, yaw=%d, pitch=%d, data=%v)",
        entityID, mobType, position, yaw, pitch, data)
}

func (p *MessageParser) ClientRecvUnknownX19(field1 int32, field2 string, field3, field4, field5, field6 int32) {
    p.printf("ClientRecvUnknownX19(field1=%d, field2=%v, field3=%d, field4=%d, field5=%d, field6=%d)",
        field1, field2, field3, field4, field5, field6)
}

func (p *MessageParser) ClientRecvEntityVelocity(entityID EntityID, x, y, z int16) {
    p.printf("ClientRecvEntityVelocity(entityID=%d, x=%d, y=%d, z=%d)",
        entityID, x, y, z)
}

func (p *MessageParser) ClientRecvUnknownX28(field1 int32, data []proto.UnknownEntityExtra) {
    p.printf("ClientRecvUnknownX28(field1=%d, data=%v)", field1, data)
}

func (p *MessageParser) ClientRecvPreChunk(position *ChunkXZ, mode bool) {
    p.printf("ClientRecvPreChunk(position=%v, mode=%v)", position, mode)
}

func (p *MessageParser) ClientRecvMapChunk(position *BlockXYZ, sizeX, sizeY, sizeZ byte, data []byte) {
    p.printf("ClientRecvMapChunk(position=%v, sizeX=%d, sizeY=%d, sizeZ=%d, len(data)=%d)",
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

    username, err := proto.ServerReadHandshake(reader)
    if err != nil {
        p.printf("serverReadHandshake error: %v", err)
        return
    }
    p.printf("serverReadHandshake(username=%v)", username)

    loginUsername, _, err := proto.ServerReadLogin(reader)
    if err != nil {
        p.printf("serverReadLogin error: %v", err)
        return
    }
    p.printf("serverReadLogin(username=%v)", loginUsername)

    for {
        err := proto.ServerReadPacket(reader, p)
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

    connectionHash, err := proto.ClientReadHandshake(reader)
    if err != nil {
        p.printf("clientReadHandshake error: %v", err)
        return
    }
    p.printf("clientReadHandshake connectionHash=%v", connectionHash)

    for {
        err := proto.ClientReadPacket(reader, p)
        if err != nil {
            if err != os.EOF {
                p.printf("ReceiveLoop failed: %v", err)
            }
            return
        }
    }
}

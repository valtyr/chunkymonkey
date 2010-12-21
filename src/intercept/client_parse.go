package intercept_parse

import (
    cm "chunkymonkey/chunkymonkey"
    "io"
    "log"
    "os"
)

type ClientMessageParser struct {}

func (p *ClientMessageParser) PacketKeepAlive() {
}

func (p *ClientMessageParser) PacketChatMessage(message string) {
    log.Printf("(C->S) PacketChatMessage(%s)", message)
}

func (p *ClientMessageParser) PacketFlying(flying bool) {
    log.Printf("(C->S) PacketFlying(%v)", flying)
}

func (p *ClientMessageParser) PacketPlayerPosition(position *cm.XYZ, stance cm.AbsoluteCoord, flying bool) {
    log.Printf("(C->S) PacketPlayerPosition(%v, %v, %v)", position, stance, flying)
}

func (p *ClientMessageParser) PacketPlayerLook(orientation *cm.Orientation, flying bool) {
    log.Printf("(C->S) PacketPlayerLook(%v, %v)", orientation, flying)
}

func (p *ClientMessageParser) PacketPlayerDigging(status cm.DigStatus, blockLoc *cm.BlockXYZ, face cm.Face) {
    log.Printf("(C->S) PacketPlayerDigging(%v, %v, %v)", status, blockLoc, face)
}

func (p *ClientMessageParser) PacketPlayerBlockPlacement(blockItemID int16, blockLoc *cm.BlockXYZ, direction cm.Face) {
    log.Printf("(C->S) PacketPlayerBlockPlacement(%d, %v, %v)", blockItemID, blockLoc, direction)
}

func (p *ClientMessageParser) PacketHoldingChange(blockItemID int16) {
    log.Printf("(C->S) PacketHoldingChange(%d)", blockItemID)
}

func (p *ClientMessageParser) PacketArmAnimation(forward bool) {
    log.Printf("(C->S) PacketArmAnimation(%v)", forward)
}

func (p *ClientMessageParser) PacketDisconnect(reason string) {
    log.Printf("(C->S) PacketDisconnect(%s)", reason)
}

// Parses messages from the client
func (p *ClientMessageParser) Parse(reader io.Reader) {
    // If we return, we should consume all input to avoid blocking the pipe
    // we're listening on. TODO Maybe we could just close it?
    defer consumeUnrecognizedInput(reader)

    defer func() {
        if err := recover(); err != nil {
            log.Printf("Parsing failed: %v", err)
        }
    }()

    username, err := cm.ReadHandshake(reader)
    if err != nil {
        log.Printf("(C->S) ReadHandshake error: %v", err)
        return
    }
    log.Printf("(C->S) ReadHandshake username=%v", username)

    loginUsername, _, err := cm.ReadLogin(reader)
    if err != nil {
        log.Print("(C->S) ReadLogin error: %v", err)
        return
    }
    log.Printf("(C->S) ReadLogin username=%v", loginUsername)

    for {
        err := cm.ReadPacket(reader, p)
        if err != nil {
            if err != os.EOF {
                log.Printf("(C->S) ReceiveLoop failed: %v", err)
            }
            return
        }
    }
}

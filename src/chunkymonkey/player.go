package main

import (
    "os"
    "io"
    "log"
    "net"
    "math"
    "bytes"
)

type Player struct {
    Entity
    game        *Game
    conn        net.Conn
    name        string
    position    XYZ
    orientation Orientation
    currentItem int16
    txQueue     chan []byte
}

const StanceNormal = 1.62

func StartPlayer(game *Game, conn net.Conn, name string) {
    player := &Player{
        game:        game,
        conn:        conn,
        name:        name,
        position:    StartPosition,
        orientation: Orientation{0, 0},
        txQueue:     make(chan []byte, 128),
    }

    game.Enqueue(func(game *Game) {
        game.AddPlayer(player)
        WriteLogin(conn, player.Entity.EntityID)
        player.start()
        player.postLogin()
    })
}

func (player *Player) start() {
    go player.ReceiveLoop()
    go player.TransmitLoop()
}

func (player *Player) PacketKeepAlive() {
}

func (player *Player) PacketChatMessage(message string) {
    log.Printf("PacketChatMessage message=%s", message)

    player.game.Enqueue(func(game *Game) { game.SendChatMessage(message) })
}

func (player *Player) PacketFlying(flying bool) {
}

func (player *Player) PacketPlayerPosition(position *XYZ, stance AbsoluteCoord, flying bool) {
    // TODO: Should keep track of when players enter/leave their mutual radius
    // of "awareness". I.e a client should receive a RemoveEntity packet when
    // the player walks out of range, and no longer receive WriteEntityTeleport
    // packets for them. The converse should happen when players come in range
    // of each other.

    player.game.Enqueue(func(game *Game) {
        var delta = XYZ{position.x - player.position.x,
            position.y - player.position.y,
            position.z - player.position.z}
        distance := math.Sqrt(float64(delta.x*delta.x + delta.y*delta.y + delta.z*delta.z))
        if distance > 10 {
            log.Printf("Discarding player position that is too far removed (%.2f, %.2f, %.2f)",
                position.x, position.y, position.z)
            return
        }

        player.position = *position

        buf := &bytes.Buffer{}
        WriteEntityTeleport(buf, player.EntityID, &player.position, &player.orientation)
        game.MulticastPacket(buf.Bytes(), player)
    })
}

func (player *Player) PacketPlayerLook(orientation *Orientation, flying bool) {
    player.game.Enqueue(func(game *Game) {
        // TODO input validation
        player.orientation = *orientation

        buf := &bytes.Buffer{}
        WriteEntityLook(buf, player.EntityID, orientation)
        game.MulticastPacket(buf.Bytes(), player)
    })
}

func (player *Player) PacketPlayerDigging(status DigStatus, blockLoc *BlockXYZ, face Face) {
    // TODO validate that the player is actually somewhere near the block

    if status == DigBlockBroke {
        // TODO validate that the player has dug long enough to stop speed
        // hacking (based on block type and tool used - non-trivial).

        player.game.Enqueue(func(game *Game) {
            chunkLoc, subLoc := blockLoc.ToChunkLocal()

            chunk := game.chunkManager.Get(chunkLoc)

            if chunk == nil {
                return
            }

            if !chunk.SetBlock(&subLoc, BlockIDAir, 0) {
                packet := &bytes.Buffer{}
                WriteBlockChange(packet, blockLoc, BlockIDAir, 0)
                game.MulticastChunkPacket(packet.Bytes(), chunkLoc)
            }
        })
    }
}

func (player *Player) PacketPlayerBlockPlacement(blockItemID int16, blockLoc *BlockXYZ, direction Face) {
    log.Printf("PacketPlayerBlockPlacement blockItemID=%d blockLoc=%v direction=%d",
        blockItemID, *blockLoc, direction)
}

func (player *Player) PacketHoldingChange(blockItemID int16) {
    log.Printf("PacketHoldingChange blockItemID=%d", blockItemID)
}

func (player *Player) PacketArmAnimation(forward bool) {
}

func (player *Player) PacketDisconnect(reason string) {
    log.Printf("PacketDisconnect reason=%s", reason)
    player.game.Enqueue(func(game *Game) {
        game.RemovePlayer(player)
        close(player.txQueue)
        player.conn.Close()
    })
}

func (player *Player) ReceiveLoop() {
    for {
        err := ReadPacket(player.conn, player)
        if err != nil {
            if err != os.EOF {
                log.Print("ReceiveLoop failed: ", err.String())
            }
            return
        }
    }
}

func (player *Player) TransmitLoop() {
    for {
        bs := <-player.txQueue
        if bs == nil {
            return // txQueue closed
        }

        _, err := player.conn.Write(bs)
        if err != nil {
            if err != os.EOF {
                log.Print("TransmitLoop failed: ", err.String())
            }
            return
        }
    }
}

func (player *Player) sendChunks(writer io.Writer) {
    playerChunkLoc := player.position.ToChunkXZ()

    for chunk := range player.game.chunkManager.ChunksInRadius(&playerChunkLoc) {
        WritePreChunk(writer, &chunk.XZ, true)
    }

    for chunk := range player.game.chunkManager.ChunksInRadius(&playerChunkLoc) {
        WriteMapChunk(writer, chunk)
    }
}

func (player *Player) TransmitPacket(packet []byte) {
    if packet == nil {
        return // skip empty packets
    }
    player.txQueue <- packet
}

func (player *Player) postLogin() {
    buf := &bytes.Buffer{}
    WriteSpawnPosition(buf, &player.position)
    player.sendChunks(buf)
    WritePlayerInventory(buf)
    WritePlayerPositionLook(buf, &player.position, &player.orientation,
        player.position.y+StanceNormal, false)
    player.TransmitPacket(buf.Bytes())
}

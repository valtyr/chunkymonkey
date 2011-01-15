package chunkymonkey

import (
    "os"
    "io"
    "log"
    "net"
    "math"
    "bytes"

    "chunkymonkey/proto"
    .   "chunkymonkey/types"
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
        proto.ServerWriteLogin(conn, player.Entity.EntityID)
        player.start()
        player.postLogin()
    })
}

func (player *Player) start() {
    go player.ReceiveLoop()
    go player.TransmitLoop()
}

func (player *Player) RecvKeepAlive() {
}

func (player *Player) RecvChatMessage(message string) {
    log.Printf("RecvChatMessage message=%s", message)

    player.game.Enqueue(func(game *Game) { game.SendChatMessage(message) })
}

func (player *Player) RecvOnGround(onGround bool) {
}

func (player *Player) RecvPlayerPosition(position *XYZ, stance AbsoluteCoord, onGround bool) {
    // TODO: Should keep track of when players enter/leave their mutual radius
    // of "awareness". I.e a client should receive a RemoveEntity packet when
    // the player walks out of range, and no longer receive WriteEntityTeleport
    // packets for them. The converse should happen when players come in range
    // of each other.

    player.game.Enqueue(func(game *Game) {
        var delta = XYZ{position.X - player.position.X,
            position.Y - player.position.Y,
            position.Z - player.position.Z}
        distance := math.Sqrt(float64(delta.X*delta.X + delta.Y*delta.Y + delta.Z*delta.Z))
        if distance > 10 {
            log.Printf("Discarding player position that is too far removed (%.2f, %.2f, %.2f)",
                position.X, position.Y, position.Z)
            return
        }

        player.position = *position

        buf := &bytes.Buffer{}
        proto.WriteEntityTeleport(buf, player.EntityID, &player.position, &player.orientation)
        game.MulticastPacket(buf.Bytes(), player)
    })
}

func (player *Player) RecvPlayerLook(orientation *Orientation, onGround bool) {
    player.game.Enqueue(func(game *Game) {
        // TODO input validation
        player.orientation = *orientation

        buf := &bytes.Buffer{}
        proto.WriteEntityLook(buf, player.EntityID, orientation)
        game.MulticastPacket(buf.Bytes(), player)
    })
}

func (player *Player) RecvPlayerDigging(status DigStatus, blockLoc *BlockXYZ, face Face) {
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

            blockType, err := chunk.GetBlock(&subLoc)
            if err {
                return
            }

            if !chunk.SetBlock(&subLoc, BlockIDAir, 0) {
                // Experimental code - we spawn earth blocks if earth/grass was dug out
                if blockType == BlockIDDirt || blockType == BlockIDGrass {
                    // TODO model the item's fall to the ground. Do we need
                    // update clients as to its final position?
                    NewPickupItem(game, ItemID(BlockIDDirt), 1, blockLoc.ToXYZInteger())
                }
            }
        })
    }
}

func (player *Player) RecvPlayerBlockPlacement(blockItemID int16, blockLoc *BlockXYZ, direction Face) {
    log.Printf("RecvPlayerBlockPlacement blockItemID=%d blockLoc=%v direction=%d",
        blockItemID, *blockLoc, direction)
}

func (player *Player) RecvHoldingChange(blockItemID int16) {
    log.Printf("RecvHoldingChange blockItemID=%d", blockItemID)
}

func (player *Player) RecvArmAnimation(forward bool) {
}

func (player *Player) RecvDisconnect(reason string) {
    log.Printf("RecvDisconnect reason=%s", reason)
    player.game.Enqueue(func(game *Game) {
        game.RemovePlayer(player)
        close(player.txQueue)
        player.conn.Close()
    })
}

func (player *Player) ReceiveLoop() {
    for {
        err := proto.ServerReadPacket(player.conn, player)
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
        proto.WritePreChunk(writer, &chunk.XZ, true)
    }

    for chunk := range player.game.chunkManager.ChunksInRadius(&playerChunkLoc) {
        chunk.SendChunkData(writer)
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
    proto.WriteSpawnPosition(buf, &player.position)
    player.sendChunks(buf)
    proto.WritePlayerInventory(buf)
    proto.ServerWritePlayerPositionLook(buf, &player.position, &player.orientation,
        player.position.Y+StanceNormal, false)
    player.TransmitPacket(buf.Bytes())
}

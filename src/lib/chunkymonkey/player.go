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
    position    AbsXYZ
    look        LookDegrees
    currentItem ItemID
    txQueue     chan []byte
}

const StanceNormal = 1.62

func StartPlayer(game *Game, conn net.Conn, name string) {
    player := &Player{
        game:     game,
        conn:     conn,
        name:     name,
        position: StartPosition,
        look:     LookDegrees{0, 0},
        txQueue:  make(chan []byte, 128),
    }

    game.Enqueue(func(game *Game) {
        game.AddPlayer(player)
        // TODO pass proper map seed and dimension
        proto.ServerWriteLogin(conn, player.Entity.EntityID, 0, DimensionNormal)
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
    player.game.Enqueue(func(game *Game) { game.SendChatMessage(message) })
}

func (player *Player) PacketEntityAction(entityID EntityID, action EntityAction) {
}

func (player *Player) PacketUseEntity(user EntityID, target EntityID, leftClick bool) {
}

func (player *Player) PacketRespawn() {
}

func (player *Player) PacketPlayer(onGround bool) {
}

func (player *Player) PacketPlayerPosition(position *AbsXYZ, stance AbsCoord, onGround bool) {
    // TODO: Should keep track of when players enter/leave their mutual radius
    // of "awareness". I.e a client should receive a RemoveEntity packet when
    // the player walks out of range, and no longer receive WriteEntityTeleport
    // packets for them. The converse should happen when players come in range
    // of each other.

    player.game.Enqueue(func(game *Game) {
        var delta = AbsXYZ{position.X - player.position.X,
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
        proto.WriteEntityTeleport(
            buf,
            player.EntityID,
            player.position.ToAbsIntXYZ(),
            player.look.ToLookBytes())
        game.MulticastPacket(buf.Bytes(), player)
    })
}

func (player *Player) PacketPlayerLook(look *LookDegrees, onGround bool) {
    player.game.Enqueue(func(game *Game) {
        // TODO input validation
        player.look = *look

        buf := &bytes.Buffer{}
        proto.WriteEntityLook(buf, player.EntityID, look.ToLookBytes())
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

            chunk.DigBlock(subLoc)
        })
    }
}

func (player *Player) PacketPlayerBlockPlacement(itemID ItemID, blockLoc *BlockXYZ, face Face, amount ItemCount, uses ItemUses) {
}

func (player *Player) PacketHoldingChange(itemID ItemID) {
}

func (player *Player) PacketEntityAnimation(entityID EntityID, animation EntityAnimation) {
}

func (player *Player) PacketWindowClose(windowID WindowID) {
}

func (player *Player) PacketWindowClick(windowID WindowID, slot SlotID, rightClick bool, txID TxID, itemID ItemID, amount ItemCount, uses ItemUses) {
}

func (player *Player) PacketSignUpdate(position *BlockXYZ, lines [4]string) {
}

func (player *Player) PacketDisconnect(reason string) {
    log.Printf("Player %s disconnected reason=%s", player.name, reason)
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

    for chunk := range player.game.chunkManager.ChunksInRadius(playerChunkLoc) {
        proto.WritePreChunk(writer, &chunk.XZ, ChunkInit)
    }

    for chunk := range player.game.chunkManager.ChunksInRadius(playerChunkLoc) {
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
    proto.WriteSpawnPosition(buf, player.position.ToBlockXYZ())
    player.sendChunks(buf)
    // TODO put in required packets for client connection
    proto.ServerWritePlayerPositionLook(buf, &player.position, &player.look,
        player.position.Y+StanceNormal, false)
    player.TransmitPacket(buf.Bytes())
}

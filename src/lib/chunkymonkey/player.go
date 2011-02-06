package player

import (
    "bytes"
    "expvar"
    "log"
    "io"
    "math"
    "net"
    "os"

    .   "chunkymonkey/entity"
    .   "chunkymonkey/interfaces"
    "chunkymonkey/proto"
    .   "chunkymonkey/types"
)

var (
    expVarPlayerConnectionCount    *expvar.Int
    expVarPlayerDisconnectionCount *expvar.Int
)

func init() {
    expVarPlayerConnectionCount = expvar.NewInt("player-connection-count")
    expVarPlayerDisconnectionCount = expvar.NewInt("player-disconnection-count")
}

type Player struct {
    Entity
    game        IGame
    conn        net.Conn
    name        string
    position    AbsXYZ
    look        LookDegrees
    currentItem ItemID

    mainQueue   chan func(IPlayer)
    txQueue     chan []byte
}

const StanceNormal = 1.62

func StartPlayer(game IGame, conn net.Conn, name string) {
    player := &Player{
        game:      game,
        conn:      conn,
        name:      name,
        position:  *game.GetStartPosition(),
        look:      LookDegrees{0, 0},
        mainQueue: make(chan func(IPlayer), 128),
        txQueue:   make(chan []byte, 128),
    }

    game.Enqueue(func(game IGame) {
        game.AddPlayer(player)
        // TODO pass proper map seed and dimension
        proto.ServerWriteLogin(conn, player.Entity.EntityID, 0, DimensionNormal)
        player.start()
        player.postLogin()
    })
}

func (player *Player) GetEntity() *Entity {
    return &player.Entity
}

func (player *Player) GetChunkPosition() *ChunkXZ {
    return player.position.ToChunkXZ()
}

func (player *Player) GetName() string {
    return player.name
}

func (player *Player) Enqueue(f func(IPlayer)) {
    player.mainQueue <- f
}

func (player *Player) SendSpawn(writer io.Writer) (err os.Error) {
    return proto.WriteNamedEntitySpawn(
        writer,
        player.Entity.EntityID, player.name,
        player.position.ToAbsIntXYZ(),
        player.look.ToLookBytes(),
        player.currentItem,
    )
}

func (player *Player) start() {
    expVarPlayerConnectionCount.Add(1)
    go player.receiveLoop()
    go player.mainLoop()
}

func (player *Player) PacketKeepAlive() {
}

func (player *Player) PacketChatMessage(message string) {
    player.game.Enqueue(func(game IGame) { game.SendChatMessage(message) })
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

    player.game.Enqueue(func(game IGame) {
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
    player.game.Enqueue(func(game IGame) {
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

        player.game.Enqueue(func(game IGame) {
            chunkLoc, subLoc := blockLoc.ToChunkLocal()

            chunk := game.GetChunkManager().Get(chunkLoc)

            if chunk == nil {
                return
            }

            chunk.Enqueue(func(chunk IChunk) {
                chunk.DestroyBlock(subLoc)
            })
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
    player.game.Enqueue(func(game IGame) {
        game.RemovePlayer(player)
        close(player.txQueue)
        player.conn.Close()
    })
}

func (player *Player) receiveLoop() {
    for {
        err := proto.ServerReadPacket(player.conn, player)
        if err != nil {
            if err != os.EOF {
                log.Print("ReceiveLoop failed: ", err.String())
            }
            expVarPlayerDisconnectionCount.Add(1)
            return
        }
    }
}

func (player *Player) mainLoop() {
    for {
        select {
        case f := <-player.mainQueue:
            f(player)
        case bs := <-player.txQueue:
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
}

func (player *Player) TransmitPacket(packet []byte) {
    if packet == nil {
        return // skip empty packets
    }
    player.txQueue <- packet
}

// Blocks until all chunks have been transmitted. Note that this must currently
// be called from within the game loop
// TODO reduce the overhead placed on the game loop to a minimum
func (player *Player) sendChunks() {
    // TODO more optimal chunk loading algorithm. Chunks near the player should
    // be sent first.

    playerChunkLoc := player.position.ToChunkXZ()

    finish := make(chan bool, ChunkRadius * ChunkRadius)
    buf := &bytes.Buffer{}
    for chunk := range player.game.GetChunkManager().ChunksInRadius(playerChunkLoc) {
        proto.WritePreChunk(buf, chunk.GetLoc(), ChunkInit)
    }
    player.TransmitPacket(buf.Bytes())

    chunkCount := 0
    for chunk := range player.game.GetChunkManager().ChunksInRadius(playerChunkLoc) {
        chunkCount++
        chunk.Enqueue(func(chunk IChunk) {
            buf := &bytes.Buffer{}
            chunk.SendChunkData(buf)
            player.TransmitPacket(buf.Bytes())
            finish <- true
        })
    }

    // Wait for all chunks to have been sent
    for ; chunkCount > 0; chunkCount-- {
        _ = <-finish
    }
}

// Blocks until all chunks have been transmitted. Note that this must currently
// be called from within the game loop
// TODO reduce the overhead placed on the game loop to a minimum
func (player *Player) postLogin() {
    player.sendChunks()

    buf := &bytes.Buffer{}
    proto.WriteSpawnPosition(buf, player.position.ToBlockXYZ())
    proto.ServerWritePlayerPositionLook(buf, &player.position, &player.look,
        player.position.Y+StanceNormal, false)
    player.TransmitPacket(buf.Bytes())
}

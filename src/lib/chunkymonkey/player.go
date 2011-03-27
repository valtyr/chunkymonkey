package player

import (
    "bytes"
    "expvar"
    "log"
    "io"
    "math"
    "net"
    "os"
    "sync"

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

    mainQueue chan func(IPlayer)
    txQueue   chan []byte
    lock      sync.Mutex
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
        player.start()
    })
}

func (player *Player) GetEntity() *Entity {
    return &player.Entity
}

func (player *Player) LockedGetChunkPosition() *ChunkXZ {
    player.lock.Lock()
    defer player.lock.Unlock()
    return player.position.ToChunkXZ()
}

func (player *Player) IsWithin(p1, p2 *ChunkXZ) bool {
    p := player.position.ToChunkXZ()
    return (p.X >= p1.X && p.X <= p2.X &&
        p.Z >= p1.Z && p.Z <= p2.Z)
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

// Start of packet handling code
// Note: any packet handlers that could change the player state or read a
// changeable state must use player.lock

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
    player.lock.Lock()
    defer player.lock.Unlock()

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

    // TODO: Should keep track of when players enter/leave their mutual radius
    // of "awareness". I.e a client should receive a RemoveEntity packet when
    // the player walks out of range, and no longer receive WriteEntityTeleport
    // packets for them. The converse should happen when players come in range
    // of each other.

    buf := &bytes.Buffer{}
    proto.WriteEntityTeleport(
        buf,
        player.EntityID,
        player.position.ToAbsIntXYZ(),
        player.look.ToLookBytes())

    player.game.Enqueue(func(game IGame) {
        game.MulticastPacket(buf.Bytes(), player)
    })
}

func (player *Player) PacketPlayerLook(look *LookDegrees, onGround bool) {
    player.lock.Lock()
    defer player.lock.Unlock()

    // TODO input validation
    player.look = *look

    buf := &bytes.Buffer{}
    proto.WriteEntityLook(buf, player.EntityID, look.ToLookBytes())

    player.game.Enqueue(func(game IGame) {
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

func (player *Player) PacketUnknown0x1b(field1, field2, field3, field4 float32, field5, field6 bool) {
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

// End of packet handling code

func (player *Player) runQueuedCall(f func(IPlayer)) {
    player.lock.Lock()
    defer player.lock.Unlock()
    f(player)
}

func (player *Player) mainLoop() {
    go postLogin(player, player.position.Copy())

    for {
        select {
        case f := <-player.mainQueue:
            player.runQueuedCall(f)
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

func chunkOrder(radius ChunkCoord, center *ChunkXZ) (locs []ChunkXZ) {
    areaEdgeSize := radius*2 + 1
    locs = make([]ChunkXZ, areaEdgeSize*areaEdgeSize)
    locs[0] = *center
    index := 1
    for curRadius := ChunkCoord(1); curRadius <= radius; curRadius++ {
        xMin := ChunkCoord(-curRadius + center.X)
        xMax := ChunkCoord(curRadius + center.X)
        zMin := ChunkCoord(-curRadius + center.Z)
        zMax := ChunkCoord(curRadius + center.Z)

        // Northern and southern rows of chunks.
        for x := xMin; x <= xMax; x++ {
            locs[index] = ChunkXZ{x, zMin}
            index++
            locs[index] = ChunkXZ{x, zMax}
            index++
        }

        // Eastern and western columns (except for where they intersect the
        // north and south rows).
        for z := zMin + 1; z < zMax; z++ {
            locs[index] = ChunkXZ{xMin, z}
            index++
            locs[index] = ChunkXZ{xMax, z}
            index++
        }
    }
    return
}

// Blocks until essential login packets have been transmitted.
func postLogin(player *Player, playerPos *AbsXYZ) {
    // TODO pass proper dimension
    buf := &bytes.Buffer{}
    proto.ServerWriteLogin(buf, player.Entity.EntityID, 0, DimensionNormal)
    player.TransmitPacket(buf.Bytes())

    playerChunkLoc := playerPos.ToChunkXZ()

    // Work out order to send chunks in. (Nearest to player first)
    chunkLocOrder := chunkOrder(ChunkRadius, playerChunkLoc)

    // Get all the chunks together that we're going to send.
    chunksChan := make(chan []IChunk)
    player.game.Enqueue(func(game IGame) {
        chunks := make([]IChunk, 0, len(chunkLocOrder))
        mgr := player.game.GetChunkManager()
        for _, chunkLoc := range chunkLocOrder {
            chunk := mgr.Get(&chunkLoc)
            if chunk != nil {
                chunks = append(chunks, chunk)
            }
        }
        chunksChan <- chunks
    })
    chunks := <-chunksChan

    // Warn client to allocate memory for the chunks we're sending.
    buf = &bytes.Buffer{}
    for _, chunk := range chunks {
        proto.WritePreChunk(buf, chunk.GetLoc(), ChunkInit)
    }
    player.TransmitPacket(buf.Bytes())

    // Send the important chunks.
    chunksSent := 0
    waitChunks := &sync.WaitGroup{}
    for _, chunk := range chunks {
        chunkLoc := chunk.GetLoc()
        dx := (chunkLoc.X - playerChunkLoc.X).Abs()
        dz := (chunkLoc.Z - playerChunkLoc.Z).Abs()
        if dx > MinChunkRadius || dz > MinChunkRadius {
            // This chunk isn't important, wait until after other login data
            // has been send before sending this and chunks following.
            break
        }
        chunksSent++
        chunk.Enqueue(func(chunk IChunk) {
            player.TransmitPacket(chunk.ChunkPacket())
            waitChunks.Done()
        })
    }
    waitChunks.Add(chunksSent)
    waitChunks.Wait()

    // Send player start position etc.
    buf = &bytes.Buffer{}
    proto.WriteSpawnPosition(player.conn, player.position.ToBlockXYZ())
    proto.ServerWritePlayerPositionLook(player.conn, &player.position, &player.look,
        player.position.Y+StanceNormal, false)
    player.TransmitPacket(buf.Bytes())

    // Send the remaining chunks
    for _, chunk := range chunks[chunksSent:] {
        chunk.Enqueue(func(chunk IChunk) {
            player.TransmitPacket(chunk.ChunkPacket())
        })
    }
}

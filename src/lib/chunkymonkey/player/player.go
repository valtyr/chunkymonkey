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
    "chunkymonkey/inventory"
    "chunkymonkey/proto"
    "chunkymonkey/slot"
    .   "chunkymonkey/types"
)

var (
    expVarPlayerConnectionCount    *expvar.Int
    expVarPlayerDisconnectionCount *expvar.Int
)

const StanceNormal = 1.62

func init() {
    expVarPlayerConnectionCount = expvar.NewInt("player-connection-count")
    expVarPlayerDisconnectionCount = expvar.NewInt("player-disconnection-count")
}

type Player struct {
    Entity
    game      IGame
    conn      net.Conn
    name      string
    position  AbsXyz
    look      LookDegrees
    chunkSubs chunkSubscriptions

    cursor    slot.Slot // Item being moved by mouse cursor.
    inventory inventory.PlayerInventory

    mainQueue chan func(IPlayer)
    txQueue   chan []byte
    lock      sync.Mutex
}

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

    player.chunkSubs.Init(player)

    player.cursor.Init()
    player.inventory.Init()

    game.Enqueue(func(game IGame) {
        game.AddPlayer(player)
        // TODO pass proper dimension
        buf := &bytes.Buffer{}
        proto.ServerWriteLogin(buf, player.EntityId, 0, DimensionNormal)
        proto.WriteSpawnPosition(buf, player.position.ToBlockXyz())
        player.TransmitPacket(buf.Bytes())
        player.start()
    })
}

func (player *Player) GetEntityId() EntityId {
    return player.EntityId
}

func (player *Player) GetEntity() *Entity {
    return &player.Entity
}

func (player *Player) LockedGetChunkPosition() *ChunkXz {
    player.lock.Lock()
    defer player.lock.Unlock()
    return player.position.ToChunkXz()
}

func (player *Player) IsWithin(p1, p2 *ChunkXz) bool {
    p := player.position.ToChunkXz()
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
    err = proto.WriteNamedEntitySpawn(
        writer,
        player.EntityId, player.name,
        player.position.ToAbsIntXyz(),
        player.look.ToLookBytes(),
        player.inventory.HeldItem().ItemType,
    )
    if err != nil {
        return
    }
    return player.inventory.SendFullEquipmentUpdate(writer, player.EntityId)
}

func (player *Player) start() {
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

func (player *Player) PacketEntityAction(entityId EntityId, action EntityAction) {
}

func (player *Player) PacketUseEntity(user EntityId, target EntityId, leftClick bool) {
}

func (player *Player) PacketRespawn() {
}

func (player *Player) PacketPlayer(onGround bool) {
}

func (player *Player) PacketPlayerPosition(position *AbsXyz, stance AbsCoord, onGround bool) {
    player.lock.Lock()
    defer player.lock.Unlock()

    var delta = AbsXyz{position.X - player.position.X,
        position.Y - player.position.Y,
        position.Z - player.position.Z}
    distance := math.Sqrt(float64(delta.X*delta.X + delta.Y*delta.Y + delta.Z*delta.Z))
    if distance > 10 {
        log.Printf("Discarding player position that is too far removed (%.2f, %.2f, %.2f)",
            position.X, position.Y, position.Z)
        return
    }
    player.position = *position
    player.chunkSubs.Move(position, nil)

    // TODO: Should keep track of when players enter/leave their mutual radius
    // of "awareness". I.e a client should receive a RemoveEntity packet when
    // the player walks out of range, and no longer receive WriteEntityTeleport
    // packets for them. The converse should happen when players come in range
    // of each other.

    buf := &bytes.Buffer{}
    proto.WriteEntityTeleport(
        buf,
        player.EntityId,
        player.position.ToAbsIntXyz(),
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
    proto.WriteEntityLook(buf, player.EntityId, look.ToLookBytes())

    player.game.Enqueue(func(game IGame) {
        game.MulticastPacket(buf.Bytes(), player)
    })
}

func (player *Player) PacketPlayerDigging(status DigStatus, blockLoc *BlockXyz, face Face) {
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

func (player *Player) PacketPlayerBlockPlacement(itemId ItemId, blockLoc *BlockXyz, face Face, amount ItemCount, uses ItemData) {
}

func (player *Player) PacketHoldingChange(slotId SlotId) {
    player.lock.Lock()
    defer player.lock.Unlock()
    player.inventory.SetHolding(slotId)
}

func (player *Player) PacketEntityAnimation(entityId EntityId, animation EntityAnimation) {
}

func (player *Player) PacketUnknown0x1b(field1, field2, field3, field4 float32, field5, field6 bool) {
}

func (player *Player) PacketWindowClose(windowId WindowId) {
}

func (player *Player) PacketWindowClick(windowId WindowId, slot SlotId, rightClick bool, txId TxId, itemId ItemId, amount ItemCount, uses ItemData) {
}

func (player *Player) PacketSignUpdate(position *BlockXyz, lines [4]string) {
}

func (player *Player) PacketDisconnect(reason string) {
    log.Printf("Player %s disconnected reason=%s", player.name, reason)
    player.game.Enqueue(func(game IGame) {
        game.RemovePlayer(player)
        player.txQueue <- nil
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
    expVarPlayerConnectionCount.Add(1)
    defer func() {
        expVarPlayerDisconnectionCount.Add(1)
        player.chunkSubs.clear()
    }()

    player.postLogin()

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

// Used to receive items picked up from chunks.
func (player *Player) OfferItem(item IItem) (taken bool) {
    player.lock.Lock()
    defer player.lock.Unlock()

    buf := &bytes.Buffer{}

    slotChanged := func(slotId SlotId, slot *slot.Slot) {
        slot.SendUpdate(buf, WindowIdInventory, slotId)
    }

    taken = player.inventory.PutItem(item, slotChanged)

    if taken {
        // Send the player message(s) to update the inventory.
        player.TransmitPacket(buf.Bytes())
    }

    return
}

// Blocks until essential login packets have been transmitted.
func (player *Player) postLogin() {
    nearbySent := func() {
        player.lock.Lock()
        defer player.lock.Unlock()

        // Send player start position etc.
        buf := &bytes.Buffer{}
        proto.ServerWritePlayerPositionLook(
            buf, &player.position, &player.look,
            player.position.Y+StanceNormal, false)

        player.inventory.SendUpdate(buf, WindowIdInventory)

        // FIXME: This could potentially deadlock with player.lock being held
        // if the txQueue is full.
        player.TransmitPacket(buf.Bytes())
    }

    player.chunkSubs.Move(&player.position, nearbySent)
}

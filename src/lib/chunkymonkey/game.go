package chunkymonkey

import (
    "bytes"
    "fmt"
    "log"
    "net"
    "os"
    "path"
    "rand"
    "time"

    "nbt/nbt"
    "chunkymonkey/proto"
    "chunkymonkey/record"
    .   "chunkymonkey/types"
)

// The player's starting position is loaded from level.dat for now
var StartPosition AbsXYZ

func loadStartPosition(worldPath string) {
    file, err := os.Open(path.Join(worldPath, "level.dat"), os.O_RDONLY, 0)
    if err != nil {
        log.Exit("loadStartPosition: ", err.String())
    }

    level, err := nbt.Read(file)
    file.Close()
    if err != nil {
        log.Exit("loadStartPosition: ", err.String())
    }

    pos := level.Lookup("/Data/Player/Pos")
    StartPosition = AbsXYZ{
        AbsCoord(pos.(*nbt.List).Value[0].(*nbt.Double).Value),
        AbsCoord(pos.(*nbt.List).Value[1].(*nbt.Double).Value),
        AbsCoord(pos.(*nbt.List).Value[2].(*nbt.Double).Value),
    }
}

type Game struct {
    chunkManager  *ChunkManager
    mainQueue     chan func(*Game)
    entityManager EntityManager
    players       map[EntityID]*Player
    items         map[EntityID]*Item
    time          TimeOfDay
    blockTypes    map[BlockID]*BlockType
    rand          *rand.Rand
}

func (game *Game) Login(conn net.Conn) {
    username, err := proto.ServerReadHandshake(conn)
    if err != nil {
        log.Print("serverReadHandshake: ", err.String())
        proto.WriteDisconnect(conn, err.String())
        conn.Close()
        return
    }
    log.Print("Client ", conn.RemoteAddr(), " connected as ", username)
    proto.ServerWriteHandshake(conn, "-")

    _, _, err = proto.ServerReadLogin(conn)
    if err != nil {
        log.Print("serverReadLogin: ", err.String())
        proto.WriteDisconnect(conn, err.String())
        conn.Close()
        return
    }

    StartPlayer(game, conn, username)
}

func (game *Game) Serve(addr string) {
    listener, e := net.Listen("tcp", addr)
    if e != nil {
        log.Exit("Listen: ", e.String())
    }
    log.Print("Listening on ", addr)

    for {
        conn, e2 := listener.Accept()
        if e2 != nil {
            log.Print("Accept: ", e2.String())
            continue
        }

        go game.Login(record.WrapConn(conn))
    }
}

// Add a player to the game
// This function sends spawn messages to all players in range.  It also spawns
// all existing players so the new player can see them.
func (game *Game) AddPlayer(player *Player) {
    game.entityManager.AddEntity(&player.Entity)
    game.players[player.EntityID] = player
    game.SendChatMessage(fmt.Sprintf("%s has joined", player.name))

    // Spawn new player for existing players
    buf := &bytes.Buffer{}
    proto.WriteNamedEntitySpawn(
        buf,
        player.EntityID, player.name,
        player.position.ToAbsIntXYZ(),
        player.look.ToLookBytes(),
        player.currentItem)

    game.MulticastRadiusPacket(buf.Bytes(), player)

    // Spawn existing players for new player
    buf = &bytes.Buffer{}
    for existing := range game.PlayersInPlayerRadius(player) {
        if existing == player {
            continue
        }

        proto.WriteNamedEntitySpawn(
            buf,
            existing.EntityID, existing.name,
            existing.position.ToAbsIntXYZ(),
            existing.look.ToLookBytes(),
            existing.currentItem)
    }
    player.TransmitPacket(buf.Bytes())
}

// Remove a player from the game
// This function sends destroy messages so the other players see the player
// disappear.
func (game *Game) RemovePlayer(player *Player) {
    // Destroy player for other players
    buf := &bytes.Buffer{}
    proto.WriteEntityDestroy(buf, player.EntityID)
    game.MulticastRadiusPacket(buf.Bytes(), player)

    game.players[player.EntityID] = nil, false
    game.entityManager.RemoveEntity(&player.Entity)
    game.SendChatMessage(fmt.Sprintf("%s has left", player.name))
}

func (game *Game) AddItem(item *Item) {
    game.entityManager.AddEntity(&item.Entity)
    game.items[item.Entity.EntityID] = item

    // Spawn new item for players
    buf := &bytes.Buffer{}
    err := item.SendSpawn(buf)
    if err != nil {
        log.Print("AddItem", err.String())
        return
    }
    game.MulticastChunkPacket(buf.Bytes(), item.position.ToChunkXZ())
}

func (game *Game) MulticastPacket(packet []byte, except *Player) {
    for _, player := range game.players {
        if player == except {
            continue
        }

        player.TransmitPacket(packet)
    }
}

func (game *Game) SendChatMessage(message string) {
    buf := &bytes.Buffer{}
    proto.WriteChatMessage(buf, message)
    game.MulticastPacket(buf.Bytes(), nil)
}

func (game *Game) Enqueue(f func(*Game)) {
    game.mainQueue <- f
}

func (game *Game) mainLoop() {
    for {
        f := <-game.mainQueue
        f(game)
    }
}

func (game *Game) timer() {
    ticker := time.NewTicker(1000000000) // 1 sec
    for {
        <-ticker.C
        game.Enqueue(func(game *Game) { game.tick() })
    }
}

func (game *Game) sendTimeUpdate() {
    buf := &bytes.Buffer{}
    proto.ServerWriteTimeUpdate(buf, game.time)

    // The "keep-alive" packet to client sent here as well, as there seems no
    // particular reason to send time and keep-alive separately for now.
    proto.WriteKeepAlive(buf)

    game.MulticastPacket(buf.Bytes(), nil)
}

func (game *Game) tick() {
    game.time += 20
    game.sendTimeUpdate()
}

func NewGame(worldPath string) (game *Game) {
    chunkManager := NewChunkManager(worldPath)
    loadStartPosition(worldPath)

    game = &Game{
        chunkManager: chunkManager,
        mainQueue:    make(chan func(*Game), 256),
        players:      make(map[EntityID]*Player),
        items:        make(map[EntityID]*Item),
        blockTypes:   LoadStandardBlockTypes(),
        rand:         rand.New(rand.NewSource(1)), // TODO use a better random seed
    }
    chunkManager.game = game

    go game.mainLoop()
    go game.timer()
    return
}

// Return a channel to iterate over all players within a chunk's radius
func (game *Game) PlayersInRadius(loc *ChunkXZ) (c chan *Player) {
    // We return any player whose chunk position is within these bounds:
    minX := loc.X - ChunkRadius
    minZ := loc.Z - ChunkRadius
    maxX := loc.X + ChunkRadius + 1
    maxZ := loc.X + ChunkRadius + 1

    c = make(chan *Player)
    go func() {
        for _, player := range game.players {
            p := player.position.ToChunkXZ()
            if p.X >= minX && p.X <= maxX && p.Z >= minZ && p.Z <= maxZ {
                c <- player
            }
        }
        close(c)
    }()
    return
}

// Return a channel to iterate over all players within a chunk's radius
func (game *Game) PlayersInPlayerRadius(player *Player) chan *Player {
    pos := player.position.ToChunkXZ()
    return game.PlayersInRadius(pos)
}

// Transmit a packet to all players in chunk radius
func (game *Game) MulticastChunkPacket(packet []byte, loc *ChunkXZ) {
    for receiver := range game.PlayersInRadius(loc) {
        receiver.TransmitPacket(packet)
    }
}

// Transmit a packet to all players in radius (except the player itself)
func (game *Game) MulticastRadiusPacket(packet []byte, sender *Player) {
    for receiver := range game.PlayersInPlayerRadius(sender) {
        if receiver == sender {
            continue
        }

        receiver.TransmitPacket(packet)
    }
}

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
    .   "chunkymonkey/entity"
    .   "chunkymonkey/interfaces"
    "chunkymonkey/player"
    "chunkymonkey/proto"
    "chunkymonkey/record"
    "chunkymonkey/serverAuth"
    .   "chunkymonkey/types"
)

type Game struct {
    chunkManager  *ChunkManager
    mainQueue     chan func(IGame)
    entityManager EntityManager
    players       map[EntityID]IPlayer
    time          TimeOfDay
    blockTypes    map[BlockID]*BlockType
    rand          *rand.Rand
    serverId      string

    // The player's starting position is loaded from level.dat for now
    startPosition AbsXYZ
}

func NewGame(worldPath string) (game *Game) {
    chunkManager := NewChunkManager(worldPath)

    game = &Game{
        chunkManager: chunkManager,
        mainQueue:    make(chan func(IGame), 256),
        players:      make(map[EntityID]IPlayer),
        blockTypes:   LoadStandardBlockTypes(),
        rand:         rand.New(rand.NewSource(time.UTC().Seconds())),
    }
    game.loadStartPosition(worldPath)
    game.serverId = fmt.Sprintf("%x", game.rand.Int63())
    //game.serverId = "-"
    chunkManager.game = game
    chunkManager.blockTypes = game.blockTypes

    go game.mainLoop()
    go game.timer()
    return
}

func (game *Game) loadStartPosition(worldPath string) {
    file, err := os.Open(path.Join(worldPath, "level.dat"), os.O_RDONLY, 0)
    if err != nil {
        log.Fatalf("loadStartPosition: %s", err.String())
    }

    level, err := nbt.Read(file)
    file.Close()
    if err != nil {
        log.Fatalf("loadStartPosition: %s", err.String())
    }

    pos := level.Lookup("/Data/Player/Pos")
    game.startPosition = AbsXYZ{
        AbsCoord(pos.(*nbt.List).Value[0].(*nbt.Double).Value),
        AbsCoord(pos.(*nbt.List).Value[1].(*nbt.Double).Value),
        AbsCoord(pos.(*nbt.List).Value[2].(*nbt.Double).Value),
    }
}

func (game *Game) Login(conn net.Conn) {
    username, err := proto.ServerReadHandshake(conn)
    if err != nil {
        log.Print("ServerReadHandshake: ", err.String())
        proto.WriteDisconnect(conn, err.String())
        conn.Close()
        return
    }
    log.Print("Client ", conn.RemoteAddr(), " connected as ", username)

    err = proto.ServerWriteHandshake(conn, game.serverId)
    if err != nil {
        log.Print("ServerWriteHandshake: ", err.String())
        proto.WriteDisconnect(conn, err.String())
        conn.Close()
        return
    }

    // TODO put authentication into a seperate module behind an interface so
    // that authentication is pluggable
    if game.serverId != "-" {
        var authenticated bool
        authenticated, err = serverAuth.CheckUserAuth(game.serverId, username)
        if !authenticated || err != nil {
            var reason string
            if err != nil {
                reason = "Authentication check failed: " + err.String()
            } else {
                reason = "Failed authentication"
            }
            log.Print("Client ", conn.RemoteAddr(), " ", reason)
            proto.WriteDisconnect(conn, reason)
            conn.Close()
            return
        }
        log.Print("Client ", conn.RemoteAddr(), " passed minecraft.net authentication")
    }

    _, _, err = proto.ServerReadLogin(conn)
    if err != nil {
        log.Print("ServerReadLogin: ", err.String())
        proto.WriteDisconnect(conn, err.String())
        conn.Close()
        return
    }

    player.StartPlayer(game, conn, username)
}

func (game *Game) Serve(addr string) {
    listener, e := net.Listen("tcp", addr)
    if e != nil {
        log.Fatalf("Listen: %s", e.String())
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

func (game *Game) GetStartPosition() *AbsXYZ {
    return &game.startPosition
}

func (game *Game) GetChunkManager() IChunkManager {
    return game.chunkManager
}

func (game *Game) AddEntity(entity *Entity) {
    game.entityManager.AddEntity(entity)
}

// Add a player to the game
// This function sends spawn messages to all players in range.  It also spawns
// all existing players so the new player can see them.
func (game *Game) AddPlayer(player IPlayer) {
    entity := player.GetEntity()
    game.AddEntity(entity)
    game.players[entity.EntityID] = player
    name := player.GetName()
    game.SendChatMessage(fmt.Sprintf("%s has joined", name))

    // Spawn new player for existing players
    player.Enqueue(func(player IPlayer) {
        buf := &bytes.Buffer{}
        if err := player.SendSpawn(buf); err != nil {
            return
        }
        game.Enqueue(func(game IGame) {
            game.MulticastRadiusPacket(buf.Bytes(), player)
        })
    })

    // Spawn existing players for new player
    for existing := range game.PlayersInPlayerRadius(player) {
        if existing == player {
            continue
        }

        existing.Enqueue(func(player IPlayer) {
            buf := &bytes.Buffer{}
            if err := player.SendSpawn(buf); err != nil {
                return
            }
            player.TransmitPacket(buf.Bytes())
        })
    }
}

// Remove a player from the game
// This function sends destroy messages so the other players see the player
// disappear.
func (game *Game) RemovePlayer(player IPlayer) {
    // Destroy player for other players
    buf := &bytes.Buffer{}
    entity := player.GetEntity()
    proto.WriteEntityDestroy(buf, entity.EntityID)
    game.MulticastRadiusPacket(buf.Bytes(), player)

    game.players[entity.EntityID] = nil, false
    game.entityManager.RemoveEntity(entity)
    game.SendChatMessage(fmt.Sprintf("%s has left", player.GetName()))
}

func (game *Game) MulticastPacket(packet []byte, except interface{}) {
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

func (game *Game) Enqueue(f func(IGame)) {
    game.mainQueue <- f
}

func (game *Game) mainLoop() {
    for {
        f := <-game.mainQueue
        f(game)
    }
}

func (game *Game) timer() {
    ticker := time.NewTicker(NanosecondsInSecond / TicksPerSecond)
    for {
        <-ticker.C
        game.Enqueue(func(igame IGame) { game.tick() })
    }
}

func (game *Game) sendTimeUpdate() {
    buf := &bytes.Buffer{}
    proto.ServerWriteTimeUpdate(buf, game.time)

    // The "keep-alive" packet to client(s) sent here as well, as there seems
    // no particular reason to send time and keep-alive separately for now.
    proto.WriteKeepAlive(buf)

    game.MulticastPacket(buf.Bytes(), nil)

    // Get chunks to send various updates
    for _, chunk := range game.chunkManager.chunks {
        chunk.Enqueue(func(chunk IChunk) {
            chunk.SendUpdate()
        })
    }
}

func (game *Game) physicsTick() {
    for _, chunk := range game.chunkManager.chunks {
        chunk.Enqueue(func(chunk IChunk) {
            chunk.PhysicsTick()
        })
    }
}

func (game *Game) tick() {
    game.time = (game.time + DayTicksPerTick) % DayTicksPerDay
    if game.time%DayTicksPerSecond == 0 {
        game.sendTimeUpdate()
    }
    game.physicsTick()
}

// Return a channel to iterate over all players within a chunk's radius
func (game *Game) PlayersInRadius(loc *ChunkXZ) (c chan IPlayer) {
    // We return any player whose chunk position is within these bounds:
    minX := loc.X - ChunkRadius
    minZ := loc.Z - ChunkRadius
    maxX := loc.X + ChunkRadius
    maxZ := loc.X + ChunkRadius

    c = make(chan IPlayer)
    go func() {
        for _, player := range game.players {
            // FIXME this reads player position from the wrong goroutine
            p := player.GetChunkPosition()
            if p.X >= minX && p.X <= maxX && p.Z >= minZ && p.Z <= maxZ {
                c <- player
            }
        }
        close(c)
    }()
    return
}

// Return a channel to iterate over all players within a chunk's radius
func (game *Game) PlayersInPlayerRadius(player IPlayer) chan IPlayer {
    pos := player.GetChunkPosition()
    return game.PlayersInRadius(pos)
}

// Transmit a packet to all players in chunk radius
func (game *Game) MulticastChunkPacket(packet []byte, loc *ChunkXZ) {
    for receiver := range game.PlayersInRadius(loc) {
        receiver.TransmitPacket(packet)
    }
}

// Transmit a packet to all players in radius (except the player itself)
func (game *Game) MulticastRadiusPacket(packet []byte, sender IPlayer) {
    for receiver := range game.PlayersInPlayerRadius(sender) {
        if receiver == sender {
            continue
        }

        receiver.TransmitPacket(packet)
    }
}

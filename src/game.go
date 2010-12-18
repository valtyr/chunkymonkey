package main

import (
    "bytes"
    "log"
    "net"
    "time"
    "fmt"
)

type Game struct {
    chunkManager  *ChunkManager
    mainQueue     chan func(*Game)
    entityManager EntityManager
    players       map[EntityID]*Player
    time          int64
}

func (game *Game) Login(conn net.Conn) {
    username, err := ReadHandshake(conn)
    if err != nil {
        log.Print("ReadHandshake: ", err.String())
        return
    }
    log.Print("Client ", conn.RemoteAddr(), " connected as ", username)
    WriteHandshake(conn, "-")

    _, _, err = ReadLogin(conn)
    if err != nil {
        log.Print("ReadLogin: ", err.String())
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

        go game.Login(WrapConn(conn))
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
    WriteNamedEntitySpawn(buf, player.EntityID, player.name, &player.position, &player.orientation, player.currentItem)
    game.MulticastRadiusPacket(buf.Bytes(), player)

    // Spawn existing players for new player
    buf = &bytes.Buffer{}
    for existing := range game.PlayersInPlayerRadius(player) {
        if existing == player {
            continue
        }

        WriteNamedEntitySpawn(buf, existing.EntityID, existing.name, &existing.position, &existing.orientation, existing.currentItem)
    }
    player.TransmitPacket(buf.Bytes())
}

// Remove a player from the game
// This function sends destroy messages so the other players see the player
// disappear.
func (game *Game) RemovePlayer(player *Player) {
    // Destroy player for other players
    buf := &bytes.Buffer{}
    WriteDestroyEntity(buf, player.EntityID)
    game.MulticastRadiusPacket(buf.Bytes(), player)

    game.players[player.EntityID] = nil, false
    game.entityManager.RemoveEntity(&player.Entity)
    game.SendChatMessage(fmt.Sprintf("%s has left", player.name))
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
    WriteChatMessage(buf, message)
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
    WriteTimeUpdate(buf, game.time)
    game.MulticastPacket(buf.Bytes(), nil)
}

func (game *Game) tick() {
    game.time += 20
    game.sendTimeUpdate()
}

func NewGame(chunkManager *ChunkManager) (game *Game) {
    game = &Game{
        chunkManager: chunkManager,
        mainQueue:    make(chan func(*Game), 256),
        players:      make(map[EntityID]*Player),
    }

    go game.mainLoop()
    go game.timer()
    return
}

// Return a channel to iterate over all players within a chunk's radius
func (game *Game) PlayersInRadius(loc ChunkXZ) (c chan *Player) {
    // We return any player whose chunk position is within these bounds:
    minX := loc.x - ChunkRadius
    minZ := loc.z - ChunkRadius
    maxX := loc.x + ChunkRadius + 1
    maxZ := loc.x + ChunkRadius + 1

    c = make(chan *Player)
    go func() {
        for _, player := range game.players {
            // FIXME something else might concurrently write to player position
            // here, so there's potential for failure.
            p := player.position.ToChunkXZ()
            if p.x >= minX && p.x <= maxX && p.z >= minZ && p.z <= maxZ {
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
func (game *Game) MulticastChunkPacket(packet []byte, loc ChunkXZ) {
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

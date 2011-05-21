package chunkymonkey

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"rand"
	"time"

	"chunkymonkey/shardserver"
	. "chunkymonkey/entity"
	"chunkymonkey/gamerules"
	. "chunkymonkey/interfaces"
	"chunkymonkey/itemtype"
	"chunkymonkey/player"
	"chunkymonkey/proto"
	"chunkymonkey/record"
	"chunkymonkey/server_auth"
	. "chunkymonkey/types"
	"chunkymonkey/worldstore"
)

type Game struct {
	chunkManager  *shardserver.LocalShardManager
	mainQueue     chan func(IGame)
	entityManager EntityManager
	players       map[EntityId]IPlayer
	time          TimeOfDay
	gameRules     gamerules.GameRules
	itemTypes     itemtype.ItemTypeMap
	rand          *rand.Rand
	serverId      string
	worldStore    *worldstore.WorldStore
	// If set, logins are not allowed.
	UnderMaintenanceMsg string
}

func NewGame(worldPath string, gameRules *gamerules.GameRules) (game *Game, err os.Error) {
	worldStore, err := worldstore.LoadWorldStore(worldPath)

	game = &Game{
		mainQueue:  make(chan func(IGame), 256),
		players:    make(map[EntityId]IPlayer),
		gameRules:  *gameRules,
		rand:       rand.New(rand.NewSource(time.UTC().Seconds())),
		worldStore: worldStore,
	}

	game.entityManager.Init()

	game.serverId = fmt.Sprintf("%x", game.rand.Int63())
	//game.serverId = "-"

	game.chunkManager = shardserver.NewLocalShardManager(worldStore.ChunkStore, game)

	go game.mainLoop()
	go game.timer()
	return
}

// login negotiates a player client login, and adds a new player if successful.
// Note that it does not run in the game's goroutine.
func (game *Game) login(conn net.Conn) {
	username, err := proto.ServerReadHandshake(conn)
	if err != nil {
		log.Print("ServerReadHandshake: ", err.String())
		proto.WriteDisconnect(conn, err.String())
		conn.Close()
		return
	}
	log.Print("Client ", conn.RemoteAddr(), " connected as ", username)
	if game.UnderMaintenanceMsg != "" {
		log.Println("Server under maintenance, kicking player:", username)
		proto.WriteDisconnect(conn, game.UnderMaintenanceMsg)
		return
	}

	err = proto.ServerWriteHandshake(conn, game.serverId)
	if err != nil {
		log.Print("ServerWriteHandshake: ", err.String())
		proto.WriteDisconnect(conn, err.String())
		conn.Close()
		return
	}

	// TODO put authentication into a seperate module behind an interface so
	// that authentication is pluggable.
	if game.serverId != "-" {
		var authenticated bool
		authenticated, err = server_auth.CheckUserAuth(game.serverId, username)
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

	_, err = proto.ServerReadLogin(conn)
	if err != nil {
		log.Print("ServerReadLogin: ", err.String())
		proto.WriteDisconnect(conn, err.String())
		conn.Close()
		return
	}

	startPosition := game.GetStartPosition()

	// TODO pass player's last position in the world, not necessarily the spawn
	// position.
	player := player.NewPlayer(game, conn, username, startPosition)

	addedChan := make(chan struct{})
	game.Enqueue(func(game IGame) {
		game.AddPlayer(player)
		addedChan<- struct{}{}
	})
	_ = <-addedChan

	player.Start()

	buf := &bytes.Buffer{}
	// TODO pass proper dimension. This is low priority, because there is
	// currently no way to update the client's dimension after login.
	proto.ServerWriteLogin(buf, player.EntityId, 0, DimensionNormal)
	proto.WriteSpawnPosition(buf, startPosition.ToBlockXyz())
	player.TransmitPacket(buf.Bytes())

	game.Enqueue(func(_ IGame) {
		game.spawnPlayer(player)
	})
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

		go game.login(record.WrapConn(conn))
	}
}

func (game *Game) GetStartPosition() AbsXyz {
	return game.worldStore.StartPosition
}

func (game *Game) GetGameRules() *gamerules.GameRules {
	return &game.gameRules
}

func (game *Game) GetChunkManager() IShardConnecter {
	return game.chunkManager
}

func (game *Game) AddEntity(entity *Entity) {
	game.entityManager.AddEntity(entity)
}

func (game *Game) RemoveEntity(entity *Entity) {
	game.entityManager.RemoveEntity(entity)
}

// AddPlayer sends spawn messages to all players in range. It also spawns all
// existing players so the new player can see them.
func (game *Game) AddPlayer(newPlayer IPlayer) {
	entity := newPlayer.GetEntity()
	game.AddEntity(entity)
	game.players[entity.EntityId] = newPlayer
}

// RemovePlayer sends destroy messages so the other players see the player
// disappear.
func (game *Game) RemovePlayer(player IPlayer) {
	// Destroy player for other players
	buf := &bytes.Buffer{}
	entity := player.GetEntity()
	proto.WriteEntityDestroy(buf, entity.EntityId)
	game.multicastRadiusPacket(buf.Bytes(), player)

	game.players[entity.EntityId] = nil, false
	game.entityManager.RemoveEntity(entity)
	game.SendChatMessage(fmt.Sprintf("%s has left", player.GetName()))
}

func (game *Game) spawnPlayer(newPlayer IPlayer) {
	name := newPlayer.GetName()
	game.SendChatMessage(fmt.Sprintf("%s has joined", name))

	// Spawn new player for existing players.
	newPlayer.Enqueue(func(newPlayer IPlayer) {
		buf := &bytes.Buffer{}
		if err := newPlayer.SendSpawn(buf); err != nil {
			return
		}
		game.Enqueue(func(_ IGame) {
			game.multicastRadiusPacket(buf.Bytes(), newPlayer)
		})
	})

	// Spawn existing players for new player.
	p1, p2 := getChunkRadius(newPlayer.LockedGetChunkPosition())
	for _, existing := range game.players {
		if existing != newPlayer {
			existing.Enqueue(func(existing IPlayer) {
				if existing.IsWithin(p1, p2) {
					buf := &bytes.Buffer{}
					if err := existing.SendSpawn(buf); err != nil {
						return
					}
					newPlayer.TransmitPacket(buf.Bytes())
				}
			})
		}
	}
}

// TODO use of MulticastPacket should be removed in favour of chunk multicast
// packets. (apply this to chat as well, or use some other solution?)
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

	// The "keep-alive" packet to client(s) sent here as well, as there
	// seems no particular reason to send time and keep-alive separately
	// for now.
	proto.WriteKeepAlive(buf)

	game.MulticastPacket(buf.Bytes(), nil)

	// TODO: Make chunk shards responsible for sending updates.
	game.chunkManager.EnqueueAllChunks(func(chunk IChunk) {
		chunk.SendUpdate()
	})
}

func (game *Game) physicsTick() {
	// TODO: Make chunk shards responsible for ticks.
	game.chunkManager.EnqueueAllChunks(func(chunk IChunk) {
		chunk.Tick()
	})
}

func (game *Game) tick() {
	game.time = (game.time + DayTicksPerTick) % DayTicksPerDay
	if game.time%DayTicksPerSecond == 0 {
		game.sendTimeUpdate()
	}
	game.physicsTick()
}

// Transmit a packet to all players near the sender (except the sender itself).
func (game *Game) multicastRadiusPacket(packet []byte, sender IPlayer) {
	game.chunkManager.EnqueueOnChunk(
		sender.LockedGetChunkPosition(),
		func(chunk IChunk) {
			chunk.MulticastPlayers(sender.GetEntityId(), packet)
		},
	)
}

func getChunkRadius(loc ChunkXz) (p1, p2 *ChunkXz) {

	p1 = &ChunkXz{
		loc.X - ChunkRadius,
		loc.Z - ChunkRadius,
	}
	p2 = &ChunkXz{
		loc.X + ChunkRadius,
		loc.Z + ChunkRadius,
	}
	return
}

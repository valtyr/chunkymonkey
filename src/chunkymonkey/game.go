package chunkymonkey

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"rand"
	"regexp"
	"time"

	. "chunkymonkey/entity"
	"chunkymonkey/gamerules"
	"chunkymonkey/player"
	"chunkymonkey/proto"
	"chunkymonkey/server_auth"
	"chunkymonkey/shardserver"
	. "chunkymonkey/types"
	"chunkymonkey/worldstore"
	"nbt"
)

// We regard usernames as valid if they don't contain "dangerous" characters.
// That is: characters that might be abused in filename components, etc.
var validPlayerUsername = regexp.MustCompile(`^[\-a-zA-Z0-9_]+$`)


type Game struct {
	chunkManager     *shardserver.LocalShardManager
	mainQueue        chan func(*Game)
	playerDisconnect chan EntityId
	entityManager    EntityManager
	players          map[EntityId]*player.Player
	time             Ticks
	serverId         string
	worldStore       *worldstore.WorldStore
	// If set, logins are not allowed.
	UnderMaintenanceMsg string
}

func NewGame(worldPath string) (game *Game, err os.Error) {
	worldStore, err := worldstore.LoadWorldStore(worldPath)
	if err != nil {
		return nil, err
	}

	game = &Game{
		mainQueue:        make(chan func(*Game), 256),
		playerDisconnect: make(chan EntityId),
		players:          make(map[EntityId]*player.Player),
		time:             worldStore.Time,
		worldStore:       worldStore,
	}

	game.entityManager.Init()

	game.serverId = fmt.Sprintf("%016x", rand.NewSource(worldStore.Seed).Int63())
	//game.serverId = "-"

	game.chunkManager = shardserver.NewLocalShardManager(worldStore.ChunkStore, &game.entityManager)

	go game.mainLoop()
	return
}

// login negotiates a player client login, and adds a new player if successful.
// Note that it does not run in the game's goroutine.
func (game *Game) login(conn net.Conn) {
	var err, clientErr os.Error

	defer func() {
		if err != nil {
			log.Print(err.String())
			if clientErr == nil {
				clientErr = os.NewError("Server error.")
			}
			proto.WriteDisconnect(conn, clientErr.String())
			conn.Close()
		}
	}()

	var username string
	if username, err = proto.ServerReadHandshake(conn); err != nil {
		clientErr = os.NewError("Handshake error.")
		return
	}

	if !validPlayerUsername.MatchString(username) {
		err = os.NewError("Bad username")
		clientErr = err
		return
	}

	log.Print("Client ", conn.RemoteAddr(), " connected as ", username)

	if game.UnderMaintenanceMsg != "" {
		err = fmt.Errorf("Server under maintenance, kicking player: %q", username)
		clientErr = os.NewError(game.UnderMaintenanceMsg)
		return
	}

	// Load player permissions.
	permissions := gamerules.Permissions.UserPermissions(username)
	if !permissions.Has("login") {
		err = fmt.Errorf("Player %q does not have login permission", username)
		clientErr = os.NewError("You do not have access to this server.")
		return
	}

	if err = proto.ServerWriteHandshake(conn, game.serverId); err != nil {
		clientErr = os.NewError("Handshake error.")
		return
	}

	if game.serverId != "-" {
		var authenticated bool
		authserver := &server_auth.ServerAuth{"http://www.minecraft.net/game/checkserver.jsp"}
		authenticated, err = authserver.Authenticate(game.serverId, username)
		if !authenticated || err != nil {
			var reason string
			if err != nil {
				reason = "Authentication check failed: " + err.String()
			} else {
				reason = "Failed authentication"
			}
			err = fmt.Errorf("Client %v: %s", conn.RemoteAddr(), reason)
			clientErr = os.NewError(reason)
			return
		}
		log.Print("Client ", conn.RemoteAddr(), " passed minecraft.net authentication")
	}

	if _, err = proto.ServerReadLogin(conn); err != nil {
		clientErr = os.NewError("Login error.")
		return
	}

	entityId := game.entityManager.NewEntity()

	var playerData nbt.ITag
	if playerData, err = game.worldStore.PlayerData(username); err != nil {
		clientErr = os.NewError("Error reading user data. Please contact the server administrator.")
		return
	}

	player := player.NewPlayer(entityId, game.chunkManager, conn, username, game.worldStore.SpawnPosition, game.playerDisconnect, game)
	if playerData != nil {
		if err = player.ReadNbt(playerData); err != nil {
			// Don't let the player log in, as they will only have default inventory
			// etc., which could lose items from them. Better for an administrator to
			// sort this out.
			err = fmt.Errorf("Error parsing player data for %q: %v", username, err)
			clientErr = os.NewError("Error reading user data. Please contact the server administrator.")
			return
		}
	}

	addedChan := make(chan struct{})
	game.enqueue(func(_ *Game) {
		game.addPlayer(player)
		addedChan <- struct{}{}
	})
	_ = <-addedChan

	player.Start()
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

		go game.login(conn)
	}
}

// addPlayer adds the player to the set of connected players.
func (game *Game) addPlayer(newPlayer *player.Player) {
	game.players[newPlayer.GetEntityId()] = newPlayer
}

func (game *Game) removePlayer(entityId EntityId) {
	game.players[entityId] = nil, false
	game.entityManager.RemoveEntityById(entityId)
}

func (game *Game) multicastPacket(packet []byte, except interface{}) {
	for _, player := range game.players {
		if player == except {
			continue
		}

		player.TransmitPacket(packet)
	}
}

func (game *Game) enqueue(f func(*Game)) {
	game.mainQueue <- f
}

func (game *Game) mainLoop() {
	ticker := time.NewTicker(NanosecondsInSecond / TicksPerSecond)

	for {
		select {
		case f := <-game.mainQueue:
			f(game)
		case <-ticker.C:
			game.tick()
		case entityId := <-game.playerDisconnect:
			game.removePlayer(entityId)
		}
	}
}

func (game *Game) sendTimeUpdate() {
	buf := new(bytes.Buffer)
	proto.ServerWriteTimeUpdate(buf, game.time)

	// The "keep-alive" packet to client(s) sent here as well, as there
	// seems no particular reason to send time and keep-alive separately
	// for now.
	proto.WriteKeepAlive(buf)

	game.multicastPacket(buf.Bytes(), nil)
}

func (game *Game) tick() {
	game.time++
	if game.time%TicksPerSecond == 0 {
		game.sendTimeUpdate()
	}
}

func (game *Game) getPlayerFromName(name string) *player.Player {
	// TODO: This should be made more efficient through a lookup, etc.
	result := make(chan *player.Player)
	game.enqueue(func(_ *Game) {
		for _, player := range game.players {
			if player.Name() == name {
				result <- player
			}
		}
		close(result)
	})
	return <-result
}

// GiveItem implements ICommandHandler.GiveItem
func (game *Game) GiveItem(name string, id, quantity, data int) {
	//	player := game.getPlayerFromName(name)
	//	item := gamerules.Slot{
	//		ItemTypeId: ItemTypeId(id),
	//		Count:      ItemCount(quantity),
	//		Data:       ItemData(data),
	//	}

	// TODO: Spawn the item created at the player's block location
}

// SendMessageToPlayer implements ICommandHandler.SendMessageToPlayer
func (game *Game) SendMessageToPlayer(name, msg string) {
	player := game.getPlayerFromName(name)

	buf := new(bytes.Buffer)
	proto.WriteChatMessage(buf, msg)
	packet := buf.Bytes()
	player.TransmitPacket(packet)
}

// BroadcastMessage implements ICommandHandler.BroadcastMessage
func (game *Game) BroadcastMessage(msg string) {
	buf := new(bytes.Buffer)
	proto.WriteChatMessage(buf, msg)

	game.enqueue(func(_ *Game) {
		game.multicastPacket(buf.Bytes(), nil)
	})
}

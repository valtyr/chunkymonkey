package chunkymonkey

import (
	"fmt"
	"log"
	"os"
	"net"

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

const (
	connTypeUnknown = iota
	connTypeLogin
	connTypeServerQuery
)

var (
	clientErrGeneral = os.NewError("server error")
	clientErrUsername = os.NewError("bad username")
	clientErrLoginDenied = os.NewError("you do not have access to this server")
	clientErrAuthFailed = os.NewError("Minecraft authentication failed")

	loginErrorConnType = os.NewError("unknown/bad connection type")
	loginErrorMaintenance = os.NewError("server under maintenance")
)

type GameInfo struct {
	game           *Game
	maintenanceMsg string
	serverId       string
	shardManager   *shardserver.LocalShardManager
	entityManager  *EntityManager
	worldStore     *worldstore.WorldStore
	authserver     server_auth.IAuthenticator
}

// Handles connections for a game on the given socket.
type ConnManager struct {
	listener net.Listener
	gameInfo *GameInfo
}

// NewConnManager creates and starts a ConnManager.
func NewConnManager(listener net.Listener, gameInfo *GameInfo) *ConnManager {
	mgr := &ConnManager{
		listener: listener,
		gameInfo: gameInfo,
	}

	go mgr.run()

	return mgr
}

func (mgr *ConnManager) run() {
	defer mgr.listener.Close()

	for {
		conn, err := mgr.listener.Accept()
		if err != nil {
			log.Print("Accept: ", err)
			return
		}

		go handleConnection(mgr.gameInfo, conn)
	}
}

type login struct {
	gameInfo *GameInfo
	conn     net.Conn

	connType  int
	username  string
}

func handleConnection(gameInfo *GameInfo, conn net.Conn) {
	var err, clientErr os.Error

	defer func() {
		if err != nil {
			log.Print("Connection denied", err.String())
			if clientErr == nil {
				clientErr = clientErrGeneral
			}
			proto.WriteDisconnect(conn, clientErr.String())
			conn.Close()
		}
	}()

	l := &login{
		gameInfo: gameInfo,
		conn:     conn,
	}

	err = proto.ServerReadPacketExpect(conn, l, []byte{
		proto.PacketIdHandshake,
		proto.PacketIdServerListPing,
	})
	if err != nil {
		clientErr = os.NewError("Login error.")
		return
	}

	switch l.connType {
	case connTypeLogin:
		err, clientErr = l.handleLogin(conn)
	case connTypeServerQuery:
		err, clientErr = l.handleServerQuery(conn)
	default:
		err = loginErrorConnType
	}
}

func (l *login) handleLogin(conn net.Conn) (err, clientErr os.Error) {
	if !validPlayerUsername.MatchString(l.username) {
		err = clientErrUsername
		clientErr = err
		return
	}

	log.Print("Client ", conn.RemoteAddr(), " connected as ", l.username)

	// TODO Allow admins to connect.
	if l.gameInfo.maintenanceMsg != "" {
		err = loginErrorMaintenance
		clientErr = os.NewError(l.gameInfo.maintenanceMsg)
		return
	}

	// Load player permissions.
	permissions := gamerules.Permissions.UserPermissions(l.username)
	if !permissions.Has("login") {
		err = fmt.Errorf("Player %q does not have login permission", l.username)
		clientErr = clientErrLoginDenied
		return
	}

	if err = proto.ServerWriteHandshake(conn, l.gameInfo.serverId); err != nil {
		clientErr = os.NewError("Handshake error.")
		return
	}

	if l.gameInfo.serverId != "-" {
		var authenticated bool
		authenticated, err = l.gameInfo.authserver.Authenticate(l.gameInfo.serverId, l.username)
		if !authenticated || err != nil {
			var reason string
			if err != nil {
				reason = "Authentication check failed: " + err.String()
			} else {
				reason = "Failed authentication"
			}
			err = fmt.Errorf("Client %v: %s", conn.RemoteAddr(), reason)
			clientErr = clientErrAuthFailed
			return
		}
		log.Print("Client ", conn.RemoteAddr(), " passed minecraft.net authentication")
	}

	err = proto.ServerReadPacketExpect(conn, l, []byte{
		proto.PacketIdLogin,
	})
	if err != nil {
		clientErr = os.NewError("Login error.")
		return
	}

	entityId := l.gameInfo.entityManager.NewEntity()

	var playerData *nbt.Compound
	if playerData, err = l.gameInfo.game.worldStore.PlayerData(l.username); err != nil {
		clientErr = os.NewError("Error reading user data. Please contact the server administrator.")
		return
	}

	player := player.NewPlayer(entityId, l.gameInfo.shardManager, conn, l.username, l.gameInfo.worldStore.SpawnPosition, l.gameInfo.game.playerDisconnect, l.gameInfo.game)
	if playerData != nil {
		if err = player.UnmarshalNbt(playerData); err != nil {
			// Don't let the player log in, as they will only have default inventory
			// etc., which could lose items from them. Better for an administrator to
			// sort this out.
			err = fmt.Errorf("Error parsing player data for %q: %v", l.username, err)
			clientErr = os.NewError("Error reading user data. Please contact the server administrator.")
			return
		}
	}

	l.gameInfo.game.playerConnect <- player
	player.Run()

	return
}

func (l *login) handleServerQuery(conn net.Conn) (err, clientErr os.Error) {
	// TODO

	return
}

func (l *login) PacketServerLogin(username string) {
}

func (l *login) PacketServerHandshake(username string) {
	l.connType = connTypeLogin
	l.username = username
}

func (l *login) PacketServerListPing() {
	l.connType = connTypeServerQuery
}

func (l *login) PacketKeepAlive(id int32) {}

func (l *login) PacketChatMessage(message string) {}

func (l *login) PacketPlayer(onGround bool) {}

func (l *login) PacketHoldingChange(slotId SlotId) {}

func (l *login) PacketEntityAction(entityId EntityId, action EntityAction) {}

func (l *login) PacketUseEntity(user EntityId, target EntityId, leftClick bool) {}

func (l *login) PacketRespawn(dimension DimensionId, unknown int8, gameType GameType, worldHeight int16, mapSeed RandomSeed) {}

func (l *login) PacketPlayerPosition(position *AbsXyz, stance AbsCoord, onGround bool) {}

func (l *login) PacketPlayerLook(look *LookDegrees, onGround bool) {}

func (l *login) PacketPlayerBlockHit(status DigStatus, blockLoc *BlockXyz, face Face) {}

func (l *login) PacketPlayerBlockInteract(itemTypeId ItemTypeId, blockLoc *BlockXyz, face Face, amount ItemCount, data ItemData) {}

func (l *login) PacketEntityAnimation(entityId EntityId, animation EntityAnimation) {}

func (l *login) PacketUnknown0x1b(field1, field2 float32, field3, field4 bool, field5, field6 float32) {}

func (l *login) PacketUnknown0x3d(field1, field2 int32, field3 int8, field4, field5 int32) {}

func (l *login) PacketWindowClose(windowId WindowId) {}

func (l *login) PacketWindowClick(windowId WindowId, slot SlotId, rightClick bool, txId TxId, shiftClick bool, expectedSlot *proto.WindowSlot) {}

func (l *login) PacketWindowTransaction(windowId WindowId, txId TxId, accepted bool) {}

func (l *login) PacketSignUpdate(position *BlockXyz, lines [4]string) {}

func (l *login) PacketDisconnect(reason string) {}

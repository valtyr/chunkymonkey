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

	"chunkymonkey/entity"
	. "chunkymonkey/interfaces"
	"chunkymonkey/inventory"
	"chunkymonkey/itemtype"
	"chunkymonkey/proto"
	"chunkymonkey/recipe"
	"chunkymonkey/slot"
	"chunkymonkey/shardserver_external"
	. "chunkymonkey/types"
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
	entity.Entity
	game           IGame
	shardConnecter shardserver_external.IShardConnecter
	conn           net.Conn
	name           string
	position       AbsXyz
	look           LookDegrees
	chunkSubs      chunkSubscriptions

	cursor       slot.Slot // Item being moved by mouse cursor.
	inventory    inventory.PlayerInventory
	curWindow    inventory.IWindow
	nextWindowId WindowId

	mainQueue chan func(IPlayer)
	txQueue   chan []byte
	lock      sync.Mutex
}

func NewPlayer(game IGame, shardConnecter shardserver_external.IShardConnecter, recipes *recipe.RecipeSet, conn net.Conn, name string, position AbsXyz) *Player {
	player := &Player{
		game:           game,
		shardConnecter: shardConnecter,
		conn:           conn,
		name:           name,
		position:       position,
		look:           LookDegrees{0, 0},

		curWindow:    nil,
		nextWindowId: WindowIdFreeMin,

		mainQueue: make(chan func(IPlayer), 128),
		txQueue:   make(chan []byte, 128),
	}

	player.cursor.Init()
	player.inventory.Init(player.EntityId, player, recipes)

	return player
}

func (player *Player) GetEntityId() EntityId {
	return player.EntityId
}

func (player *Player) GetEntity() *entity.Entity {
	return &player.Entity
}

func (player *Player) LockedGetChunkPosition() ChunkXz {
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

func (player *Player) SendSpawn(writer io.Writer) (err os.Error) {
	heldSlot, _ := player.inventory.HeldItem()
	heldItemId := heldSlot.GetItemTypeId()
	if heldItemId < 0 {
		heldItemId = 0
	}

	err = proto.WriteNamedEntitySpawn(
		writer,
		player.EntityId, player.name,
		player.position.ToAbsIntXyz(),
		player.look.ToLookBytes(),
		heldItemId,
	)
	if err != nil {
		return
	}
	return player.inventory.SendFullEquipmentUpdate(writer)
}

func (player *Player) GetHeldItemType() *itemtype.ItemType {
	slot, _ := player.inventory.HeldItem()
	return slot.ItemType
}

func (player *Player) TakeOneHeldItem(into *slot.Slot) {
	player.inventory.TakeOneHeldItem(into)
}

func (player *Player) Start() {
	go player.receiveLoop()
	go player.transmitLoop()
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
	player.chunkSubs.Move(position)

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

func (player *Player) PacketPlayerBlockHit(status DigStatus, blockLoc *BlockXyz, face Face) {
	// TODO validate that the player is actually somewhere near the block

	// TODO validate that the player has dug long enough to stop speed
	// hacking (based on block type and tool used - non-trivial).

	if face != FaceNull {
		chunkLoc, subLoc := blockLoc.ToChunkLocal()

		player.shardConnecter.EnqueueOnChunk(*chunkLoc, func(chunk shardserver_external.IChunk) {
			chunk.PlayerBlockHit(player, subLoc, status)
		})

	} else {
		// TODO player dropped item
	}
}

func (player *Player) PacketPlayerBlockInteract(itemId ItemTypeId, blockLoc *BlockXyz, face Face, amount ItemCount, uses ItemData) {
	if face < FaceMinValid || face > FaceMaxValid {
		// TODO sometimes FaceNull means something. This case should be covered.
		log.Printf("Player/PacketPlayerBlockInteract: invalid face %d", face)
		return
	}

	placeChunkLoc, _ := blockLoc.ToChunkLocal()

	player.shardConnecter.EnqueueOnChunk(*placeChunkLoc, func(chunk shardserver_external.IChunk) {
		chunk.PlayerBlockInteract(player, blockLoc, face)
	})
}

func (player *Player) PacketHoldingChange(slotId SlotId) {
	player.lock.Lock()
	defer player.lock.Unlock()
	player.inventory.SetHolding(slotId)
}

func (player *Player) PacketEntityAnimation(entityId EntityId, animation EntityAnimation) {
}

func (player *Player) PacketUnknown0x1b(field1, field2 float32, field3, field4 bool, field5, field6 float32) {
}

func (player *Player) PacketWindowClose(windowId WindowId) {
	player.lock.Lock()
	defer player.lock.Unlock()

	if player.curWindow != nil && player.curWindow.GetWindowId() == windowId {
		player.curWindow.Finalize(false)
	}
}

func (player *Player) PacketWindowClick(windowId WindowId, slotId SlotId, rightClick bool, txId TxId, shiftClick bool, itemId ItemTypeId, amount ItemCount, uses ItemData) {
	player.lock.Lock()
	defer player.lock.Unlock()

	// Note that the parameters itemId, amount and uses are all currently
	// ignored. The item(s) involved are worked out from the server-side data.

	// Determine which inventory window is involved.
	// TODO support for more windows

	var clickedWindow inventory.IWindow
	if windowId == WindowIdInventory {
		clickedWindow = &player.inventory
	} else if player.curWindow != nil && player.curWindow.GetWindowId() == windowId {
		clickedWindow = player.curWindow
	} else {
		log.Printf(
			"Warning: ignored window click on unknown window ID %d",
			windowId)
	}

	buf := &bytes.Buffer{}
	accepted := false

	if clickedWindow != nil {
		accepted = clickedWindow.Click(slotId, &player.cursor, rightClick, shiftClick)

		// We send slot updates in case we have custom max counts that differ
		// from the client's own model.
		player.cursor.SendUpdate(buf, WindowIdCursor, SlotIdCursor)
	}

	// Inform client of operation status.
	proto.WriteWindowTransaction(buf, windowId, txId, accepted)

	player.TransmitPacket(buf.Bytes())
}

func (player *Player) PacketWindowTransaction(windowId WindowId, txId TxId, accepted bool) {
	// TODO investigate when this packet is sent from the client and what it
	// means when it does get sent.
}

func (player *Player) PacketSignUpdate(position *BlockXyz, lines [4]string) {
}

func (player *Player) PacketDisconnect(reason string) {
	log.Printf("Player %s disconnected reason=%s", player.name, reason)
	player.game.Enqueue(func(game IGame) {
		game.RemovePlayer(player)
	})
	player.txQueue <- nil
	player.mainQueue <- nil
	player.conn.Close()
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

func (player *Player) transmitLoop() {
	for {
		bs, ok := <-player.txQueue

		if !ok || bs == nil {
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

func (player *Player) TransmitPacket(packet []byte) {
	if packet == nil {
		return // skip empty packets
	}
	player.txQueue <- packet
}

func (player *Player) runQueuedCall(f func(IPlayer)) {
	player.lock.Lock()
	defer player.lock.Unlock()
	f(player)
}

func (player *Player) mainLoop() {
	expVarPlayerConnectionCount.Add(1)
	defer expVarPlayerDisconnectionCount.Add(1)

	player.chunkSubs.Init(player.shardConnecter, player.EntityId, player, &player.position)
	defer player.chunkSubs.Close()

	player.postLogin()

	for {
		f, ok := <-player.mainQueue
		if !ok || f == nil {
			return
		}
		player.runQueuedCall(f)
	}
}

// Enqueue queues a function to run with the player lock within the player's
// mainloop.
func (player *Player) Enqueue(f func(IPlayer)) {
	if f == nil {
		return
	}
	player.mainQueue <- f
}

// WithLock runs a function with the player lock within the calling goroutine.
func (player *Player) WithLock(f func(IPlayer)) {
	player.lock.Lock()
	defer player.lock.Unlock()
	f(player)
}

// Used to receive items picked up from chunks. It is synchronous so that the
// passed item can be looked at by the caller afterwards to see if it has been
// consumed.
func (player *Player) OfferItem(item *slot.Slot) {
	player.lock.Lock()
	defer player.lock.Unlock()

	player.inventory.PutItem(item)

	return
}

// OpenWindow queues a request that the player opens the given window type.
// TODO this should be passed an appropriate *Inventory for inventories that
// are tied to the world (particularly for chests).
func (player *Player) OpenWindow(invTypeId InvTypeId, inventory interface{}) {
	player.Enqueue(func(_ IPlayer) {
		player.closeCurrentWindow(true)
		window := player.inventory.NewWindow(invTypeId, player.nextWindowId, inventory)
		if window == nil {
			return
		}

		buf := &bytes.Buffer{}
		if err := window.WriteWindowOpen(buf); err != nil {
			window.Finalize(false)
			return
		}
		if err := window.WriteWindowItems(buf); err != nil {
			window.Finalize(false)
			return
		}
		player.TransmitPacket(buf.Bytes())

		player.curWindow = window
		if player.nextWindowId >= WindowIdFreeMax {
			player.nextWindowId = WindowIdFreeMin
		} else {
			player.nextWindowId++
		}
	})
}

// closeCurrentWindow closes any open window. It must be called with
// player.lock held.
func (player *Player) closeCurrentWindow(sendClosePacket bool) {
	if player.curWindow != nil {
		player.curWindow.Finalize(sendClosePacket)
	}
	player.curWindow = nil
}

// Blocks until essential login packets have been transmitted.
func (player *Player) postLogin() {
	// TODO Old version of chunkSubscriptions.Move() that was called here had
	// stuff for a callback when the nearest chunks had been sent so that player
	// position would only be sent when nearby chunks were out. Some replacement
	// for this will be needed. Possibly a message could be queued to the current
	// shard following on from chunkSubscriptions's initialization that would ask
	// the shard to send out the following packets - this would result in them
	// being sent at least after the chunks that are in the current shard have
	// been sent.

	// Send player start position etc.
	buf := new(bytes.Buffer)
	proto.ServerWritePlayerPositionLook(
		buf,
		&player.position, player.position.Y+StanceNormal,
		&player.look, false)
	player.inventory.WriteWindowItems(buf)
	packet := buf.Bytes()

	// Enqueue on the shard as a hacky way to defer the packet send until after
	// the initial chunk data has been sent.
	player.chunkSubs.curShard.Enqueue(func() {
		player.TransmitPacket(packet)
	})
}

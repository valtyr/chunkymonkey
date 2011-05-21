package interfaces

import (
	"io"
	"os"

	"chunkymonkey/entity"
	"chunkymonkey/gamerules"
	"chunkymonkey/itemtype"
	"chunkymonkey/slot"
	. "chunkymonkey/types"
)

type IPlayer interface {
	// Safe to call from outside of player's own goroutine.
	GetEntityId() EntityId
	GetEntity() *entity.Entity // Only the game mainloop may modify the return value
	GetName() string           // Do not modify return value
	LockedGetChunkPosition() ChunkXz
	TransmitPacket(packet []byte)
	// Offers an item to the player. If the player completely consumes
	// it, then item.Count will be 0 afterwards. This function is called from
	// the item's parent chunk's goroutine, so all methods are safely
	// accessible.
	OfferItem(item *slot.Slot)
	OpenWindow(invTypeId InvTypeId, inventory interface{})

	Enqueue(f func(IPlayer))
	WithLock(f func(IPlayer))

	// Everything below must be called from within Enqueue or WithLock.

	SendSpawn(writer io.Writer) (err os.Error)
	IsWithin(p1, p2 *ChunkXz) bool
	GetHeldItemType() *itemtype.ItemType
	TakeOneHeldItem(into *slot.Slot)
}

type IGame interface {
	// Safe to call from outside of Enqueue:
	GetGameRules() *gamerules.GameRules

	Enqueue(f func(IGame))

	// Everything below must be called from within Enqueue

	AddPlayer(player IPlayer)
	RemovePlayer(player IPlayer)
	MulticastPacket(packet []byte, except interface{})
	SendChatMessage(message string)
}

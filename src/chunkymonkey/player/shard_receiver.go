package player

import (
	"chunkymonkey/proto"
	"chunkymonkey/slot"
	. "chunkymonkey/types"
)

// playerShardReceiver receives events from chunk shards and acts upon them. It
// implements stub.IPlayerConnection.
type playerShardReceiver struct {
	player *Player
}

func (psr *playerShardReceiver) Init(player *Player) {
	psr.player = player
}

func (psr *playerShardReceiver) GetEntityId() EntityId {
	return psr.player.EntityId
}

func (psr *playerShardReceiver) TransmitPacket(packet []byte) {
	psr.player.TransmitPacket(packet)
}

func (psr *playerShardReceiver) ReqInventorySubscribed(block BlockXyz, invTypeId InvTypeId, slots []proto.WindowSlot) {
	psr.player.Enqueue(func(_ *Player) {
		psr.player.reqInventorySubscribed(&block, invTypeId, slots)
	})
}

func (psr *playerShardReceiver) ReqInventorySlotUpdate(block BlockXyz, slot slot.Slot, slotId SlotId) {
	psr.player.Enqueue(func(_ *Player) {
		psr.player.reqInventorySlotUpdate(&block, &slot, slotId)
	})
}

func (psr *playerShardReceiver) ReqInventoryProgressUpdate(block BlockXyz, prgBarId PrgBarId, value PrgBarValue) {
	psr.player.Enqueue(func(_ *Player) {
		psr.player.reqInventoryProgressUpdate(&block, prgBarId, value)
	})
}

func (psr *playerShardReceiver) ReqInventoryCursorUpdate(block BlockXyz, cursor slot.Slot) {
	psr.player.Enqueue(func(_ *Player) {
		psr.player.reqInventoryCursorUpdate(&block, &cursor)
	})
}

func (psr *playerShardReceiver) ReqInventoryTxState(block BlockXyz, txId TxId, accepted bool) {
	psr.player.Enqueue(func(_ *Player) {
		psr.player.reqInventoryTxState(&block, txId, accepted)
	})
}

func (psr *playerShardReceiver) ReqInventoryUnsubscribed(block BlockXyz) {
	psr.player.Enqueue(func(_ *Player) {
		psr.player.reqInventoryUnsubscribed(&block)
	})
}

func (psr *playerShardReceiver) ReqPlaceHeldItem(target BlockXyz, wasHeld slot.Slot) {
	psr.player.Enqueue(func(_ *Player) {
		psr.player.reqPlaceHeldItem(&target, &wasHeld)
	})
}

func (psr *playerShardReceiver) ReqOfferItem(fromChunk ChunkXz, entityId EntityId, item slot.Slot) {
	psr.player.Enqueue(func(_ *Player) {
		psr.player.reqOfferItem(&fromChunk, entityId, &item)
	})
}

func (psr *playerShardReceiver) ReqGiveItem(atPosition AbsXyz, item slot.Slot) {
	psr.player.Enqueue(func(_ *Player) {
		psr.player.reqGiveItem(&atPosition, &item)
	})
}

package gamerules

import (
	"os"

	"chunkymonkey/nbtutil"
	. "chunkymonkey/types"
	"nbt"
)

// blockInventory is the data stored in Chunk.SetTileEntity by some block
// aspects that contain inventories. It also implements IInventorySubscriber to
// relay events to player(s) subscribed to the inventories.
type blockInventory struct {
	// TODO Think about only keeping a reference to the parent chunk and the
	// block position and removing use of BlockInstance (some values inside it
	// will change within the chunk independently of this instance, which is
	// misleading).
	instance           BlockInstance

	inv                IInventory
	subscribers        map[EntityId]IPlayerClient
	ejectOnUnsubscribe bool
	invTypeId          InvTypeId
	block              BlockXyz
}

// newBlockInventory creates a new blockInventory.
func newBlockInventory(instance *BlockInstance, inv IInventory, ejectOnUnsubscribe bool, invTypeId InvTypeId) *blockInventory {
	blkInv := &blockInventory{
		inv:                inv,
		subscribers:        make(map[EntityId]IPlayerClient),
		ejectOnUnsubscribe: ejectOnUnsubscribe,
		invTypeId:          invTypeId,
	}

	if instance != nil {
		blkInv.instance = *instance
	}

	blkInv.inv.SetSubscriber(blkInv)

	return blkInv
}

func (blkInv *blockInventory) ReadNbt(tag nbt.ITag) (err os.Error) {
	if err = blkInv.inv.ReadNbt(tag); err != nil {
		return
	}

	if blkInv.instance.BlockLoc, err = nbtutil.ReadBlockXyzCompound(tag); err != nil {
		return
	}

	return nil
}

func (blkInv *blockInventory) WriteNbt() nbt.ITag {
	// TODO
	return nil
}

func (blkInv *blockInventory) SetChunk(chunk IChunkBlock) {
	blkInv.instance.Chunk = chunk
}

func (blkInv *blockInventory) Block() BlockXyz {
	return blkInv.instance.BlockLoc
}

func (blkInv *blockInventory) Click(player IPlayerClient, click *Click) {
	txState := blkInv.inv.Click(click)

	player.InventoryCursorUpdate(blkInv.instance.BlockLoc, click.Cursor)

	// Inform client of operation status.
	player.InventoryTxState(blkInv.instance.BlockLoc, click.TxId, txState == TxStateAccepted)
}

func (blkInv *blockInventory) SlotUpdate(slot *Slot, slotId SlotId) {
	for _, subscriber := range blkInv.subscribers {
		subscriber.InventorySlotUpdate(blkInv.instance.BlockLoc, *slot, slotId)
	}
}

func (blkInv *blockInventory) ProgressUpdate(prgBarId PrgBarId, value PrgBarValue) {
	for _, subscriber := range blkInv.subscribers {
		subscriber.InventoryProgressUpdate(blkInv.instance.BlockLoc, prgBarId, value)
	}
}

func (blkInv *blockInventory) AddSubscriber(player IPlayerClient) {
	entityId := player.GetEntityId()
	blkInv.subscribers[entityId] = player

	// Register self for automatic removal when IPlayerClient unsubscribes
	// from the chunk.
	blkInv.instance.Chunk.AddOnUnsubscribe(entityId, blkInv)

	slots := blkInv.inv.MakeProtoSlots()

	player.InventorySubscribed(blkInv.instance.BlockLoc, blkInv.invTypeId, slots)
}

func (blkInv *blockInventory) RemoveSubscriber(entityId EntityId) {
	blkInv.subscribers[entityId] = nil, false
	blkInv.instance.Chunk.RemoveOnUnsubscribe(entityId, blkInv)
	if blkInv.ejectOnUnsubscribe && len(blkInv.subscribers) == 0 {
		blkInv.EjectItems()
	}
}

func (blkInv *blockInventory) Destroyed() {
	for _, subscriber := range blkInv.subscribers {
		subscriber.InventoryUnsubscribed(blkInv.instance.BlockLoc)
		blkInv.instance.Chunk.RemoveOnUnsubscribe(subscriber.GetEntityId(), blkInv)
	}
	blkInv.subscribers = nil
}

// Unsubscribed implements IUnsubscribed. It removes a player's
// subscription to the inventory when they unsubscribe from the chunk.
func (blkInv *blockInventory) Unsubscribed(entityId EntityId) {
	blkInv.subscribers[entityId] = nil, false
}

// EjectItems removes all items from the inventory and drops them at the
// location of the block.
func (blkInv *blockInventory) EjectItems() {
	items := blkInv.inv.TakeAllItems()

	for _, slot := range items {
		spawnItemInBlock(&blkInv.instance, slot.ItemTypeId, slot.Count, slot.Data)
	}
}

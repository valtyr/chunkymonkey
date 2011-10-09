package gamerules

import (
	"os"

	"chunkymonkey/types"
	"nbt"
)

// blockInventory is the data stored in Chunk.SetTileEntity by some block
// aspects that contain inventories. It also implements IInventorySubscriber to
// relay events to player(s) subscribed to the inventories.
type blockInventory struct {
	tileEntity
	inv                IInventory
	subscribers        map[types.EntityId]IPlayerClient
	ejectOnUnsubscribe bool
	invTypeId          types.InvTypeId
}

// newBlockInventory creates a new blockInventory.
func newBlockInventory(instance *BlockInstance, inv IInventory, ejectOnUnsubscribe bool, invTypeId types.InvTypeId) *blockInventory {
	blkInv := &blockInventory{
		inv:                inv,
		subscribers:        make(map[types.EntityId]IPlayerClient),
		ejectOnUnsubscribe: ejectOnUnsubscribe,
		invTypeId:          invTypeId,
	}

	if instance != nil {
		blkInv.chunk = instance.Chunk
		blkInv.blockLoc = instance.BlockLoc
	}

	blkInv.inv.SetSubscriber(blkInv)

	return blkInv
}

func (blkInv *blockInventory) UnmarshalNbt(tag *nbt.Compound) (err os.Error) {
	if err = blkInv.tileEntity.UnmarshalNbt(tag); err != nil {
		return
	}

	if err = blkInv.inv.UnmarshalNbt(tag); err != nil {
		return
	}

	return nil
}

func (blkInv *blockInventory) MarshalNbt(tag *nbt.Compound) (err os.Error) {
	if err = blkInv.tileEntity.MarshalNbt(tag); err != nil {
		return
	}

	if err = blkInv.inv.MarshalNbt(tag); err != nil {
		return
	}

	return nil
}

func (blkInv *blockInventory) Click(player IPlayerClient, click *Click) {
	txState := blkInv.inv.Click(click)

	player.InventoryCursorUpdate(blkInv.blockLoc, click.Cursor)

	// Inform client of operation status.
	player.InventoryTxState(blkInv.blockLoc, click.TxId, txState == types.TxStateAccepted)
}

func (blkInv *blockInventory) SlotUpdate(slot *Slot, slotId types.SlotId) {
	for _, subscriber := range blkInv.subscribers {
		subscriber.InventorySlotUpdate(blkInv.blockLoc, *slot, slotId)
	}
}

func (blkInv *blockInventory) ProgressUpdate(prgBarId types.PrgBarId, value types.PrgBarValue) {
	for _, subscriber := range blkInv.subscribers {
		subscriber.InventoryProgressUpdate(blkInv.blockLoc, prgBarId, value)
	}
}

func (blkInv *blockInventory) AddSubscriber(player IPlayerClient) {
	entityId := player.GetEntityId()
	blkInv.subscribers[entityId] = player

	// Register self for automatic removal when IPlayerClient unsubscribes
	// from the chunk.
	blkInv.chunk.AddOnUnsubscribe(entityId, blkInv)

	slots := blkInv.inv.MakeProtoSlots()

	player.InventorySubscribed(blkInv.blockLoc, blkInv.invTypeId, slots)
}

func (blkInv *blockInventory) RemoveSubscriber(entityId types.EntityId) {
	blkInv.subscribers[entityId] = nil, false
	blkInv.chunk.RemoveOnUnsubscribe(entityId, blkInv)
	if blkInv.ejectOnUnsubscribe && len(blkInv.subscribers) == 0 {
		blkInv.EjectItems()
	}
}

func (blkInv *blockInventory) Destroyed() {
	for _, subscriber := range blkInv.subscribers {
		subscriber.InventoryUnsubscribed(blkInv.blockLoc)
		blkInv.chunk.RemoveOnUnsubscribe(subscriber.GetEntityId(), blkInv)
	}
	blkInv.subscribers = nil
}

// Unsubscribed implements IUnsubscribed. It removes a player's
// subscription to the inventory when they unsubscribe from the chunk.
func (blkInv *blockInventory) Unsubscribed(entityId types.EntityId) {
	blkInv.subscribers[entityId] = nil, false
}

// EjectItems removes all items from the inventory and drops them at the
// location of the block.
func (blkInv *blockInventory) EjectItems() {
	items := blkInv.inv.TakeAllItems()

	for _, slot := range items {
		spawnItemInBlock(blkInv.chunk, blkInv.blockLoc, slot.ItemTypeId, slot.Count, slot.Data)
	}
}

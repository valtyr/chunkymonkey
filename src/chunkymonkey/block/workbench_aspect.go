package block

import (
	"chunkymonkey/inventory"
	"chunkymonkey/slot"
	"chunkymonkey/stub"
	. "chunkymonkey/types"
)

func makeWorkbenchAspect() (aspect IBlockAspect) {
	return &WorkbenchAspect{}
}

// WorkbenchAspect is the behaviour for the workbench block that allows 3x3
// crafting.
type WorkbenchAspect struct {
	StandardAspect
}

func (aspect *WorkbenchAspect) Name() string {
	return "Workbench"
}

func (aspect *WorkbenchAspect) Hit(instance *BlockInstance, player stub.IPlayerConnection, digStatus DigStatus) (destroyed bool) {
	destroyed = aspect.StandardAspect.Hit(instance, player, digStatus)
	if destroyed {
		aspect.ejectItems(instance)
	}
	return
}

func (aspect *WorkbenchAspect) Interact(instance *BlockInstance, player stub.IPlayerConnection) {
	extra, ok := instance.Chunk.BlockExtra(&instance.SubLoc).(*workbenchExtra)
	if !ok {
		ejectItems := func() {
			aspect.ejectItems(instance)
		}
		inv := inventory.NewWorkbenchInventory(instance.Chunk.RecipeSet())
		extra = newWorkbenchExtra(&instance.BlockLoc, instance.Chunk, inv, ejectItems)
		instance.Chunk.SetBlockExtra(&instance.SubLoc, extra)
	}

	extra.AddSubscriber(player)
}

func (aspect *WorkbenchAspect) Click(instance *BlockInstance, player stub.IPlayerConnection, cursor *slot.Slot, rightClick bool, shiftClick bool, slotId SlotId) {
	extra, ok := instance.Chunk.BlockExtra(&instance.SubLoc).(*workbenchExtra)
	if !ok {
		// TODO send transaction failure, maybe send the cursor state unchanged
		// right back?
		player.ReqInventoryCursorUpdate(instance.BlockLoc, *cursor)
		return
	}

	extra.inv.Click(slotId, cursor, rightClick, shiftClick)
	player.ReqInventoryCursorUpdate(instance.BlockLoc, *cursor)
}

func (aspect *WorkbenchAspect) ejectItems(instance *BlockInstance) {
	workbenchInv, ok := instance.Chunk.BlockExtra(&instance.SubLoc).(*inventory.WorkbenchInventory)
	if !ok {
		return
	}

	items := workbenchInv.TakeAllItems()
	for _, slot := range items {
		spawnItemInBlock(instance, slot.ItemType, slot.Count, slot.Data)
	}
}


// workbenchExtra is the data stored in Chunk.SetBlockExtra. It also implements
// IInventorySubscriber to relay events to player(s) subscribed.
type workbenchExtra struct {
	blockLoc       BlockXyz
	chunk          IChunkBlock
	inv            *inventory.WorkbenchInventory
	subscribers    map[EntityId]stub.IPlayerConnection
	onUnsubscribed func()
}

func newWorkbenchExtra(blockLoc *BlockXyz, chunk IChunkBlock, inv *inventory.WorkbenchInventory, onUnsubscribed func()) *workbenchExtra {
	extra := &workbenchExtra{
		blockLoc:       *blockLoc,
		chunk:          chunk,
		inv:            inv,
		subscribers:    make(map[EntityId]stub.IPlayerConnection),
		onUnsubscribed: onUnsubscribed,
	}

	inv.SetSubscriber(extra)

	return extra
}

func (extra *workbenchExtra) SlotUpdate(slot *slot.Slot, slotId SlotId) {
	for _, subscriber := range extra.subscribers {
		subscriber.ReqInventorySlotUpdate(extra.blockLoc, *slot, slotId)
	}
}

func (extra *workbenchExtra) AddSubscriber(player stub.IPlayerConnection) {
	entityId := player.GetEntityId()
	extra.subscribers[entityId] = player

	// Register self for automatic removal when IPlayerConnection unsubscribes
	// from the chunk.
	extra.chunk.AddOnUnsubscribe(entityId, extra)

	slots := extra.inv.MakeProtoSlots()

	player.ReqInventorySubscribed(extra.blockLoc, InvTypeIdWorkbench, slots)
}

func (extra *workbenchExtra) RemoveSubscriber(entityId EntityId) {
	extra.subscribers[entityId] = nil, false
	extra.chunk.RemoveOnUnsubscribe(entityId, extra)
	if len(extra.subscribers) == 0 && extra.onUnsubscribed != nil {
		extra.onUnsubscribed()
	}
}

func (extra *workbenchExtra) Destroyed() {
	for _, subscriber := range extra.subscribers {
		subscriber.ReqInventoryUnsubscribed(extra.blockLoc)
		extra.chunk.RemoveOnUnsubscribe(subscriber.GetEntityId(), extra)
	}
	extra.subscribers = nil
}

// Unsubscribed implements block.IUnsubscribed. It removes a player's
// subscription to the inventory when they unsubscribe from the chunk.
func (extra *workbenchExtra) Unsubscribed(entityId EntityId) {
	extra.subscribers[entityId] = nil, false
}

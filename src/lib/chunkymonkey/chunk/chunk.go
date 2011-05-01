// Map chunks

package chunk

import (
	"bytes"
	"log"
	"rand"
	"sync"
	"time"

	"chunkymonkey/block"
	. "chunkymonkey/interfaces"
	"chunkymonkey/item"
	"chunkymonkey/itemtype"
	"chunkymonkey/proto"
	"chunkymonkey/chunkstore"
	. "chunkymonkey/types"
)

const (
	// Assumed values for size of player axis-aligned bounding box (AAB).
	playerAabH = AbsCoord(0.75) // Each side of player.
	playerAabY = AbsCoord(2.00) // From player's feet position upwards.
)


// A chunk is slice of the world map
type Chunk struct {
	mainQueue     chan func(IChunk)
	mgr           *ChunkManager
	loc           ChunkXz
	blocks        []byte
	blockData     []byte
	blockLight    []byte
	skyLight      []byte
	heightMap     []byte
	items         map[EntityId]*item.Item
	rand          *rand.Rand
	neighbours    neighboursCache
	cachedPacket  []byte                       // Cached packet data for this block.
	subscribers   map[IChunkSubscriber]bool    // Subscribers getting updates from the chunk.
	subscriberPos map[IChunkSubscriber]*AbsXyz // Player positions that are near or in the chunk.
}

func newChunkFromReader(reader chunkstore.ChunkReader, mgr *ChunkManager) (chunk *Chunk) {
	chunk = &Chunk{
		mainQueue:     make(chan func(IChunk), 256),
		mgr:           mgr,
		loc:           *reader.ChunkLoc(),
		blocks:        reader.Blocks(),
		blockData:     reader.BlockData(),
		skyLight:      reader.SkyLight(),
		blockLight:    reader.BlockLight(),
		heightMap:     reader.HeightMap(),
		items:         make(map[EntityId]*item.Item),
		rand:          rand.New(rand.NewSource(time.UTC().Seconds())),
		subscribers:   make(map[IChunkSubscriber]bool),
		subscriberPos: make(map[IChunkSubscriber]*AbsXyz),
	}
	chunk.neighbours.init()
	go chunk.mainLoop()
	return
}

func blockIndex(subLoc *SubChunkXyz) (index int32, shift byte, ok bool) {
	if subLoc.X < 0 || subLoc.Y < 0 || subLoc.Z < 0 || subLoc.X >= ChunkSizeH || subLoc.Y >= ChunkSizeY || subLoc.Z >= ChunkSizeH {
		ok = false
		index = 0
	} else {
		ok = true

		index = int32(subLoc.Y) + (int32(subLoc.Z) * ChunkSizeY) + (int32(subLoc.X) * ChunkSizeY * ChunkSizeH)

		if index%2 == 0 {
			// Low nibble
			shift = 0
		} else {
			// High nibble
			shift = 4
		}
	}
	return
}

// Sets a block and its data. Returns true if the block was not changed.
func (chunk *Chunk) setBlock(blockLoc *BlockXyz, subLoc *SubChunkXyz, index int32, shift byte, blockType BlockId, blockMetadata byte) {

	// Invalidate cached packet
	chunk.cachedPacket = nil

	chunk.blocks[index] = byte(blockType)

	mask := byte(0x0f) << shift
	twoBlockData := chunk.blockData[index/2]
	twoBlockData = ((blockMetadata << shift) & mask) | (twoBlockData & ^mask)
	chunk.blockData[index/2] = twoBlockData

	// Tell players that the block changed
	packet := &bytes.Buffer{}
	proto.WriteBlockChange(packet, blockLoc, blockType, blockMetadata)
	chunk.multicastSubscribers(packet.Bytes())

	// Update neighbour caches of this change
	chunk.neighbours.setBlock(subLoc, blockType)

	return
}

func (chunk *Chunk) GetLoc() *ChunkXz {
	return &chunk.loc
}

func (chunk *Chunk) Enqueue(f func(IChunk)) {
	chunk.mainQueue <- f
}

func (chunk *Chunk) mainLoop() {
	for {
		f := <-chunk.mainQueue
		f(chunk)
	}
}

func (chunk *Chunk) GetRand() *rand.Rand {
	return chunk.rand
}

func (chunk *Chunk) GetItemType(itemTypeId ItemTypeId) (itemType *itemtype.ItemType, ok bool) {
	itemType, ok = chunk.mgr.itemTypes[itemTypeId]
	return
}

func (chunk *Chunk) AddItem(item *item.Item) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	chunk.mgr.game.Enqueue(func(game IGame) {
		entity := item.GetEntity()
		game.AddEntity(entity)
		chunk.items[entity.EntityId] = item
		wg.Done()
	})
	wg.Wait()

	// Spawn new item for players
	buf := &bytes.Buffer{}
	item.SendSpawn(buf)
	chunk.multicastSubscribers(buf.Bytes())
}

func (chunk *Chunk) TransferItem(item *item.Item) {
	chunk.items[item.GetEntity().EntityId] = item
}

func (chunk *Chunk) GetBlock(subLoc *SubChunkXyz) (blockType BlockId, ok bool) {
	index, _, ok := blockIndex(subLoc)
	if !ok {
		return
	}

	blockType = BlockId(chunk.blocks[index])

	return
}

func (chunk *Chunk) DigBlock(subLoc *SubChunkXyz, digStatus DigStatus) (ok bool) {
	index, shift, ok := blockIndex(subLoc)
	if !ok {
		return
	}

	blockTypeId := BlockId(chunk.blocks[index])
	blockLoc := chunk.loc.ToBlockXyz(subLoc)

	if blockType, ok := chunk.mgr.blockTypes.Get(blockTypeId); ok && blockType.Destructable {
		if blockType.Aspect.Dig(chunk, blockLoc, digStatus) {
			chunk.setBlock(blockLoc, subLoc, index, shift, BlockIdAir, 0)
		}
	} else {
		log.Printf("Attempted to destroy unknown block Id %d", blockTypeId)
		ok = false
	}

	return
}

func (chunk *Chunk) PlaceBlock(againstLoc *BlockXyz, againstFace Face, blockId BlockId) (ok bool) {
	// Check if the block being built against allows such placement (E.g
	// fences, water, etc. do not).
	againstBlockType, _, blockUnknownId := chunk.blockQuery(againstLoc)
	if ok = !blockUnknownId; !ok {
		return
	}
	if ok = againstBlockType.Attachable; !ok {
		return
	}

	// The position to put the block at.
	dx, dy, dz := againstFace.GetDxyz()
	placeAtLoc := &BlockXyz{
		againstLoc.X + dx,
		againstLoc.Y + dy,
		againstLoc.Z + dz,
	}
	placeChunkLoc, placeSubLoc := placeAtLoc.ToChunkLocal()
	if ok = (placeChunkLoc.X == chunk.loc.X && placeChunkLoc.Z == chunk.loc.Z); !ok {
		log.Print("Chunk/PlaceBlock: block not inside this chunk")
		return
	}
	index, shift, ok := blockIndex(placeSubLoc)
	if !ok {
		log.Print("Chunk/PlaceBlock: invalid position within chunk")
		return
	}

	// Blocks can only replace certain blocks.
	blockTypeId := BlockId(chunk.blocks[index])
	blockType, ok := chunk.mgr.blockTypes.Get(blockTypeId)
	if !ok {
		return
	}
	if ok = blockType.Replaceable; !ok {
		return
	}

	placeBlockLoc := chunk.loc.ToBlockXyz(placeSubLoc)
	// TODO block metadata
	chunk.setBlock(placeBlockLoc, placeSubLoc, index, shift, blockId, 0)
	ok = true

	return
}

// Used to read the BlockId of a block that's either in the chunk, or
// immediately adjoining it in a neighbouring chunk via the side caches.
func (chunk *Chunk) blockQuery(blockLoc *BlockXyz) (blockType *block.BlockType, isWithinChunk bool, blockUnknownId bool) {
	chunkLoc, subLoc := blockLoc.ToChunkLocal()

	var blockTypeId BlockId
	var ok bool

	if chunkLoc.X == chunk.loc.X && chunkLoc.Z == chunk.loc.Z {
		// The item is asking about this chunk.
		blockTypeId, _ = chunk.GetBlock(subLoc)
		isWithinChunk = true
	} else {
		// The item is asking about a separate chunk.
		isWithinChunk = false

		ok, blockTypeId = chunk.neighbours.GetCachedBlock(
			chunk.loc.X-chunkLoc.X,
			chunk.loc.Z-chunkLoc.Z,
			subLoc,
		)

		if !ok {
			// The chunk side isn't cached or isn't a neighbouring block.
			blockUnknownId = true
			return
		}
	}

	blockType, ok = chunk.mgr.blockTypes.Get(blockTypeId)
	if !ok {
		log.Printf(
			"Chunk/blockQuery found unknown block type Id %d at %+v",
			blockTypeId, blockLoc,
		)
		blockUnknownId = true
	}

	return
}

func (chunk *Chunk) Tick() {
	// Update neighbouring chunks of block changes in this chunk
	chunk.neighbours.flush()

	blockQuery := func(blockLoc *BlockXyz) (isSolid bool, isWithinChunk bool) {
		blockType, isWithinChunk, blockUnknownId := chunk.blockQuery(blockLoc)
		if blockUnknownId {
			// If we are in doubt, we assume that the block asked about is
			// solid (this way items don't fly off the side of the map
			// needlessly).
			isSolid = true
		} else {
			isSolid = blockType.Solid
		}
		return
	}

	destroyedEntityIds := []EntityId{}
	leftItems := []*item.Item{}

	for _, item := range chunk.items {
		if item.Tick(blockQuery) {
			if item.GetPosition().Y <= 0 {
				// Item fell out of the world
				destroyedEntityIds = append(
					destroyedEntityIds, item.GetEntity().EntityId)
			} else {
				leftItems = append(leftItems, item)
			}
		}
	}

	if len(leftItems) > 0 {
		// Remove items from this chunk
		for _, item := range leftItems {
			chunk.items[item.GetEntity().EntityId] = nil, false
		}

		// Send items to new chunk
		chunk.mgr.game.Enqueue(func(game IGame) {
			mgr := game.GetChunkManager()
			for _, item := range leftItems {
				chunkLoc := item.GetPosition().ToChunkXz()
				blockChunk := mgr.Get(chunkLoc)
				blockChunk.Enqueue(func(blockChunk IChunk) {
					blockChunk.AddItem(item)
				})
			}
		})
	}

	if len(destroyedEntityIds) > 0 {
		buf := &bytes.Buffer{}
		for _, entityId := range destroyedEntityIds {
			proto.WriteEntityDestroy(buf, entityId)
			chunk.items[entityId] = nil, false
		}
		chunk.multicastSubscribers(buf.Bytes())
	}
}

func (chunk *Chunk) AddSubscriber(subscriber IChunkSubscriber) {
	chunk.subscribers[subscriber] = true
	subscriber.TransmitPacket(chunk.chunkPacket())

	// Send spawns of all items in the chunk
	if len(chunk.items) > 0 {
		buf := &bytes.Buffer{}
		for _, item := range chunk.items {
			item.SendSpawn(buf)
		}
		subscriber.TransmitPacket(buf.Bytes())
	}
}

func (chunk *Chunk) RemoveSubscriber(subscriber IChunkSubscriber, sendPacket bool) {
	chunk.subscribers[subscriber] = false, false
	if sendPacket {
		buf := &bytes.Buffer{}
		proto.WritePreChunk(buf, &chunk.loc, ChunkUnload)
		subscriber.TransmitPacket(buf.Bytes())
	}
}

func (chunk *Chunk) multicastSubscribers(packet []byte) {
	for subscriber, _ := range chunk.subscribers {
		subscriber.TransmitPacket(packet)
	}
}

func (chunk *Chunk) SetSubscriberPosition(subscriber IChunkSubscriber, pos *AbsXyz) {
	chunk.subscriberPos[subscriber] = pos, pos != nil
	if pos != nil {
		// Does the subscriber overlap with any items?
		minX := pos.X - playerAabH
		maxX := pos.X + playerAabH
		minZ := pos.Z - playerAabH
		maxZ := pos.Z + playerAabH
		minY := pos.Y
		maxY := pos.Y + playerAabY

		for entityId, item := range chunk.items {
			// TODO This check should be performed when items move as well.
			pos := item.GetPosition()
			if pos.X >= minX && pos.X <= maxX && pos.Y >= minY && pos.Y <= maxY && pos.Z >= minZ && pos.Z <= maxZ {
				slot := item.GetSlot()
				subscriber.OfferItem(slot)
				if slot.Count == 0 {
					// The item has been accepted and completely consumed.

					buf := &bytes.Buffer{}

					// Tell all subscribers to animate the item flying at the
					// subscriber.
					proto.WriteItemCollect(buf, entityId, subscriber.GetEntityId())

					// Tell all subscribers that the item's entity is
					// destroyed.
					proto.WriteEntityDestroy(buf, entityId)

					chunk.multicastSubscribers(buf.Bytes())

					chunk.items[entityId] = nil, false
				}

				// TODO Check for how to properly handle partially consumed
				// items? Probably not high priority since all dropped items
				// have a count of 1 at the moment. Might need to respawn the
				// item with a new count. Do the clients even care what the
				// count is or if it changes? Or if an item is "collected" but
				// still exists?
			}
		}
	}
}

func (chunk *Chunk) chunkPacket() []byte {
	if chunk.cachedPacket == nil {
		buf := &bytes.Buffer{}
		proto.WriteMapChunk(buf, &chunk.loc, chunk.blocks, chunk.blockData, chunk.blockLight, chunk.skyLight)
		chunk.cachedPacket = buf.Bytes()
	}

	return chunk.cachedPacket
}

func (chunk *Chunk) SendUpdate() {
	buf := &bytes.Buffer{}
	for _, item := range chunk.items {
		item.SendUpdate(buf)
	}
	chunk.multicastSubscribers(buf.Bytes())
}

func (chunk *Chunk) sideCacheSetNeighbour(side ChunkSideDir, neighbour *Chunk) {
	chunk.neighbours.sideCacheSetNeighbour(side, neighbour, chunk.blocks)
}

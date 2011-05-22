package shardserver

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"rand"
	"time"

	"chunkymonkey/block"
	"chunkymonkey/chunkstore"
	"chunkymonkey/item"
	"chunkymonkey/itemtype"
	"chunkymonkey/mob"
	"chunkymonkey/proto"
	"chunkymonkey/recipe"
	"chunkymonkey/slot"
	"chunkymonkey/shardserver_external"
	. "chunkymonkey/types"
)

var enableMobs = flag.Bool(
	"enableMobs", false, "EXPERIMENTAL: spawn mobs.")

// A chunk is slice of the world map.
type Chunk struct {
	mgr        *LocalShardManager
	shard      *ChunkShard
	loc        ChunkXz
	blocks     []byte
	blockData  []byte
	blockLight []byte
	skyLight   []byte
	heightMap  []byte
	// spawn are typically mobs or items.
	// TODO: (discuss) Maybe split this back into mobs and items?
	// There are many more users of "spawn" than of only mobs or items. So
	// I'm inclined to leave it as is.
	spawn        map[EntityId]shardserver_external.INonPlayerSpawn
	blockExtra   map[BlockIndex]interface{} // Used by IBlockAspect to store private specific data.
	rand         *rand.Rand
	neighbours   neighboursCache
	cachedPacket []byte                                              // Cached packet data for this block.
	subscribers  map[EntityId]shardserver_external.IPlayerConnection // Players getting updates from the chunk.
	playersData  map[EntityId]*playerData                            // Some player data for player(s) in the chunk.
}

func newChunkFromReader(reader chunkstore.IChunkReader, mgr *LocalShardManager, shard *ChunkShard) (chunk *Chunk) {
	chunk = &Chunk{
		mgr:         mgr,
		shard:       shard,
		loc:         *reader.ChunkLoc(),
		blocks:      reader.Blocks(),
		blockData:   reader.BlockData(),
		skyLight:    reader.SkyLight(),
		blockLight:  reader.BlockLight(),
		heightMap:   reader.HeightMap(),
		spawn:       make(map[EntityId]shardserver_external.INonPlayerSpawn),
		blockExtra:  make(map[BlockIndex]interface{}),
		rand:        rand.New(rand.NewSource(time.UTC().Seconds())),
		subscribers: make(map[EntityId]shardserver_external.IPlayerConnection),
		playersData: make(map[EntityId]*playerData),
	}
	chunk.neighbours.init()
	return
}

func (chunk *Chunk) String() string {
	return fmt.Sprintf("Chunk[%d,%d]", chunk.loc.X, chunk.loc.Z)
}

// Sets a block and its data. Returns true if the block was not changed.
func (chunk *Chunk) setBlock(blockLoc *BlockXyz, subLoc *SubChunkXyz, index BlockIndex, blockType BlockId, blockMetadata byte) {

	// Invalidate cached packet.
	chunk.cachedPacket = nil

	index.SetBlockId(chunk.blocks, blockType)
	index.SetBlockData(chunk.blockData, blockMetadata)

	chunk.blockExtra[index] = nil, false

	// Tell players that the block changed.
	packet := &bytes.Buffer{}
	proto.WriteBlockChange(packet, blockLoc, blockType, blockMetadata)
	chunk.MulticastPlayers(-1, packet.Bytes())

	// Update neighbour caches of this change.
	chunk.neighbours.setBlock(subLoc, blockType)

	return
}

func (chunk *Chunk) GetRand() *rand.Rand {
	return chunk.rand
}

func (chunk *Chunk) GetItemType(itemTypeId ItemTypeId) (itemType *itemtype.ItemType, ok bool) {
	itemType, ok = chunk.mgr.gameRules.ItemTypes[itemTypeId]
	return
}

func (chunk *Chunk) TransferSpawn(s shardserver_external.INonPlayerSpawn) {
	chunk.spawn[s.GetEntity().EntityId] = s
}

// AddSpawn creates a mob or item in this chunk and notifies the new spawn to
// all chunk subscribers.
func (chunk *Chunk) AddSpawn(s shardserver_external.INonPlayerSpawn) {
	e := s.GetEntity()
	chunk.mgr.entityMgr.AddEntity(e)

	// Spawn new item/mob for players.
	buf := &bytes.Buffer{}
	s.SendSpawn(buf)
	chunk.MulticastPlayers(-1, buf.Bytes())
}

func (chunk *Chunk) removeSpawn(s shardserver_external.INonPlayerSpawn) {
	e := s.GetEntity()
	chunk.mgr.entityMgr.RemoveEntity(e)
	chunk.spawn[e.EntityId] = nil, false
	// Tell all subscribers that the spawn's entity is destroyed.
	buf := &bytes.Buffer{}
	proto.WriteEntityDestroy(buf, e.EntityId)
	chunk.MulticastPlayers(-1, buf.Bytes())
}

func (chunk *Chunk) GetBlockExtra(subLoc *SubChunkXyz) interface{} {
	if index, ok := subLoc.BlockIndex(); ok {
		if extra, ok := chunk.blockExtra[index]; ok {
			return extra
		}
	}
	return nil
}

func (chunk *Chunk) SetBlockExtra(subLoc *SubChunkXyz, extra interface{}) {
	if index, ok := subLoc.BlockIndex(); ok {
		chunk.blockExtra[index] = extra, extra != nil
	}
}

func (chunk *Chunk) getBlockIndexByBlockXyz(blockLoc *BlockXyz) (index BlockIndex, subLoc *SubChunkXyz, ok bool) {
	chunkLoc, subLoc := blockLoc.ToChunkLocal()

	if chunkLoc.X != chunk.loc.X || chunkLoc.Z != chunk.loc.Z {
		log.Printf(
			"%v.getBlockIndexByBlockXyz: position (%T%#v) is not within chunk",
			chunk, blockLoc, blockLoc)
		return 0, nil, false
	}

	index, ok = subLoc.BlockIndex()
	if !ok {
		log.Printf(
			"%v.getBlockIndexByBlockXyz: invalid position (%T%#v) within chunk",
			chunk, blockLoc, blockLoc)
	}

	return
}

func (chunk *Chunk) GetBlock(subLoc *SubChunkXyz) (blockType BlockId, ok bool) {
	index, ok := subLoc.BlockIndex()
	if !ok {
		return
	}

	blockType = index.GetBlockId(chunk.blocks)

	return
}

func (chunk *Chunk) GetRecipeSet() *recipe.RecipeSet {
	return chunk.mgr.gameRules.Recipes
}

func (chunk *Chunk) hitBlock(player shardserver_external.IPlayerConnection, held slot.Slot, digStatus DigStatus, target *BlockXyz, face Face) {

	index, subLoc, ok := chunk.getBlockIndexByBlockXyz(target)
	if !ok {
		return
	}

	blockTypeId := index.GetBlockId(chunk.blocks)

	if blockType, ok := chunk.mgr.gameRules.BlockTypes.Get(blockTypeId); ok && blockType.Destructable {
		blockData := index.GetBlockData(chunk.blockData)
		blockLoc := chunk.loc.ToBlockXyz(subLoc)

		blockInstance := &block.BlockInstance{
			Chunk:    chunk,
			BlockLoc: *blockLoc,
			SubLoc:   *subLoc,
			Data:     blockData,
		}
		if blockType.Aspect.Hit(blockInstance, player, digStatus) {
			chunk.setBlock(blockLoc, subLoc, index, BlockIdAir, 0)
		}
	} else {
		log.Printf("%v.HitBlock: Attempted to destroy unknown block Id %d", chunk, blockTypeId)
	}

	return
}

func (chunk *Chunk) interactBlock(player shardserver_external.IPlayerConnection, held slot.Slot, target *BlockXyz, againstFace Face) {
	// TODO use held item to better check of if the player is trying to place a
	// block vs. perform some other interaction (e.g hoeing dirt).
	// RequestPlaceItem() will need to check again as the item might have changed
	// in the meantime, of course.

	index, subLoc, ok := chunk.getBlockIndexByBlockXyz(target)
	if !ok {
		return
	}

	blockTypeId := BlockId(chunk.blocks[index])
	blockType, ok := chunk.mgr.gameRules.BlockTypes.Get(blockTypeId)
	if !ok {
		log.Printf(
			"%v.PlayerBlockInteract: unknown target block type %d at target position (%#v)",
			chunk, blockTypeId, target)
		return
	}

	if blockType.Attachable {
		// The player is interacting with a block that can be attached to.

		// Work out the position to put the block at.
		// TODO check for overflow, especially in Y.
		dx, dy, dz := againstFace.GetDxyz()
		destLoc := BlockXyz{
			target.X + dx,
			target.Y + dy,
			target.Z + dz,
		}

		player.RequestPlaceHeldItem(destLoc)
	} else {
		// Player is otherwise interacting with the block.
		blockInstance := &block.BlockInstance{
			Chunk:    chunk,
			BlockLoc: *target,
			SubLoc:   *subLoc,
			Data:     index.GetBlockData(chunk.blockData),
		}
		blockType.Aspect.Interact(blockInstance, player)
	}

	return
}

// placeBlock attempts to place a block. This is called by PlayerBlockInteract
// in the situation where the player interacts with an attachable block
// (potentially in a different chunk to the one where the block gets placed).
func (chunk *Chunk) placeItem(player shardserver_external.IPlayerConnection, target BlockXyz, slot slot.Slot) {
	// TODO defer a check for remaining items in slot, and do something with them
	// (send to player or drop on the ground).

	index, subLoc, ok := chunk.getBlockIndexByBlockXyz(&target)
	if !ok {
		return
	}

	// Blocks can only replace certain blocks.
	blockTypeId := index.GetBlockId(chunk.blocks)
	blockType, ok := chunk.mgr.gameRules.BlockTypes.Get(blockTypeId)
	if !ok || !blockType.Replaceable {
		return
	}

	// TODO more flexible item checking for block placement (e.g placing seed
	// items on farmland doesn't fit this current simplistic model). The block
	// type for the block being placed against should probably contain this logic
	// (i.e farmland block should know about the seed item).
	if slot.Count < 1 || slot.ItemType == nil || slot.ItemType.Id < BlockIdMin || slot.ItemType.Id > BlockIdMax {
		// Not a placeable item.
		return
	}

	// TODO block metadata
	chunk.setBlock(&target, subLoc, index, BlockId(slot.ItemType.Id), 0)

	slot.Decrement()
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

	blockType, ok = chunk.mgr.gameRules.BlockTypes.Get(blockTypeId)
	if !ok {
		log.Printf(
			"%v.blockQuery found unknown block type Id %d at %+v",
			chunk, blockTypeId, blockLoc)
		blockUnknownId = true
	}

	return
}

func (chunk *Chunk) tick() {
	// Update neighbouring chunks of block changes in this chunk
	chunk.neighbours.flush()

	blockQuery := func(blockLoc *BlockXyz) (isSolid bool, isWithinChunk bool) {
		blockType, isWithinChunk, blockUnknownId := chunk.blockQuery(blockLoc)
		if blockUnknownId {
			// If we are in doubt, we assume that the block asked about is
			// solid (this way objects don't fly off the side of the map
			// needlessly).
			isSolid = true
		} else {
			isSolid = blockType.Solid
		}
		return
	}
	outgoingSpawns := []shardserver_external.INonPlayerSpawn{}

	for _, e := range chunk.spawn {
		if e.Tick(blockQuery) {
			if e.Position().Y <= 0 {
				// Item or mob fell out of the world.
				chunk.removeSpawn(e)
			} else {
				outgoingSpawns = append(outgoingSpawns, e)
			}
		}
	}

	if len(outgoingSpawns) > 0 {
		// Transfer spawns to new chunk.
		for _, e := range outgoingSpawns {
			// Remove mob/items from this chunk.
			chunk.spawn[e.GetEntity().EntityId] = nil, false

			// Transfer to other chunk.
			chunkLoc := e.Position().ToChunkXz()

			// TODO Batch spawns up into a request per shard if there are efficiency
			// concerns in sending them individually.
			chunk.mgr.EnqueueOnChunk(chunkLoc, func(blockChunk shardserver_external.IChunk) {
				blockChunk.TransferSpawn(e)
			})
		}
	}

	// XXX: Testing hack. If player is in a chunk with no mobs, spawn a pig.
	if *enableMobs {
		for _, playerData := range chunk.playersData {
			loc := playerData.position.ToChunkXz()
			if chunk.isSameChunk(&loc) {
				ms := chunk.mobs()
				if len(ms) == 0 {
					log.Printf("%v.Tick: spawning a mob at %v", chunk, playerData.position)
					m := mob.NewPig(&playerData.position, &AbsVelocity{5, 5, 5})
					chunk.AddSpawn(&m.Mob)
				}
				break
			}
		}
	}
}

func (chunk *Chunk) mobs() (s []*mob.Mob) {
	s = make([]*mob.Mob, 0, 3)
	for _, e := range chunk.spawn {
		switch e.(type) {
		case *mob.Mob:
			s = append(s, e.(*mob.Mob))
		}
	}
	return
}

func (chunk *Chunk) items() (s []*item.Item) {
	s = make([]*item.Item, 0, 10)
	for _, e := range chunk.spawn {
		switch e.(type) {
		case *item.Item:
			s = append(s, e.(*item.Item))
		}
	}
	return
}

func (chunk *Chunk) addPlayer(entityId EntityId, player shardserver_external.IPlayerConnection) {
	chunk.subscribers[entityId] = player

	buf := new(bytes.Buffer)
	proto.WritePreChunk(buf, &chunk.loc, ChunkInit)
	player.TransmitPacket(buf.Bytes())

	player.TransmitPacket(chunk.chunkPacket())

	// Send spawns of all mobs/items in the chunk.
	if len(chunk.spawn) > 0 {
		buf := new(bytes.Buffer)
		for _, e := range chunk.spawn {
			e.SendSpawn(buf)
		}
		player.TransmitPacket(buf.Bytes())
	}

	// Spawn existing players for new player.
	if len(chunk.playersData) > 0 {
		playersPacket := new(bytes.Buffer)
		for _, existing := range chunk.playersData {
			if existing.entityId != entityId {
				existing.sendSpawn(playersPacket)
			}
		}
		player.TransmitPacket(playersPacket.Bytes())
	}
}

func (chunk *Chunk) removePlayer(entityId EntityId, sendPacket bool) {
	player, ok := chunk.subscribers[entityId]

	if ok && sendPacket {
		chunk.subscribers[entityId] = nil, false
		buf := &bytes.Buffer{}
		proto.WritePreChunk(buf, &chunk.loc, ChunkUnload)
		// TODO send PacketEntityDestroy packets for spawns in this chunk.
		player.TransmitPacket(buf.Bytes())
	}
}

func (chunk *Chunk) MulticastPlayers(exclude EntityId, packet []byte) {
	for entityId, player := range chunk.subscribers {
		if entityId != exclude {
			player.TransmitPacket(packet)
		}
	}
}

func (chunk *Chunk) addPlayerData(entityId EntityId, name string, pos AbsXyz, look LookBytes, held ItemTypeId) {
	// TODO add other initial data in here.
	newPlayerData := &playerData{
		entityId:   entityId,
		name:       name,
		position:   pos,
		look:       look,
		heldItemId: held,
	}
	chunk.playersData[entityId] = newPlayerData

	// Spawn new player for existing players.
	newPlayerPacket := new(bytes.Buffer)
	newPlayerData.sendSpawn(newPlayerPacket)
	chunk.MulticastPlayers(entityId, newPlayerPacket.Bytes())
}

func (chunk *Chunk) removePlayerData(entityId EntityId) {
	chunk.playersData[entityId] = nil, false
}

func (chunk *Chunk) setPlayerPositionLook(entityId EntityId, pos AbsXyz, look LookBytes, moved bool) {
	data, ok := chunk.playersData[entityId]

	if !ok {
		log.Printf(
			"%v.setPlayerPosition: called for EntityId (%d) not present as playerData.",
			chunk, entityId,
		)
		return
	}

	data.position = pos
	data.look = look

	// Update subscribers.
	buf := new(bytes.Buffer)
	data.sendPositionLook(buf)
	chunk.MulticastPlayers(entityId, buf.Bytes())

	if moved {
		/* TODO
		// Does the player overlap with any items?
		for _, item := range chunk.items() {
			// TODO This check should be performed when items move as well.
			if data.OverlapsItem(item) {
				slot := item.GetSlot()
				player.OfferItem(slot)
				if slot.Count == 0 {
					// The item has been accepted and completely consumed.

					buf := &bytes.Buffer{}

					// Tell all subscribers to animate the item flying at the
					// player.
					proto.WriteItemCollect(buf, EntityId(entityId), player.GetEntityId())
					chunk.MulticastPlayers(-1, buf.Bytes())
					chunk.removeSpawn(item)
				}

				// TODO Check for how to properly handle partially consumed
				// items? Probably not high priority since all dropped items
				// have a count of 1 at the moment. Might need to respawn the
				// item with a new count. Do the clients even care what the
				// count is or if it changes? Or if an item is "collected" but
				// still exists?
			}
		}
		*/
	}
}

func (chunk *Chunk) chunkPacket() []byte {
	if chunk.cachedPacket == nil {
		buf := new(bytes.Buffer)
		proto.WriteMapChunk(buf, &chunk.loc, chunk.blocks, chunk.blockData, chunk.blockLight, chunk.skyLight)
		chunk.cachedPacket = buf.Bytes()
	}

	return chunk.cachedPacket
}

func (chunk *Chunk) sendUpdate() {
	buf := &bytes.Buffer{}
	for _, e := range chunk.spawn {
		e.SendUpdate(buf)
	}
	chunk.MulticastPlayers(-1, buf.Bytes())
}

func (chunk *Chunk) sideCacheSetNeighbour(side ChunkSideDir, neighbour *Chunk) {
	chunk.neighbours.sideCacheSetNeighbour(side, neighbour, chunk.blocks)
}

func (chunk *Chunk) isSameChunk(otherChunkLoc *ChunkXz) bool {
	return otherChunkLoc.X == chunk.loc.X && otherChunkLoc.Z == chunk.loc.Z
}

func (chunk *Chunk) EnqueueGeneric(fn func()) {
	chunk.shard.enqueueRequest(&runGeneric{fn})
}

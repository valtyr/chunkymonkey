package shardserver

import (
	"sync"

	"chunkymonkey/chunkstore"
	"chunkymonkey/entity"
	"chunkymonkey/gamerules"
	"chunkymonkey/shardserver_external"
	. "chunkymonkey/types"
)

// localShardConnection implements IShardConnection for LocalShardManager.
type localShardConnection struct {
	entityId EntityId
	player   shardserver_external.ITransmitter
	shard    *ChunkShard
}

func newLocalShardConnection(entityId EntityId, player shardserver_external.ITransmitter, shard *ChunkShard) *localShardConnection {
	return &localShardConnection{
		entityId: entityId,
		player:   player,
		shard:    shard,
	}
}

func (conn *localShardConnection) Disconnect() {
	// TODO This inefficiently unsubscribes from all chunks, even if not
	// subscribed to.
	conn.shard.EnqueueAllChunks(func(chunk shardserver_external.IChunk) {
		chunk.RemovePlayer(conn.entityId, false)
	})
}

func (conn *localShardConnection) Enqueue(fn func()) {
	conn.shard.Enqueue(fn)
}

func (conn *localShardConnection) SubscribeChunk(chunkLoc ChunkXz) {
	conn.shard.EnqueueOnChunk(chunkLoc, func(chunk shardserver_external.IChunk) {
		chunk.AddPlayer(conn.entityId, conn.player)
	})
}

func (conn *localShardConnection) UnsubscribeChunk(chunkLoc ChunkXz) {
	conn.shard.EnqueueOnChunk(chunkLoc, func(chunk shardserver_external.IChunk) {
		chunk.RemovePlayer(conn.entityId, true)
	})
}

func (conn *localShardConnection) MulticastPlayers(chunkLoc ChunkXz, exclude EntityId, packet []byte) {
	conn.shard.EnqueueOnChunk(chunkLoc, func(chunk shardserver_external.IChunk) {
		chunk.(*Chunk).MulticastPlayers(exclude, packet)
	})
}

// LocalShardManager contains all chunk shards and can look them up. It
// implements IShardConnecter and is for use in hosting all shards in the local
// process.
type LocalShardManager struct {
	entityMgr  *entity.EntityManager
	chunkStore chunkstore.IChunkStore
	gameRules  *gamerules.GameRules
	shards     map[uint64]*ChunkShard
	lock       sync.Mutex
}

func NewLocalShardManager(chunkStore chunkstore.IChunkStore, entityMgr *entity.EntityManager, gameRules *gamerules.GameRules) *LocalShardManager {
	return &LocalShardManager{
		entityMgr:  entityMgr,
		chunkStore: chunkStore,
		gameRules:  gameRules,
		shards:     make(map[uint64]*ChunkShard),
	}
}

func (mgr *LocalShardManager) getShard(loc ShardXz) *ChunkShard {
	shardKey := loc.Key()
	if shard, ok := mgr.shards[shardKey]; ok {
		// Shard already exists.
		return shard
	}

	// Create shard.
	shard := NewChunkShard(mgr, loc)
	mgr.shards[shardKey] = shard
	go shard.serve()

	return shard
}

func (mgr *LocalShardManager) ShardConnect(entityId EntityId, player shardserver_external.ITransmitter, shardLoc ShardXz) shardserver_external.IShardConnection {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	shard := mgr.getShard(shardLoc)
	return newLocalShardConnection(entityId, player, shard)
}

// TODO remove Enqueue* methods

// EnqueueAllChunks runs a given function on all loaded chunks.
func (mgr *LocalShardManager) EnqueueAllChunks(fn func(chunk shardserver_external.IChunk)) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	for _, shard := range mgr.shards {
		shard.EnqueueAllChunks(fn)
	}
}

// EnqueueOnChunk runs a function on the chunk at the given location. If the
// chunk does not exist, it does nothing.
func (mgr *LocalShardManager) EnqueueOnChunk(loc ChunkXz, fn func(chunk shardserver_external.IChunk)) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	shard := mgr.getShard(loc.ToShardXz())
	shard.EnqueueOnChunk(loc, fn)
}

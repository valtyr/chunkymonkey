package shardserver

import (
	"chunkymonkey/gamerules"
	. "chunkymonkey/interfaces"
	"chunkymonkey/chunkstore"
	. "chunkymonkey/types"
)

// localShardConnection implements IShardConnection for LocalShardManager.
type localShardConnection struct {
	entityId EntityId
	player   ITransmitter
	shard    *ChunkShard
}

func newLocalShardConnection(entityId EntityId, player ITransmitter, shard *ChunkShard) *localShardConnection {
	return &localShardConnection{
		entityId: entityId,
		player:   player,
		shard:    shard,
	}
}

func (conn *localShardConnection) SubscribeChunk(chunkLoc ChunkXz) {
	conn.shard.EnqueueOnChunk(chunkLoc, func(chunk IChunk) {
		chunk.AddPlayer(conn.entityId, conn.player)
	})
}

func (conn *localShardConnection) UnsubscribeChunk(chunkLoc ChunkXz) {
	conn.shard.EnqueueOnChunk(chunkLoc, func(chunk IChunk) {
		chunk.RemovePlayer(conn.entityId, true)
	})
}

func (conn *localShardConnection) Enqueue(fn func()) {
	conn.shard.Enqueue(fn)
}

func (conn *localShardConnection) Disconnect() {
	// TODO This inefficiently unsubscribes from all chunks, even if not
	// subscribed to.
	conn.shard.EnqueueAllChunks(func(chunk IChunk) {
		chunk.RemovePlayer(conn.entityId, false)
	})
}

// LocalShardManager contains all chunk shards and can look them up. It
// implements IShardConnecter and is for use in hosting all shards in the local
// process.
type LocalShardManager struct {
	game       IGame
	chunkStore chunkstore.IChunkStore
	gameRules  *gamerules.GameRules
	shards     map[uint64]*ChunkShard
}

func NewLocalShardManager(chunkStore chunkstore.IChunkStore, game IGame) *LocalShardManager {
	return &LocalShardManager{
		game:       game,
		chunkStore: chunkStore,
		gameRules:  game.GetGameRules(),
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

func (mgr *LocalShardManager) ShardConnect(entityId EntityId, player ITransmitter, shardLoc ShardXz) IShardConnection {
	shard := mgr.getShard(shardLoc)
	return newLocalShardConnection(entityId, player, shard)
}

// EnqueueAllChunks runs a given function on all loaded chunks.
func (mgr *LocalShardManager) EnqueueAllChunks(fn func(chunk IChunk)) {
	mgr.game.Enqueue(func(_ IGame) {
		for _, shard := range mgr.shards {
			shard.EnqueueAllChunks(fn)
		}
	})
}

// EnqueueOnChunk runs a function on the chunk at the given location. If the
// chunk does not exist, it does nothing.
func (mgr *LocalShardManager) EnqueueOnChunk(loc ChunkXz, fn func(chunk IChunk)) {
	mgr.game.Enqueue(func(_ IGame) {
		shard := mgr.getShard(loc.ToShardXz())
		shard.EnqueueOnChunk(loc, fn)
	})
}

package chunk

import (
	"chunkymonkey/gamerules"
	. "chunkymonkey/interfaces"
	"chunkymonkey/chunkstore"
	. "chunkymonkey/types"
)

// ChunkManager contains all chunk shards and can look them up.
// TODO Rename to ShardManager.
type ChunkManager struct {
	game       IGame
	chunkStore chunkstore.IChunkStore
	gameRules  *gamerules.GameRules
	shards     map[uint64]*ChunkShard
}

func NewChunkManager(chunkStore chunkstore.IChunkStore, game IGame) *ChunkManager {
	return &ChunkManager{
		game:       game,
		chunkStore: chunkStore,
		gameRules:  game.GetGameRules(),
		shards:     make(map[uint64]*ChunkShard),
	}
}

func (mgr *ChunkManager) getShard(loc ShardXz) *ChunkShard {
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

// EnqueueAllChunks runs a given function on all loaded chunks.
func (mgr *ChunkManager) EnqueueAllChunks(fn func(chunk IChunk)) {
	mgr.game.Enqueue(func(_ IGame) {
		for _, shard := range mgr.shards {
			shard.EnqueueAllChunks(fn)
		}
	})
}

// EnqueueOnChunk runs a function on the chunk at the given location. If the
// chunk does not exist, it does nothing.
func (mgr *ChunkManager) EnqueueOnChunk(loc ChunkXz, fn func(chunk IChunk)) {
	mgr.game.Enqueue(func(_ IGame) {
		shard := mgr.getShard(loc.ToShardXz())
		shard.EnqueueOnChunk(loc, fn)
	})
}

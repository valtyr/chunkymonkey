package chunk

import (
	. "chunkymonkey/types"
)

const (
	// Each shard is ShardSize * ShardSize chunks square.
	ShardSize      = 16
	chunksPerShard = ShardSize * ShardSize
)

type ShardCoord int32

type ShardXz struct {
	X, Z ShardCoord
}

func (loc *ShardXz) toChunkXz() ChunkXz {
	return ChunkXz{
		X: ChunkCoord(loc.X * ShardSize),
		Z: ChunkCoord(loc.Z * ShardSize),
	}
}

type ChunkShard struct {
	mgr            *ChunkManager
	loc            ShardXz
	originChunkLoc ChunkXz // The lowest X and Z located chunk in the shard.
	farChunkLoc    ChunkXz // The highest X and Z located chunk located just beyond the shard.
	chunks         [chunksPerShard]*Chunk
}

func NewChunkShard(mgr *ChunkManager, loc ShardXz) (shard *ChunkShard) {
	farShardLoc := loc
	farShardLoc.X++
	farShardLoc.Z++

	shard = &ChunkShard{
		mgr:            mgr,
		loc:            loc,
		originChunkLoc: loc.toChunkXz(),
		farChunkLoc:    farShardLoc.toChunkXz(),
	}

	return
}

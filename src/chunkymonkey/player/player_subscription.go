package player

import (
	"log"

	"chunkymonkey/shardserver_external"
	. "chunkymonkey/types"
)

// shardRef holds a reference to a shard connection and context for the number
// of subscribed chunks inside the shard.
type shardRef struct {
	shard shardserver_external.IShardConnection
	count int
}

// chunkSubscriptions is part of the player frontend code, and maintains:
// * the shards to be connected to,
// * the chunks that to be subscribed to (via their shards),
// * and moving the player from one shard to another.
type chunkSubscriptions struct {
	connecter   shardserver_external.IShardConnecter
	player      shardserver_external.ITransmitter
	entityId    EntityId
	curShardLoc ShardXz                               // Shard the player is currently in.
	curChunkLoc ChunkXz                               // Chunk the player is currently in.
	curShard    shardserver_external.IShardConnection // Shard the player is hosted on.
	shards      map[uint64]shardRef                   // Connections to shards.
}

func (sub *chunkSubscriptions) Init(connecter shardserver_external.IShardConnecter, entityId EntityId, player shardserver_external.ITransmitter, initialPos *AbsXyz) {
	sub.connecter = connecter
	sub.player = player
	sub.entityId = entityId
	sub.curShardLoc = initialPos.ToShardXz()
	sub.curChunkLoc = initialPos.ToChunkXz()
	sub.shards = make(map[uint64]shardRef)

	initialChunkLocs := orderedChunkSquare(sub.curChunkLoc, ChunkRadius)
	sub.subscribeToChunks(initialChunkLocs)

	sub.curShard = sub.shards[sub.curShardLoc.Key()].shard
}

func (sub *chunkSubscriptions) Move(newLoc *AbsXyz) {
	newChunkLoc := newLoc.ToChunkXz()
	if newChunkLoc.X == sub.curChunkLoc.X && newChunkLoc.Z == sub.curChunkLoc.Z {
		return
	}
	sub.moveToChunk(newChunkLoc)

	newShardLoc := newLoc.ToShardXz()
	if newShardLoc.X == sub.curShardLoc.X && newShardLoc.Z == sub.curShardLoc.Z {
		return
	}
	sub.moveToShard(newShardLoc)
}

// Close closes down all shard connections. Use when the player is
// disconnected.
func (sub *chunkSubscriptions) Close() {
	for key, ref := range sub.shards {
		ref.shard.Disconnect()
		sub.shards[key] = shardRef{}, false
	}
}

// subscribeToChunks connects to shards and subscribes to chunks for the chunk
// locations given.
func (sub *chunkSubscriptions) subscribeToChunks(chunkLocs []ChunkXz) {
	for _, chunkLoc := range chunkLocs {
		shardLoc := chunkLoc.ToShardXz()
		shardKey := shardLoc.Key()
		ref, ok := sub.shards[shardKey]
		if !ok {
			ref = shardRef{
				shard: sub.connecter.ShardConnect(sub.entityId, sub.player, shardLoc),
				count: 0,
			}
			sub.shards[shardKey] = ref
		}
		ref.shard.SubscribeChunk(chunkLoc)
		ref.count++
	}
}

// unsubscribeFromChunks unsubscribes from chunks for the chunk locations
// given, and disconnects from shards where there are no subscribed chunks.
func (sub *chunkSubscriptions) unsubscribeFromChunks(chunkLocs []ChunkXz) {
	for _, chunkLoc := range chunkLocs {
		shardLoc := chunkLoc.ToShardXz()
		shardKey := shardLoc.Key()
		if ref, ok := sub.shards[shardKey]; ok {
			ref.shard.UnsubscribeChunk(chunkLoc)
			ref.count--
			if ref.count <= 0 {
				ref.shard.Disconnect()
				sub.shards[shardKey] = shardRef{}, false
			}
		} else {
			// Odd - we don't have a shard connection for that chunk.
			log.Print("chunkSubscriptions.unsubscribeFromChunks() attempted to " +
				"unsubscribe from chunk in unconnected shard.")
		}
	}
}

// moveToChunk subscribes to chunks that are newly in range, and unsubscribes
// to those that have just left.
func (sub *chunkSubscriptions) moveToChunk(newChunkLoc ChunkXz) {
	addChunkLocs := squareDifference(newChunkLoc, sub.curChunkLoc, ChunkRadius)
	sub.subscribeToChunks(addChunkLocs)

	delChunkLocs := squareDifference(sub.curChunkLoc, newChunkLoc, ChunkRadius)
	sub.unsubscribeFromChunks(delChunkLocs)
}

func (sub *chunkSubscriptions) moveToShard(newShardLoc ShardXz) {
	// The new current shard is assumed to be present in sub.shards already.
	sub.curShard = sub.shards[newShardLoc.Key()].shard
}

// squareDifference computes the ChunkXz values that are in "square radius" of
// centerA, but not in "square radius" of centerB.
//
// E.g, input squares:
// centerA={1,1}, centerB={2,2}, radius = 1
// A = in square A, B = in square B, C = in both
//
// AAA
// ACCB
// ACCB
//  BBB
//
// Results will be:
//
// AAA
// A
// A
func squareDifference(centerA, centerB ChunkXz, radius ChunkCoord) []ChunkXz {
	// TODO Currently this is the exact same slow simple O(n²) algorithm used in
	// the test to produce the expected result. This could be optimized somewhat.
	// Even if a general optimization is too much effort, the trivial cases of
	// moving ±1X and/or ±1Z are the very common cases, and the simple dumb
	// algorithm can be used otherwise.
	areaEdgeSize := radius*2 + 1
	result := make([]ChunkXz, 0, areaEdgeSize)
	for x := centerA.X - radius; x <= centerA.X+radius; x++ {
		for z := centerA.Z - radius; z <= centerA.Z+radius; z++ {
			if x >= centerB.X-radius && x <= centerB.X+radius && z >= centerB.Z-radius && z <= centerB.Z+radius {
				// {x, z} is within square B. Don't include this.
				continue
			}
			result = append(result, ChunkXz{x, z})
		}
	}
	return result
}

// orderedChunkSquare creates a slice of chunk locations in a square centered
// on `center`, with sides `radius` chunks away from the center. The chunk
// locations are output in approx this order for radius=2 (where lower numbered
// chunks are earlier):
//
// 33333
// 32223
// 32123
// 32223
// 33333
func orderedChunkSquare(center ChunkXz, radius ChunkCoord) (locs []ChunkXz) {
	areaEdgeSize := radius*2 + 1
	locs = make([]ChunkXz, areaEdgeSize*areaEdgeSize)
	locs[0] = center
	index := 1
	for curRadius := ChunkCoord(1); curRadius <= radius; curRadius++ {
		xMin := ChunkCoord(-curRadius + center.X)
		xMax := ChunkCoord(curRadius + center.X)
		zMin := ChunkCoord(-curRadius + center.Z)
		zMax := ChunkCoord(curRadius + center.Z)

		// Northern and southern rows of chunks.
		for x := xMin; x <= xMax; x++ {
			locs[index] = ChunkXz{x, zMin}
			index++
			locs[index] = ChunkXz{x, zMax}
			index++
		}

		// Eastern and western columns (except for where they intersect the
		// north and south rows).
		for z := zMin + 1; z < zMax; z++ {
			locs[index] = ChunkXz{xMin, z}
			index++
			locs[index] = ChunkXz{xMax, z}
			index++
		}
	}
	return
}

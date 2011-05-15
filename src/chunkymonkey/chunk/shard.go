package chunk

import (
	"fmt"
	"log"

	"chunkymonkey/chunkstore"
	. "chunkymonkey/interfaces"
	. "chunkymonkey/types"
)

const chunksPerShard = ShardSize * ShardSize

// chunkXzToChunkIndex assumes that locDelta is offset relative to the shard
// origin.
func chunkXzToChunkIndex(locDelta *ChunkXz) int {
	return int(locDelta.X)*ShardSize + int(locDelta.Z)
}

// ChunkShard represents a square shard of chunks that share a master
// goroutine.
type ChunkShard struct {
	mgr            *LocalShardManager
	loc            ShardXz
	originChunkLoc ChunkXz // The lowest X and Z located chunk in the shard.
	chunks         [chunksPerShard]*Chunk
	requests       chan iShardRequest
}

func NewChunkShard(mgr *LocalShardManager, loc ShardXz) (shard *ChunkShard) {
	shard = &ChunkShard{
		mgr:            mgr,
		loc:            loc,
		originChunkLoc: loc.ToChunkXz(),
		requests:       make(chan iShardRequest, 256),
	}

	return
}

// serve services shard requests in the foreground.
func (shard *ChunkShard) serve() {
	for {
		request := <-shard.requests
		request.perform(shard)
	}
}

func (shard *ChunkShard) String() string {
	return fmt.Sprintf("ChunkShard[%#v/%#v]", shard.loc, shard.originChunkLoc)
}

// Get returns the Chunk at at given coordinates, loading it if it is not
// already loaded.
func (shard *ChunkShard) Get(loc *ChunkXz) IChunk {
	locDelta := ChunkXz{
		X: loc.X - shard.originChunkLoc.X,
		Z: loc.Z - shard.originChunkLoc.Z,
	}

	if locDelta.X < 0 || locDelta.Z < 0 || locDelta.X > ShardSize || locDelta.Z > ShardSize {
		log.Printf("%v.Get(%#v): chunk requested from outside of shard", shard, loc)
		return nil
	}

	chunkIndex := locDelta.X*ShardSize + locDelta.Z

	chunk := shard.chunks[chunkIndex]

	// Chunk already loaded.
	if chunk != nil {
		return chunk
	}

	chunk = shard.loadChunk(loc, &locDelta)
	shard.chunks[chunkIndex] = chunk

	return chunk
}

// loadChunk loads the specified chunk from store, and returns it.
// loc - The absolute world position of the chunk.
// locDelta - The relative position of the chunk within the shard.
func (shard *ChunkShard) loadChunk(loc *ChunkXz, locDelta *ChunkXz) *Chunk {
	chunkResult := <-shard.mgr.chunkStore.LoadChunk(loc)
	chunkReader, err := chunkResult.Reader, chunkResult.Err
	if err != nil {
		if _, ok := err.(chunkstore.NoSuchChunkError); !ok {
			log.Printf("%v.load(%#v): chunk loading error: %v", shard, loc, err)
			return nil
		} else {
			// Chunk doesn't exist in store.
			// TODO Generate new chunks.
			return nil
		}
	}

	chunk := newChunkFromReader(chunkReader, shard.mgr, shard)

	// Notify neighbouring chunk(s) (if any) that this chunk is now active, and
	// notify this chunk of its active neighbours
	linkNeighbours := func(from ChunkSideDir) {
		dx, dz := from.GetDxz()
		nLoc := ChunkXz{
			X: locDelta.X + dx,
			Z: locDelta.Z + dz,
		}
		if nLoc.X < 0 || nLoc.Z < 0 || nLoc.X > ShardSize || nLoc.Z > ShardSize {
			// Link to neighbouring chunk outside the shard.
			// TODO This should also link to chunks outside the shard. Although the
			// architecure of this is likely to change radically or go away.
		} else {
			// Link to neighbouring chunk within the shard.
			chunkIndex := chunkXzToChunkIndex(locDelta)
			neighbour := shard.chunks[chunkIndex]
			if neighbour != nil {
				to := from.GetOpposite()
				chunk.sideCacheSetNeighbour(from, neighbour)
				neighbour.sideCacheSetNeighbour(to, chunk)
			}
		}
	}
	// TODO Corresponding unlinking when a chunk is unloaded.
	linkNeighbours(ChunkSideEast)
	linkNeighbours(ChunkSideSouth)
	linkNeighbours(ChunkSideWest)
	linkNeighbours(ChunkSideNorth)

	return chunk
}

// TODO Consider replacing all uses of Enqueue* to implementations of
// iShardRequest without closures. We'll probably need to do this when
// distributing anyway.

// EnqueueAllChunks runs a given function on all loaded chunks in the shard.
func (shard *ChunkShard) EnqueueAllChunks(fn func(chunk IChunk)) {
	shard.requests <- &runOnAllChunks{fn}
}

// EnqueueOnChunk runs a function on the chunk at the given location. If the
// chunk does not exist, it does nothing.
func (shard *ChunkShard) EnqueueOnChunk(loc ChunkXz, fn func(chunk IChunk)) {
	shard.requests <- &runOnChunk{loc, fn}
}

func (shard *ChunkShard) enqueueRequest(req iShardRequest) {
	shard.requests <- req
}

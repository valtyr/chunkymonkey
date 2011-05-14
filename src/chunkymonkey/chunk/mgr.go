package chunk

import (
	"log"

	"chunkymonkey/gamerules"
	. "chunkymonkey/interfaces"
	"chunkymonkey/chunkstore"
	. "chunkymonkey/types"
)

// ChunkManager contains all chunks and can look them up
type ChunkManager struct {
	game       IGame
	chunkStore chunkstore.IChunkStore
	gameRules  *gamerules.GameRules
	chunks     map[uint64]*Chunk
}

func NewChunkManager(chunkStore chunkstore.IChunkStore, game IGame) *ChunkManager {
	return &ChunkManager{
		game:       game,
		chunkStore: chunkStore,
		gameRules:  game.GetGameRules(),
		chunks:     make(map[uint64]*Chunk),
	}
}

// Get a chunk at given coordinates
func (mgr *ChunkManager) Get(loc *ChunkXz) (c IChunk) {
	var ok bool
	var chunk *Chunk
	key := loc.ChunkKey()

	if chunk, ok = mgr.chunks[key]; ok {
		c = chunk
		return
	}

	chunkReader, err := mgr.chunkStore.LoadChunk(loc)
	if err != nil {
		if _, ok := err.(chunkstore.NoSuchChunkError); !ok {
			log.Printf("ChunkManager.Get(%+v): %s", loc, err.String())
			return
		} else {
			// Chunk doesn't exist in store.
			// TODO Generate new chunks.
			return
		}
	}

	chunk = newChunkFromReader(chunkReader, mgr)
	c = chunk

	// Notify neighbouring chunk(s) (if any) that this chunk is now active, and
	// notify this chunk of its active neighbours
	linkNeighbours := func(from ChunkSideDir) {
		dx, dz := from.GetDxz()
		loc := ChunkXz{
			X: loc.X + dx,
			Z: loc.Z + dz,
		}
		neighbour, exists := mgr.chunks[loc.ChunkKey()]
		if exists {
			to := from.GetOpposite()
			chunk.Enqueue(func(_ IChunk) {
				chunk.sideCacheSetNeighbour(from, neighbour)
			})
			neighbour.Enqueue(func(_ IChunk) {
				neighbour.sideCacheSetNeighbour(to, chunk)
			})
		}
	}
	// TODO corresponding unlinking when a chunk is unloaded
	linkNeighbours(ChunkSideEast)
	linkNeighbours(ChunkSideSouth)
	linkNeighbours(ChunkSideWest)
	linkNeighbours(ChunkSideNorth)

	mgr.chunks[key] = chunk
	return
}

func (mgr *ChunkManager) EnqueueAllChunks(fn func(chunk IChunk)) {
	mgr.game.Enqueue(func(_ IGame) {
		for _, chunk := range mgr.chunks {
			chunk.Enqueue(fn)
		}
	})
}

// Enqueues a function to run on the chunk at the given location. If the chunk
// does not exist, it does nothing.
func (mgr *ChunkManager) EnqueueOnChunk(loc *ChunkXz, fn func(chunk IChunk)) {
	mgr.game.Enqueue(func(_ IGame) {
		chunk := mgr.Get(loc)
		if chunk != nil {
			chunk.Enqueue(fn)
		}
	})
}

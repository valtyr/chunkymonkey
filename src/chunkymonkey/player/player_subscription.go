package player

import (
	"bytes"
	"sync"

	. "chunkymonkey/interfaces"
	"chunkymonkey/proto"
	. "chunkymonkey/types"
)

const (
	// The radius within which player's exact location is relayed to chunks.
	// For item pickup the only values that make any sense are 0 or 1.
	positionUpdateRadius = ChunkCoord(1)
	numPosChunksEdge     = positionUpdateRadius*2 + 1
	numPosChunks         = numPosChunksEdge * numPosChunksEdge
)

// Keeps track of chunks that the player is subscribed to, and adds/removes the
// player from chunks as they move around. Also keeps nearby chunks up to date
// with the player's position.
type chunkSubscriptions struct {
	player        *Player
	chunks        map[uint64]ChunkXz // Map from ChunkKeys to chunk locations subscribed to.
	chunksWithLoc map[uint64]ChunkXz // Chunks that have previously received a position update.
	curChunk      ChunkXz            // Current chunk that the player is on.
	curChunkValid bool               // States if curChunkValid is valid.
}

func (cs *chunkSubscriptions) Init(player *Player) {
	cs.player = player
	cs.chunks = make(map[uint64]ChunkXz)
	cs.chunksWithLoc = make(map[uint64]ChunkXz)
	cs.curChunkValid = false
}

// Updates the player's location to the new position, and
// subscribes/unsubscribes to chunks as necessary.
//
// Subscribes to all chunks in MinChunkRadius, then calls nearbySent(), then
// subscribes to the remaining chunks out to ChunkRadius. Note that the bulk of
// this work happens in a seperate goroutine (including the call to nearbySent).
func (cs *chunkSubscriptions) Move(newPos *AbsXyz, nearbySent func()) {
	chunkLocsToAdd, chunkLocsToRemove, changedChunk := cs.moveChunkLocs(newPos)

	// Subscribe/unsubscribe from chunk updates.
	if len(chunkLocsToAdd) > 0 || len(chunkLocsToRemove) > 0 {
		// Remember the chunks that we're going to be subscribed to.
		for _, chunkLoc := range chunkLocsToAdd {
			cs.chunks[chunkLoc.ChunkKey()] = chunkLoc
		}
		// Forget the chunks that we're removing subscription from.
		for _, chunkLoc := range chunkLocsToRemove {
			cs.chunks[chunkLoc.ChunkKey()] = ChunkXz{}, false
		}

		// Send subscribe/unsubscribe messages in the background.
		go addRemoveChunkSubs(cs.player, cs.curChunk, nearbySent, chunkLocsToAdd, chunkLocsToRemove)
	}

	cs.updateChunksWithLoc(newPos, changedChunk)
}

// Removes all subscriptions to chunks without sending packets to "unload" the
// chunks from the client.
func (cs *chunkSubscriptions) clear() {
	player := cs.player
	for _, chunkLoc := range cs.chunks {
		player.game.GetChunkManager().EnqueueOnChunk(chunkLoc, func(chunk IChunk) {
			chunk.RemovePlayer(player, false)
		})
	}
}

// Work out which chunks to send following a move, and in which order. Also
// returns the chunks that should be removed.
func (cs *chunkSubscriptions) moveChunkLocs(newPos *AbsXyz) (chunkLocsToAdd []ChunkXz, chunkLocsToRemove []ChunkXz, changedChunk bool) {
	if cs.curChunkValid {
		// Player moving within the world. We remove old chunks and add new ones.
		var newChunkLoc ChunkXz
		newPos.UpdateChunkXz(&newChunkLoc)

		if newChunkLoc.X == cs.curChunk.X && newChunkLoc.Z == cs.curChunk.Z {
			// Still on the same chunk - nothing to be done.
			return nil, nil, false
		}

		// Very sub-optimal way to determine new chunks to load and their
		// order, but should be correct in all general cases.
		// TODO work out a more optimal method
		allChunkLocs := chunkOrder(ChunkRadius, &newChunkLoc)
		chunkLocsToAdd = make([]ChunkXz, 0, len(allChunkLocs))
		for _, chunkLoc := range allChunkLocs {
			dx := (chunkLoc.X - newChunkLoc.X).Abs()
			dz := (chunkLoc.Z - newChunkLoc.Z).Abs()
			_, haveChunk := cs.chunks[chunkLoc.ChunkKey()]
			if dx <= ChunkRadius && dz <= ChunkRadius && !haveChunk {
				chunkLocsToAdd = append(chunkLocsToAdd, chunkLoc)
			}
		}
		// Remove old chunks
		// TODO work out a more optimal method
		chunkLocsToRemove = make([]ChunkXz, 0, len(allChunkLocs))
		for _, chunkLoc := range cs.chunks {
			dx := (chunkLoc.X - newChunkLoc.X).Abs()
			dz := (chunkLoc.Z - newChunkLoc.Z).Abs()
			if dx > ChunkRadius || dz > ChunkRadius {
				chunkLocsToRemove = append(chunkLocsToRemove, chunkLoc)
			}
		}

		cs.curChunk = newChunkLoc
	} else {
		// Player arriving in the world. We send all nearby chunks.
		newPos.UpdateChunkXz(&cs.curChunk)
		cs.curChunkValid = true
		chunkLocsToAdd = chunkOrder(ChunkRadius, &cs.curChunk)
	}
	changedChunk = true
	return
}

func (cs *chunkSubscriptions) updateChunksWithLoc(newPos *AbsXyz, changedChunk bool) {
	mgr := cs.player.game.GetChunkManager()

	// Update immediately adjacent chunks with player position.
	if changedChunk {
		// Remove previously adjacent chunks.
		for key, chunkLoc := range cs.chunksWithLoc {
			dx := (chunkLoc.X - cs.curChunk.X).Abs()
			dz := (chunkLoc.Z - cs.curChunk.Z).Abs()
			if dx > positionUpdateRadius || dz > positionUpdateRadius {
				// Tell chunk to forget about player position.
				mgr.EnqueueOnChunk(chunkLoc, func(chunk IChunk) {
					chunk.SetPlayerPosition(cs.player, nil)
				})
				cs.chunksWithLoc[key] = ChunkXz{}, false
			}
		}

		// Add newly adjacent chunks.
		var cursor ChunkXz
		// Assumes that ChunkRadius > positionUpdateRadius to avoid doing
		// additional chunkmanager lookups.
		for cursor.X = cs.curChunk.X - positionUpdateRadius; cursor.X <= cs.curChunk.X+positionUpdateRadius; cursor.X++ {
			for cursor.Z = cs.curChunk.Z - positionUpdateRadius; cursor.Z <= cs.curChunk.Z+positionUpdateRadius; cursor.Z++ {
				key := cursor.ChunkKey()
				if chunk, ok := cs.chunksWithLoc[key]; !ok {
					if chunk, ok = cs.chunks[key]; ok {
						cs.chunksWithLoc[key] = chunk
					}
				}
			}
		}
	}

	// Send player position updates to adjacent chunks.
	posCopy := newPos.Copy()
	for _, chunkLoc := range cs.chunksWithLoc {
		mgr.EnqueueOnChunk(chunkLoc, func(chunk IChunk) {
			chunk.SetPlayerPosition(cs.player, posCopy)
		})
	}
}

func chunkOrder(radius ChunkCoord, center *ChunkXz) (locs []ChunkXz) {
	areaEdgeSize := radius*2 + 1
	locs = make([]ChunkXz, areaEdgeSize*areaEdgeSize)
	locs[0] = *center
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

func addRemoveChunkSubs(player *Player, curChunk ChunkXz, nearbySent func(), chunkLocsToAdd []ChunkXz, chunkLocsToRemove []ChunkXz) {
	mgr := player.game.GetChunkManager()

	if len(chunkLocsToAdd) > 0 {
		chunksSent := 0
		if nearbySent != nil {
			// Warn client to allocate memory for the chunks we're sending.
			buf := &bytes.Buffer{}
			for _, chunkLoc := range chunkLocsToAdd {
				proto.WritePreChunk(buf, &chunkLoc, ChunkInit)
			}
			player.TransmitPacket(buf.Bytes())

			// Send the important chunks.
			waitChunks := &sync.WaitGroup{}
			for _, chunkLoc := range chunkLocsToAdd {
				dx := (chunkLoc.X - curChunk.X).Abs()
				dz := (chunkLoc.Z - curChunk.Z).Abs()
				if dx > MinChunkRadius || dz > MinChunkRadius {
					// This chunk isn't important, wait until after other login
					// data has been sent before sending this and chunks
					// following.
					break
				}
				chunksSent++
				waitChunks.Add(1)
				mgr.EnqueueOnChunk(chunkLoc, func(chunk IChunk) {
					chunk.AddPlayer(player)
					waitChunks.Done()
				})
			}
			waitChunks.Wait()

			nearbySent()
		}

		// Send the remaining chunks. We don't need to wait for these to complete
		// before continuing with talking to the client.
		for _, chunkLoc := range chunkLocsToAdd[chunksSent:] {
			mgr.EnqueueOnChunk(chunkLoc, func(chunk IChunk) {
				chunk.AddPlayer(player)
			})
		}
	}

	// Unsubscribe from old chunks
	for _, chunkLoc := range chunkLocsToRemove {
		mgr.EnqueueOnChunk(chunkLoc, func(chunk IChunk) {
			chunk.RemovePlayer(player, true)
		})
	}
}

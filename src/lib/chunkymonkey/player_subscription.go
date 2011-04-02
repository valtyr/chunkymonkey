package player

import (
    "bytes"
    "sync"

    .   "chunkymonkey/interfaces"
    "chunkymonkey/proto"
    .   "chunkymonkey/types"
)

// Keeps track of chunks that the player is subscribed to, and adds/removes the
// player from chunks as they move around.
// TODO implement the movement add/remove tracking
type chunkSubscriptions struct {
    player        *Player
    chunks        map[uint64]IChunk // Map from ChunkKeys to chunks
    curChunk      ChunkXZ // Current chunk that the player is on.
    curChunkValid bool // States if curChunkValid is valid.
}

func (cs *chunkSubscriptions) init(player *Player) {
    cs.player = player
    cs.chunks = make(map[uint64]IChunk)
    cs.curChunkValid = false
}

// Updates the player's location to the new position, and
// subscribes/unsubscribes to chunks as necessary.
//
// Subscribes to all chunks in MinChunkRadius, then calls nearbySent(), then
// subscribes to the remaining chunks out to ChunkRadius. Note that the bulk of
// this work happens in a seperate goroutine (including the call to nearbySent).
func (cs *chunkSubscriptions) move(newPos *AbsXYZ, nearbySent func()) {
    chunkLocLoadOrder, chunksToRemove := cs.moveChunkLocs(newPos)

    if len(chunkLocLoadOrder) == 0 && len(chunksToRemove) == 0 {
        return
    }

    // Get all the chunks together that we're going to send.
    chunksChan := make(chan []IChunk)
    cs.player.game.Enqueue(func(game IGame) {
        chunks := make([]IChunk, 0, len(chunkLocLoadOrder))
        mgr := cs.player.game.GetChunkManager()
        for _, chunkLoc := range chunkLocLoadOrder {
            if chunk := mgr.Get(&chunkLoc); chunk != nil {
                chunks = append(chunks, chunk)
            }
        }
        chunksChan <- chunks
    })
    chunksToAdd := <-chunksChan

    // Remember the chunks that we're going to be subscribed to.
    for _, chunk := range chunksToAdd {
        cs.chunks[chunk.GetLoc().ChunkKey()] = chunk
    }
    // Forget the chunks that we're removing subscription from.
    for _, chunk := range chunksToRemove {
        cs.chunks[chunk.GetLoc().ChunkKey()] = nil, false
    }

    // Send all chunks in the background (otherwise we can deadlock on a full
    // txQueue in the player goroutine).
    go func() {
        chunksSent := 0
        if len(chunksToAdd) > 0 {
            // Warn client to allocate memory for the chunks we're sending.
            buf := &bytes.Buffer{}
            for _, chunk := range chunksToAdd {
                proto.WritePreChunk(buf, chunk.GetLoc(), ChunkInit)
            }
            cs.player.TransmitPacket(buf.Bytes())

            // Send the important chunks.
            waitChunks := &sync.WaitGroup{}
            for _, chunk := range chunksToAdd {
                chunkLoc := chunk.GetLoc()
                dx := (chunkLoc.X - cs.curChunk.X).Abs()
                dz := (chunkLoc.Z - cs.curChunk.Z).Abs()
                if dx > MinChunkRadius || dz > MinChunkRadius {
                    // This chunk isn't important, wait until after other login data
                    // has been sent before sending this and chunks following.
                    break
                }
                chunksSent++
                waitChunks.Add(1)
                chunk.Enqueue(func(chunk IChunk) {
                    chunk.AddSubscriber(cs.player)
                    waitChunks.Done()
                })
            }
            waitChunks.Wait()
        }

        if nearbySent != nil {
            nearbySent()
        }

        if len(chunksToAdd) > 0 {
            // Send the remaining chunks. We don't need to wait for these to complete
            // before continuing with talking to the client.
            for _, chunk := range chunksToAdd[chunksSent:] {
                chunk.Enqueue(func(chunk IChunk) {
                    chunk.AddSubscriber(cs.player)
                })
            }
        }

        // Unsubscribe from old chunks
        for _, chunk := range chunksToRemove {
            chunk.Enqueue(func(chunk IChunk) {
                chunk.RemoveSubscriber(cs.player, true)
            })
        }
    }()
}

// Removes all subscriptions to chunks.
func (cs *chunkSubscriptions) clear() {
    for _, chunk := range cs.chunks {
        chunk.RemoveSubscriber(cs.player, false)
    }
}

// Work out which chunks to send following a move, and in which order. Also
// returns the chunks that should be removed.
func (cs *chunkSubscriptions) moveChunkLocs(newPos *AbsXYZ) (chunkLocLoadOrder []ChunkXZ, chunksToRemove []IChunk) {
    if cs.curChunkValid {
        // Player moving within the world. We remove old chunks and add new ones.
        var newChunkLoc ChunkXZ
        newPos.UpdateChunkXZ(&newChunkLoc)

        if newChunkLoc.X == cs.curChunk.X && newChunkLoc.Z == cs.curChunk.Z {
            // Still on the same chunk - nothing to be done.
            return nil, nil
        }

        // Very sub-optimal way to determine new chunks to load and their
        // order, but should be correct in all general cases.
        // TODO work out a more optimal method
        allChunkLocs := chunkOrder(ChunkRadius, &newChunkLoc)
        chunkLocLoadOrder = make([]ChunkXZ, 0, len(allChunkLocs))
        for _, chunkLoc := range allChunkLocs {
            dx := (chunkLoc.X - newChunkLoc.X).Abs()
            dz := (chunkLoc.Z - newChunkLoc.Z).Abs()
            _, haveChunk := cs.chunks[chunkLoc.ChunkKey()]
            if dx <= ChunkRadius && dz <= ChunkRadius && !haveChunk {
                chunkLocLoadOrder = append(chunkLocLoadOrder, chunkLoc)
            }
        }
        // Remove old chunks
        // TODO work out a more optimal method
        chunksToRemove = make([]IChunk, 0, len(allChunkLocs))
        for _, chunk := range cs.chunks {
            chunkLoc := chunk.GetLoc()
            dx := (chunkLoc.X - newChunkLoc.X).Abs()
            dz := (chunkLoc.Z - newChunkLoc.Z).Abs()
            if dx > ChunkRadius || dz > ChunkRadius {
                chunksToRemove = append(chunksToRemove, chunk)
            }
        }

        cs.curChunk = newChunkLoc
    } else {
        // Player arriving in the world. We send all nearby chunks.
        newPos.UpdateChunkXZ(&cs.curChunk)
        cs.curChunkValid = true
        chunkLocLoadOrder = chunkOrder(ChunkRadius, &cs.curChunk)
    }
    return
}

func chunkOrder(radius ChunkCoord, center *ChunkXZ) (locs []ChunkXZ) {
    areaEdgeSize := radius*2 + 1
    locs = make([]ChunkXZ, areaEdgeSize*areaEdgeSize)
    locs[0] = *center
    index := 1
    for curRadius := ChunkCoord(1); curRadius <= radius; curRadius++ {
        xMin := ChunkCoord(-curRadius + center.X)
        xMax := ChunkCoord(curRadius + center.X)
        zMin := ChunkCoord(-curRadius + center.Z)
        zMax := ChunkCoord(curRadius + center.Z)

        // Northern and southern rows of chunks.
        for x := xMin; x <= xMax; x++ {
            locs[index] = ChunkXZ{x, zMin}
            index++
            locs[index] = ChunkXZ{x, zMax}
            index++
        }

        // Eastern and western columns (except for where they intersect the
        // north and south rows).
        for z := zMin + 1; z < zMax; z++ {
            locs[index] = ChunkXZ{xMin, z}
            index++
            locs[index] = ChunkXZ{xMax, z}
            index++
        }
    }
    return
}

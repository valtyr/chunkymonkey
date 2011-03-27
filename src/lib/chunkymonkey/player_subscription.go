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
    player *Player
    chunks map[IChunk]bool
}

func (cs *chunkSubscriptions) init(player *Player) {
    cs.player = player
    cs.chunks = make(map[IChunk]bool)
}

// Subscribes to all chunks in MinChunkRadius, then calls nearbySent(), then
// subscribes to the remaining chunks out to ChunkRadius. Note that the bulk of
// this work happens in a seperate goroutine.
func (cs *chunkSubscriptions) subscribeFresh(nearbySent func()) {
    playerChunkLoc := cs.player.position.ToChunkXZ()

    // Work out order to send chunks in. (Nearest to player first)
    chunkLocOrder := chunkOrder(ChunkRadius, playerChunkLoc)

    // Get all the chunks together that we're going to send.
    chunksChan := make(chan []IChunk)
    cs.player.game.Enqueue(func(game IGame) {
        chunks := make([]IChunk, 0, len(chunkLocOrder))
        mgr := cs.player.game.GetChunkManager()
        for _, chunkLoc := range chunkLocOrder {
            chunk := mgr.Get(&chunkLoc)
            if chunk != nil {
                chunks = append(chunks, chunk)
            }
        }
        chunksChan <- chunks
    })
    chunks := <-chunksChan

    // Remember the chunks that we're going to be subscribed to.
    for _, chunk := range chunks {
        cs.chunks[chunk] = true
    }

    // Send all chunks in the background.
    go func() {
        // Warn client to allocate memory for the chunks we're sending.
        buf := &bytes.Buffer{}
        for _, chunk := range chunks {
            proto.WritePreChunk(buf, chunk.GetLoc(), ChunkInit)
        }
        cs.player.TransmitPacket(buf.Bytes())

        // Send the important chunks.
        chunksSent := 0
        waitChunks := &sync.WaitGroup{}
        for _, chunk := range chunks {
            chunkLoc := chunk.GetLoc()
            dx := (chunkLoc.X - playerChunkLoc.X).Abs()
            dz := (chunkLoc.Z - playerChunkLoc.Z).Abs()
            if dx > MinChunkRadius || dz > MinChunkRadius {
                // This chunk isn't important, wait until after other login data
                // has been send before sending this and chunks following.
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

        nearbySent()

        // Send the remaining chunks. We don't need to wait for these to complete
        // before continuing with talking to the client.
        for _, chunk := range chunks[chunksSent:] {
            chunk.Enqueue(func(chunk IChunk) {
                chunk.AddSubscriber(cs.player)
            })
        }
    }()
}

// Removes all subscriptions to chunks.
func (cs *chunkSubscriptions) clear() {
    for chunk := range cs.chunks {
        chunk.RemoveSubscriber(cs.player)
    }
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

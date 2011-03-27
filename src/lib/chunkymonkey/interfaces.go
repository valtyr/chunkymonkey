package interfaces

import (
    "io"
    "os"
    "rand"

    "chunkymonkey/entity"
    "chunkymonkey/physics"
    .   "chunkymonkey/types"
)

// Subset of player methods used by chunks.
type IPacketSender interface {
    // Used to send a packet to some remote party. This may be called from any
    // goroutine.
    TransmitPacket(packet []byte)
}

type IPlayer interface {
    IPacketSender

    // Safe to call from outside of player's own goroutine.
    GetEntity() *entity.Entity // Only the game mainloop may modify the return value
    GetName() string           // Do not modify return value
    LockedGetChunkPosition() *ChunkXZ

    Enqueue(f func(IPlayer))

    // Everything below must be called from within Enqueue

    SendSpawn(writer io.Writer) (err os.Error)
    IsWithin(p1, p2 *ChunkXZ) bool
}

type IItem interface {
    // Safe to call from outside of chunk's own goroutine
    GetEntity() *entity.Entity // Only the game mainloop may modify the return value

    // Item methods must be called from the goroutine of their parent chunk.
    // Note that items move between chunks.
    GetPosition() *AbsXYZ
    SendSpawn(writer io.Writer) (err os.Error)
    SendUpdate(writer io.Writer) (err os.Error)
    Tick(blockQuery physics.BlockQueryFn) (leftBlock bool)
}

type IBlockType interface {
    Destroy(chunk IChunk, blockLoc *BlockXYZ) bool
    IsSolid() bool
    GetName() string
    GetTransparency() int8
}

type IChunk interface {
    // Safe to call from outside of Enqueue:
    GetLoc() *ChunkXZ // Do not modify return value

    Enqueue(f func(IChunk))

    // Everything below must be called from within Enqueue

    // Called from game loop to run physics etc. within the chunk for a single
    // tick.
    Tick()

    // Intended for use by blocks/entities within the chunk.
    GetRand() *rand.Rand
    AddItem(item IItem)
    // Tells the chunk to take posession of the item.
    TransferItem(item IItem)
    GetBlock(subLoc *SubChunkXYZ) (blockType BlockID, ok bool)
    DestroyBlock(subLoc *SubChunkXYZ) (ok bool)

    // Register and unregister subscribers to receive information about the
    // chunk. When added, a subscriber will immediately receive complete chunk
    // information via their TransmitPacket method, and changes thereafter via
    // the same mechanism.
    AddSubscriber(subscriber IPacketSender)
    RemoveSubscriber(subscriber IPacketSender)

    // Get packet data for the chunk
    SendUpdate()
}

type IChunkManager interface {
    // Must currently be called from with the owning IGame's Enqueue:
    Get(loc *ChunkXZ) (chunk IChunk)
    ChunksInRadius(loc *ChunkXZ) <-chan IChunk
    ChunksActive() <-chan IChunk
}

type IGame interface {
    // Safe to call from outside of Enqueue:
    GetStartPosition() *AbsXYZ      // Do not modify return value
    GetChunkManager() IChunkManager // Respect calling methods on the return value within Enqueue
    GetBlockTypes() map[BlockID]IBlockType

    Enqueue(f func(IGame))

    // Everything below must be called from within Enqueue

    AddEntity(entity *entity.Entity)
    AddPlayer(player IPlayer)
    RemovePlayer(player IPlayer)
    MulticastPacket(packet []byte, except interface{})
    MulticastChunkPacket(packet []byte, loc *ChunkXZ)
    MulticastRadiusPacket(packet []byte, sender IPlayer)
    SendChatMessage(message string)
}

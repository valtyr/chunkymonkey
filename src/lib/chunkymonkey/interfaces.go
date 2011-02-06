package interfaces

import (
    "io"
    "os"

    "chunkymonkey/entity"
    "chunkymonkey/physics"
    . "chunkymonkey/types"
)

type IPlayer interface {
    // Safe to call from outside of player's own goroutine.
    GetEntity() *entity.Entity // Only the game mainloop may modify the return value
    GetName() string // Do not modify return value

    TransmitPacket(packet []byte)
    Enqueue(f func(IPlayer))

    // Must be called from within Enqueue:
    SendSpawn(writer io.Writer) (err os.Error)
    GetChunkPosition() *ChunkXZ
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

type IChunk interface {
    // Safe to call from outside of Enqueue:
    GetLoc() *ChunkXZ // Do not modify return value

    Enqueue(f func(IChunk))

    // Must be called from within Enqueue:
    AddItem(item IItem)
    TransferItem(item IItem)
    DestroyBlock(subLoc *SubChunkXYZ) (ok bool)
    SendChunkData(writer io.Writer) (err os.Error)
    GetBlock(subLoc *SubChunkXYZ) (blockType BlockID, ok bool)
    SendUpdate()
    PhysicsTick()
}

type IChunkManager interface {
    // Must currently be called from with the owning IGame's Enqueue:
    Get(loc *ChunkXZ) (chunk IChunk)
    ChunksInRadius(loc *ChunkXZ) (c chan IChunk)
}

type IGame interface {
    // Safe to call from outside of Enqueue:
    GetStartPosition() *AbsXYZ // Do not modify return value
    GetChunkManager() IChunkManager // Respect calling methods on the return value within Enqueue

    Enqueue(f func(IGame))

    // Must be called from within Enqueue:
    AddEntity(entity *entity.Entity)
    AddPlayer(player IPlayer)
    RemovePlayer(player IPlayer)
    MulticastPacket(packet []byte, except interface{})
    MulticastChunkPacket(packet []byte, loc *ChunkXZ)
    MulticastRadiusPacket(packet []byte, sender IPlayer)
    SendChatMessage(message string)
}

package store_beta

import (
    "encoding/binary"
    "fmt"
    "os"
    "path"

    "chunkymonkey/chunk/store"
    .   "chunkymonkey/types"
)

const (
    chunkFileEdge       = 32
    chunkFileEdgeShift  = 5
    chunkFileSectorSize = 4096
)

// Reads data from stored Beta worlds.
type chunkStoreBeta struct {
    worldPath string
}

func NewChunkStoreBeta(worldPath string) store.ChunkStore {
    return &chunkStoreBeta{
        worldPath: worldPath,
    }
}

func chunkFilePath(worldPath string, chunkLoc *ChunkXZ) string {
    return path.Join(
        worldPath,
        "region",
        fmt.Sprintf(
            "r.%d.%d.mcr",
            chunkLoc.X>>chunkFileEdgeShift,
            chunkLoc.Z>>chunkFileEdgeShift,
        ),
    )
}

func (s *chunkStoreBeta) LoadChunk(chunkLoc *ChunkXZ) (reader store.ChunkReader, err os.Error) {
    // TODO cache limited number of likely-to-be-used chunkFileReader objs
    filePath := chunkFilePath(s.worldPath, chunkLoc)
    cfr, err := newChunkFileReader(filePath)
    if err != nil {
        return
    }

    return cfr.ReadChunkData(chunkLoc)
}

// A chunk file header entry.
type chunkOffset uint32

// Returns true if the offset value states that the chunk is present in the
// file.
func (o chunkOffset) IsPresent() bool {
    return o != 0
}

func (o chunkOffset) Get() (sectorCount, sectorIndex int) {
    sectorCount = int(o & 0xff)
    sectorIndex = int(o >> 8)
    return
}

type chunkFileHeader [chunkFileEdge * chunkFileEdge]chunkOffset

// Returns the chunk offset data for the given chunk. It assumes that chunkLoc
// is within the chunk file - discarding upper bits of the X and Z coords.
func (h chunkFileHeader) GetOffset(chunkLoc *ChunkXZ) chunkOffset {
    x := chunkLoc.X & (chunkFileEdge - 1)
    z := chunkLoc.Z & (chunkFileEdge - 1)
    return h[x+(z<<chunkFileEdgeShift)]
}

// Handle on a chunk file - used to read chunk data from the file.
type chunkFileReader struct {
    offsets chunkFileHeader
}

func newChunkFileReader(filePath string) (cfr *chunkFileReader, err os.Error) {
    file, err := os.Open(filePath, os.O_RDONLY, 0)
    if err != nil {
        return
    }
    defer file.Close()

    cfr = &chunkFileReader{}

    err = binary.Read(file, binary.BigEndian, &cfr.offsets)
    if err != nil {
        cfr = nil
        return
    }

    return
}

func (cfr *chunkFileReader) ReadChunkData(chunkLoc *ChunkXZ) (r *chunkReader, err os.Error) {
    // TODO
    return
}

// Reads data from stored Beta chunks.
type chunkReader struct {

}

func (r *chunkReader) ChunkLoc() *ChunkXZ {
    // TODO
    return nil
}

func (r *chunkReader) Blocks() []byte {
    // TODO
    return nil
}

func (r *chunkReader) BlockData() []byte {
    // TODO
    return nil
}

func (r *chunkReader) BlockLight() []byte {
    // TODO
    return nil
}

func (r *chunkReader) SkyLight() []byte {
    // TODO
    return nil
}

func (r *chunkReader) HeightMap() []byte {
    // TODO
    return nil
}

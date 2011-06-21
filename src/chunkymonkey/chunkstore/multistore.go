package chunkstore

import (
	. "chunkymonkey/types"
)

// MultiStore provides the ability to load a chunk from one or more potential
// sources of chunk data. The primary purpose of this is to read from a
// persistant store first, then fall back to generating a chunk if the
// persistant store does not have it. MultiStore implements IChunkStore.
type MultiStore struct {
	stores []IChunkStore
}

func NewMultiStore(stores []IChunkStore) *MultiStore {
	return &MultiStore{
		stores: stores,
	}
}

func (s *MultiStore) LoadChunk(chunkLoc *ChunkXz) (result <-chan ChunkResult) {
	resultChan := make(chan ChunkResult)

	// TODO This very rapidly creates large numbers of goroutines when generating
	// chunks and exhausts memory. Should generate from a pool instead.
	go s.loadChunk(chunkLoc, resultChan)

	return resultChan
}

func (s *MultiStore) loadChunk(chunkLoc *ChunkXz, resultChan chan<- ChunkResult) {
	for _, store := range s.stores {
		result := <-store.LoadChunk(chunkLoc)
		if result.Err != nil {
			if _, ok := result.Err.(NoSuchChunkError); ok {
				// Fall through to next chunk store.
				continue
			}
		}

		resultChan<- result
	}

	resultChan<- ChunkResult{
		Reader: nil,
		Err:    NoSuchChunkError(false),
	}
}

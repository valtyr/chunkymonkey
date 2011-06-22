package chunkstore

import (
	"os"

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
	s := &MultiStore{
		stores: stores,
	}

	return s
}

func (s *MultiStore) LoadChunk(chunkLoc ChunkXz) (reader IChunkReader, err os.Error) {
	for _, store := range s.stores {
		result := <-store.LoadChunk(chunkLoc)

		if result.Err == nil {
			return result.Reader, result.Err
		} else {
			if _, ok := result.Err.(NoSuchChunkError); ok {
				// Fall through to next chunk store.
				continue
			}
			return nil, result.Err
		}
	}

	return nil, NoSuchChunkError(false)
}

package chunkstore

import (
	"io"
	"log"
	"os"

	"chunkymonkey/gamerules"
	. "chunkymonkey/types"
	"nbt"
)

// Returned to chunks to pull their data from.
type nbtChunkReader struct {
	chunkTag nbt.ITag
}

// Load a chunk from its NBT representation
func newNbtChunkReader(reader io.Reader) (r *nbtChunkReader, err os.Error) {
	chunkTag, err := nbt.Read(reader)
	if err != nil {
		return
	}

	r = &nbtChunkReader{
		chunkTag: chunkTag,
	}

	return
}

func (r *nbtChunkReader) ChunkLoc() ChunkXz {
	return ChunkXz{
		X: ChunkCoord(r.chunkTag.Lookup("Level/xPos").(*nbt.Int).Value),
		Z: ChunkCoord(r.chunkTag.Lookup("Level/zPos").(*nbt.Int).Value),
	}
}

func (r *nbtChunkReader) Blocks() []byte {
	return r.chunkTag.Lookup("Level/Blocks").(*nbt.ByteArray).Value
}

func (r *nbtChunkReader) BlockData() []byte {
	return r.chunkTag.Lookup("Level/Data").(*nbt.ByteArray).Value
}

func (r *nbtChunkReader) BlockLight() []byte {
	return r.chunkTag.Lookup("Level/BlockLight").(*nbt.ByteArray).Value
}

func (r *nbtChunkReader) SkyLight() []byte {
	return r.chunkTag.Lookup("Level/SkyLight").(*nbt.ByteArray).Value
}

func (r *nbtChunkReader) HeightMap() []byte {
	return r.chunkTag.Lookup("Level/HeightMap").(*nbt.ByteArray).Value
}

func (r *nbtChunkReader) Entities() (entities []gamerules.INonPlayerEntity) {
	entityListTag, ok := r.chunkTag.Lookup("Level/Entities").(*nbt.List)
	if !ok {
		return
	}

	entities = make([]gamerules.INonPlayerEntity, 0, len(entities))

	for _, entityTag := range entityListTag.Value {
		entityObjectId, ok := entityTag.Lookup("id").(*nbt.String)

		if !ok {
			log.Printf("missing or bad entity type ID in NBT: %s", entityObjectId)
		} else {
			if entity := gamerules.NewEntityByTypeName(entityObjectId.Value); entity == nil {
				log.Printf("Found unhandled entity type: %s", entityObjectId.Value)
			} else {
				if err := entity.ReadNbt(entityTag); err != nil {
					log.Printf("Error reading entity NBT: %s", err)
				} else {
					entities = append(entities, entity)
				}
			}
		}
	}

	return
}

func (r *nbtChunkReader) RootTag() nbt.ITag {
	return r.chunkTag
}

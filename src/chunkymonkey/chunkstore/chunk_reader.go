package chunkstore

import (
	"io"
	"log"
	"os"

	"chunkymonkey/gamerules"
	"chunkymonkey/nbtutil"
	. "chunkymonkey/types"
	"nbt"
)

// Returned to chunks to pull their data from.
type chunkReader struct {
	chunkTag nbt.ITag
}

// Load a chunk from its NBT representation
func newChunkReader(reader io.Reader) (r *chunkReader, err os.Error) {
	chunkTag, err := nbt.Read(reader)
	if err != nil {
		return
	}

	r = &chunkReader{
		chunkTag: chunkTag,
	}

	return
}

func (r *chunkReader) ChunkLoc() ChunkXz {
	return ChunkXz{
		X: ChunkCoord(r.chunkTag.Lookup("Level/xPos").(*nbt.Int).Value),
		Z: ChunkCoord(r.chunkTag.Lookup("Level/zPos").(*nbt.Int).Value),
	}
}

func (r *chunkReader) Blocks() []byte {
	return r.chunkTag.Lookup("Level/Blocks").(*nbt.ByteArray).Value
}

func (r *chunkReader) BlockData() []byte {
	return r.chunkTag.Lookup("Level/Data").(*nbt.ByteArray).Value
}

func (r *chunkReader) BlockLight() []byte {
	return r.chunkTag.Lookup("Level/BlockLight").(*nbt.ByteArray).Value
}

func (r *chunkReader) SkyLight() []byte {
	return r.chunkTag.Lookup("Level/SkyLight").(*nbt.ByteArray).Value
}

func (r *chunkReader) HeightMap() []byte {
	return r.chunkTag.Lookup("Level/HeightMap").(*nbt.ByteArray).Value
}

func (r *chunkReader) Entities() (entities []gamerules.INonPlayerEntity) {
	entityListTag, ok := r.chunkTag.Lookup("Level/Entities").(*nbt.List)
	if !ok {
		return
	}

	entities = make([]gamerules.INonPlayerEntity, 0, len(entities))

	for _, entity := range entityListTag.Value {
		// Position within the chunk
		pos, err := nbtutil.ReadAbsXyz(entity, "Pos")
		if err != nil {
			continue
		}

		// Motion
		velocity, err := nbtutil.ReadAbsVelocity(entity, "Motion")
		if err != nil {
			continue
		}

		// Look
		look, err := nbtutil.ReadLookDegrees(entity, "Rotation")
		if err != nil {
			continue
		}

		_ = entity.Lookup("OnGround").(*nbt.Byte).Value
		_ = entity.Lookup("FallDistance").(*nbt.Float).Value
		_ = entity.Lookup("Air").(*nbt.Short).Value
		_ = entity.Lookup("Fire").(*nbt.Short).Value

		var newEntity gamerules.INonPlayerEntity
		entityObjectId := entity.Lookup("id").(*nbt.String).Value

		switch entityObjectId {
		case "Item":
			itemInfo := entity.Lookup("Item").(*nbt.Compound)

			// Grab the basic item data
			id := ItemTypeId(itemInfo.Lookup("id").(*nbt.Short).Value)
			count := ItemCount(itemInfo.Lookup("Count").(*nbt.Byte).Value)
			data := ItemData(itemInfo.Lookup("Damage").(*nbt.Short).Value)
			newEntity = gamerules.NewItem(id, count, data, &pos, &velocity, 0)
		case "Chicken":
			newEntity = gamerules.NewHen(&pos, &velocity, &look)
		case "Cow":
			newEntity = gamerules.NewCow(&pos, &velocity, &look)
		case "Creeper":
			newEntity = gamerules.NewCreeper(&pos, &velocity, &look)
		case "Pig":
			newEntity = gamerules.NewPig(&pos, &velocity, &look)
		case "Sheep":
			newEntity = gamerules.NewSheep(&pos, &velocity, &look)
		case "Skeleton":
			newEntity = gamerules.NewSkeleton(&pos, &velocity, &look)
		case "Squid":
			newEntity = gamerules.NewSquid(&pos, &velocity, &look)
		case "Spider":
			newEntity = gamerules.NewSpider(&pos, &velocity, &look)
		case "Wolf":
			newEntity = gamerules.NewWolf(&pos, &velocity, &look)
		case "Zombie":
			newEntity = gamerules.NewZombie(&pos, &velocity, &look)
		default:
			// Handle all other objects
			objType, ok := ObjTypeMap[entityObjectId]
			if ok {
				newEntity = gamerules.NewObject(objType, &pos, &velocity)
			} else {
				log.Printf("Found unhandled entity type: %s", entityObjectId)
			}
		}

		entities = append(entities, newEntity)
	}

	return
}

func (r *chunkReader) RootTag() nbt.ITag {
	return r.chunkTag
}

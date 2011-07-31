package chunkstore

import (
	"io"
	"log"
	"os"

	"chunkymonkey/gamerules"
	"chunkymonkey/mob"
	"chunkymonkey/nbtutil"
	"chunkymonkey/object"
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

func (r *chunkReader) Entities() (entities []object.INonPlayerEntity) {
	entityListTag, ok := r.chunkTag.Lookup("Level/Entities").(*nbt.List)
	if !ok {
		return
	}

	entities = make([]object.INonPlayerEntity, 0, len(entities))

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

		var newEntity object.INonPlayerEntity
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
			newEntity = mob.NewHen(&pos, &velocity, &look)
		case "Cow":
			newEntity = mob.NewCow(&pos, &velocity, &look)
		case "Creeper":
			newEntity = mob.NewCreeper(&pos, &velocity, &look)
		case "Pig":
			newEntity = mob.NewPig(&pos, &velocity, &look)
		case "Sheep":
			newEntity = mob.NewSheep(&pos, &velocity, &look)
		case "Skeleton":
			newEntity = mob.NewSkeleton(&pos, &velocity, &look)
		case "Squid":
			newEntity = mob.NewSquid(&pos, &velocity, &look)
		case "Spider":
			newEntity = mob.NewSpider(&pos, &velocity, &look)
		case "Wolf":
			newEntity = mob.NewWolf(&pos, &velocity, &look)
		case "Zombie":
			newEntity = mob.NewZombie(&pos, &velocity, &look)
		default:
			// Handle all other objects
			objType, ok := ObjTypeMap[entityObjectId]
			if ok {
				newEntity = object.NewObject(objType, &pos, &velocity)
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

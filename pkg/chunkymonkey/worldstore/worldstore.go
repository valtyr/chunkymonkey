// Responsible for reading the overall world persistent state.
// Eventually this should also be responsible for writing it as well.
package worldstore

import (
	"compress/gzip"
	"fmt"
	"os"
	"path"

	"chunkymonkey/chunkstore"
	"chunkymonkey/generation"
	. "chunkymonkey/types"
	"chunkymonkey/nbt"
)

type WorldStore struct {
	WorldPath string

	LevelData     *nbt.NamedTag
	ChunkStore    chunkstore.IChunkStore
	StartPosition AbsXyz
}

func LoadWorldStore(worldPath string) (world *WorldStore, err os.Error) {
	levelData, err := loadLevelData(worldPath)
	if err != nil {
		return
	}

	startPosition, err := absXyzFromNbt(levelData, "/Data/Player/Pos")
	if err != nil {
		// TODO Hack - remove this when SMP loading is supported properly.
		startPosition = AbsXyz{ChunkSizeH/2, ChunkSizeY, ChunkSizeH/2}
	}

	var chunkStores []chunkstore.IChunkStore
	persistantChunkStore, err := chunkstore.ChunkStoreForLevel(worldPath, levelData)
	if err != nil {
		return
	}
	chunkStores = append(chunkStores, chunkstore.NewChunkService(persistantChunkStore))

	seed, ok := levelData.Lookup("/Data/RandomSeed").(*nbt.Long)
	if ok {
		chunkStores = append(chunkStores, chunkstore.NewChunkService(generation.NewTestGenerator(seed.Value)))
	}

	for _, store := range chunkStores {
		go store.Serve()
	}

	world = &WorldStore{
		WorldPath:     worldPath,
		LevelData:     levelData,
		ChunkStore:    chunkstore.NewChunkService(chunkstore.NewMultiStore(chunkStores)),
		StartPosition: startPosition,
	}

	go world.ChunkStore.Serve()

	return
}

func loadLevelData(worldPath string) (levelData *nbt.NamedTag, err os.Error) {
	file, err := os.Open(path.Join(worldPath, "level.dat"))
	if err != nil {
		return
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return
	}
	defer gzipReader.Close()

	levelData, err = nbt.Read(gzipReader)

	return
}

func absXyzFromNbt(tag nbt.ITag, path string) (pos AbsXyz, err os.Error) {
	posList, posOk := tag.Lookup(path).(*nbt.List)
	if !posOk {
		err = BadType(path)
		return
	}
	x, xOk := posList.Value[0].(*nbt.Double)
	y, yOk := posList.Value[1].(*nbt.Double)
	z, zOk := posList.Value[2].(*nbt.Double)
	if !xOk || !yOk || !zOk {
		err = BadType(path)
		return
	}

	pos = AbsXyz{
		AbsCoord(x.Value),
		AbsCoord(y.Value),
		AbsCoord(z.Value),
	}
	return
}

type BadType string

func (err BadType) String() string {
	return fmt.Sprintf("Bad type in level.dat for %s", string(err))
}

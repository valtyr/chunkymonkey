package gamerules

import (
	"os"

	"chunkymonkey/nbtutil"
	"chunkymonkey/types"
	"nbt"
)

type tileEntity struct {
	chunk    IChunkBlock
	blockLoc types.BlockXyz
}

func (tileEntity *tileEntity) UnmarshalNbt(tag *nbt.Compound) (err os.Error) {
	if tileEntity.blockLoc, err = nbtutil.ReadBlockXyzCompound(tag); err != nil {
		return
	}

	return nil
}

func (tileEntity *tileEntity) MarshalNbt(tag *nbt.Compound) (err os.Error) {
	nbtutil.WriteBlockXyzCompound(tag, tileEntity.blockLoc)
	return nil
}

func (tileEntity *tileEntity) SetChunk(chunk IChunkBlock) {
	tileEntity.chunk = chunk
}

func (tileEntity *tileEntity) Block() types.BlockXyz {
	return tileEntity.blockLoc
}

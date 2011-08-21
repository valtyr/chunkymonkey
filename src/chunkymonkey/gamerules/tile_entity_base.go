package gamerules

import (
	"os"

	"chunkymonkey/nbtutil"
	. "chunkymonkey/types"
	"nbt"
)

type tileEntity struct {
	chunk    IChunkBlock
	blockLoc BlockXyz
}

func (blkInv *tileEntity) UnmarshalNbt(tag nbt.ITag) (err os.Error) {
	if blkInv.blockLoc, err = nbtutil.ReadBlockXyzCompound(tag); err != nil {
		return
	}

	return nil
}

func (blkInv *tileEntity) MarshalNbt(tag *nbt.Compound) (err os.Error) {
	nbtutil.WriteBlockXyzCompound(tag, blkInv.blockLoc)
	return nil
}

func (blkInv *tileEntity) SetChunk(chunk IChunkBlock) {
	blkInv.chunk = chunk
}

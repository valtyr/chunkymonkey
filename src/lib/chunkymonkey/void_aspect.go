package block

import (
    . "chunkymonkey/types"
)

func makeVoidAspect() (aspect IBlockAspect) {
    return &VoidAspect{}
}

// Behaviour of a "void" block which has no behaviour.
type VoidAspect struct {}

func (aspect *VoidAspect) Dig(chunk IChunkBlock, blockLoc *BlockXyz, digStatus DigStatus) (destroyed bool) {
    destroyed = false
    return
}

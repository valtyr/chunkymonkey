package block

import (
	"chunkymonkey/shardserver_external"
	. "chunkymonkey/types"
)

func makeVoidAspect() (aspect IBlockAspect) {
	return &VoidAspect{}
}

// Behaviour of a "void" block which has no behaviour.
type VoidAspect struct{}

func (aspect *VoidAspect) Name() string {
	return "Void"
}

func (aspect *VoidAspect) Hit(instance *BlockInstance, player shardserver_external.IPlayerConnection, digStatus DigStatus) (destroyed bool) {
	destroyed = false
	return
}

func (aspect *VoidAspect) Interact(instance *BlockInstance, player shardserver_external.IPlayerConnection) {
}

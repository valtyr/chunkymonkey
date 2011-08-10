package gamerules

import (
	"os"
)

// Behaviour of a sapling block, takes care of growing or dying depending on
// world conditions.
func makeSaplingAspect() (aspect IBlockAspect) {
	return &SaplingAspect{
		&StandardAspect{},
	}
}

type SaplingAspect struct {
	*StandardAspect
}

func (aspect *SaplingAspect) Name() string {
	return "Sapling"
}

func (aspect *SaplingAspect) Check() os.Error {
	return nil
}

func (aspect *SaplingAspect) Tick(instance *BlockInstance) bool {
	// TODO: Need to find a way to bootstrap a tick
	return true
}

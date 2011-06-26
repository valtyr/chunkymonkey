// Utility functions for reading/writing values to NBT files.
package nbtutil

import (
	. "chunkymonkey/types"
	"nbt"
)

func ReadFloat2(tag *nbt.Compound, path string) (x, y float32, ok bool) {
	list, ok := tag.Lookup(path).(*nbt.List)
	if !ok || len(list.Value) != 2 {
		return
	}

	xv, xok := list.Value[0].(*nbt.Float)
	yv, yok := list.Value[1].(*nbt.Float)

	if ok = xok && yok; ok {
		x, y = xv.Value, yv.Value
	}

	return
}

func ReadDouble3(tag *nbt.Compound, path string) (x, y, z float64, ok bool) {
	list, ok := tag.Lookup(path).(*nbt.List)
	if !ok || len(list.Value) != 3 {
		return
	}

	xv, xok := list.Value[0].(*nbt.Double)
	yv, yok := list.Value[1].(*nbt.Double)
	zv, zok := list.Value[2].(*nbt.Double)

	if ok = xok && yok && zok; ok {
		x, y, z = xv.Value, yv.Value, zv.Value
	}

	return
}

func ReadAbsXyz(tag *nbt.Compound, path string) (pos *AbsXyz, ok bool) {
	x, y, z, ok := ReadDouble3(tag, path)
	if !ok {
		return
	}

	return &AbsXyz{AbsCoord(x), AbsCoord(y), AbsCoord(z)}, true
}

func ReadAbsVelocity(tag *nbt.Compound, path string) (pos *AbsVelocity, ok bool) {
	x, y, z, ok := ReadDouble3(tag, path)
	if !ok {
		return
	}

	return &AbsVelocity{AbsVelocityCoord(x), AbsVelocityCoord(y), AbsVelocityCoord(z)}, true
}

func ReadLookDegrees(tag *nbt.Compound, path string) (pos *LookDegrees, ok bool) {
	x, y, ok := ReadFloat2(tag, path)
	if !ok {
		return
	}

	return &LookDegrees{AngleDegrees(x), AngleDegrees(y)}, true
}

package chunkstore

import (
    "testing"

    .   "chunkymonkey/types"
)

func TestChunkFilePath(t *testing.T) {
    type Test struct {
        loc      ChunkXZ
        expected string
    }

    tests := []Test{
        {ChunkXZ{0, 0}, "/foo/region/r.0.0.mcr"},
        {ChunkXZ{31, 0}, "/foo/region/r.0.0.mcr"},
        {ChunkXZ{0, 31}, "/foo/region/r.0.0.mcr"},
        {ChunkXZ{32, 0}, "/foo/region/r.1.0.mcr"},
        {ChunkXZ{0, 32}, "/foo/region/r.0.1.mcr"},
        {ChunkXZ{-1, 0}, "/foo/region/r.-1.0.mcr"},
        {ChunkXZ{0, -1}, "/foo/region/r.0.-1.mcr"},
        {ChunkXZ{-32, 0}, "/foo/region/r.-1.0.mcr"},
        {ChunkXZ{0, -32}, "/foo/region/r.0.-1.mcr"},
        {ChunkXZ{-33, 0}, "/foo/region/r.-2.0.mcr"},
        {ChunkXZ{0, -33}, "/foo/region/r.0.-2.mcr"},
    }

    for _, test := range tests {
        result := regionFilePath("/foo", &test.loc)
        if test.expected != result {
            t.Errorf(
                "regionFilePath(\"/foo\", %+v) expected %#v but got %#v",
                test.loc, test.expected, result)
        }
    }
}

func TestChunkOffset(t *testing.T) {
    if chunkOffset(0).IsPresent() {
        t.Errorf("chunkOffset(0).IsPresent() should return false")
    }

    type Test struct {
        offset         chunkOffset
        expSectorIndex uint32
        expSectorCount uint32
    }

    tests := []Test{
        {0x00000101, 1, 1},
        {0x00000aff, 10, 255},
    }

    for _, test := range tests {
        sectorCount, sectorIndex := test.offset.Get()
        if test.expSectorCount != sectorCount || test.expSectorIndex != sectorIndex {
            t.Errorf(
                "chunkOffset(0x%x) expected (%d, %d) but got (%d, %d)",
                test.offset,
                test.expSectorCount, test.expSectorIndex,
                sectorCount, sectorIndex)
        }
    }
}

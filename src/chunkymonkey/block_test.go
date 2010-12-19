package chunkymonkey

import (
    "testing"
)

func TestLoadStandardBlocks(t *testing.T) {
    b := make(map[BlockID]*Block)
    LoadStandardBlocks(b)

    type BlockTransTest struct {
        id                    BlockID
        expected_transparency int8
    }

    var BlockTransTests = []BlockTransTest{
        // A few blocks should be transparent
        BlockTransTest{BlockIDAir, 0},
        BlockTransTest{BlockIDSignPost, 0},
        BlockTransTest{BlockIDGlass, 0},

        // Some should be semi-transparent
        BlockTransTest{BlockIDLeaves, 1},
        BlockTransTest{BlockIDWater, 3},
        BlockTransTest{BlockIDIce, 3},

        // Some should be opaque
        BlockTransTest{BlockIDStone, -1},
        BlockTransTest{BlockIDDirt, -1},
    }

    for _, r := range BlockTransTests {
        block := b[r.id]
        if r.expected_transparency != block.transparency {
            t.Errorf("block #%d (%s) expected transparency %d, got %d",
                block.id, block.name, r.expected_transparency, block.transparency)
        }
    }
}

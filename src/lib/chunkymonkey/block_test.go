package block

import (
    "testing"
    .   "chunkymonkey/types"
)

func TestLoadStandardBlocksOpacity(t *testing.T) {
    b := LoadStandardBlockTypes()

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
        if r.expected_transparency != block.GetTransparency() {
            t.Errorf("block #%d (%s) expected transparency %d, got %d",
                r.id, block.GetName(), r.expected_transparency, block.GetTransparency())
        }
    }
}

func TestLoadStandardBlocksSolidity(t *testing.T) {
    b := LoadStandardBlockTypes()

    type Test struct {
        id             BlockID
        expected_solid bool
    }

    var tests = []Test{
        // Most blocks should be solid
        Test{BlockIDStone, true},
        Test{BlockIDDirt, true},
        Test{BlockIDFence, true},
        Test{BlockIDWorkbench, true},

        // Some should be non-solid
        Test{BlockIDWater, false},
        Test{BlockIDLava, false},
        Test{BlockIDYellowFlower, false},
    }

    for _, r := range tests {
        block := b[r.id]
        if r.expected_solid != block.IsSolid() {
            t.Errorf("block #%d (%s) expected IsSolid %t, got %t",
                r.id, block.GetName(), r.expected_solid, block.IsSolid())
        }
    }
}

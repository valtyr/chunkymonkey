package block

import (
    "testing"
    . "chunkymonkey/types"
)

func TestLoadStandardBlocksOpacity(t *testing.T) {
    b := LoadStandardBlockTypes()

    type BlockTransTest struct {
        id                    BlockId
        expected_transparency int8
    }

    var BlockTransTests = []BlockTransTest{
        // A few blocks should be transparent
        BlockTransTest{BlockIdAir, 0},
        BlockTransTest{BlockIdSignPost, 0},
        BlockTransTest{BlockIdGlass, 0},

        // Some should be semi-transparent
        BlockTransTest{BlockIdLeaves, 1},
        BlockTransTest{BlockIdWater, 3},
        BlockTransTest{BlockIdIce, 3},

        // Some should be opaque
        BlockTransTest{BlockIdStone, -1},
        BlockTransTest{BlockIdDirt, -1},
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
        id             BlockId
        expected_solid bool
    }

    var tests = []Test{
        // Most blocks should be solid
        Test{BlockIdStone, true},
        Test{BlockIdDirt, true},
        Test{BlockIdFence, true},
        Test{BlockIdWorkbench, true},

        // Some should be non-solid
        Test{BlockIdWater, false},
        Test{BlockIdLava, false},
        Test{BlockIdYellowFlower, false},
    }

    for _, r := range tests {
        block := b[r.id]
        if r.expected_solid != block.IsSolid() {
            t.Errorf("block #%d (%s) expected IsSolid %t, got %t",
                r.id, block.GetName(), r.expected_solid, block.IsSolid())
        }
    }
}

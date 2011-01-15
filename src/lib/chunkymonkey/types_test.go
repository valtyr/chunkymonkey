package types

import (
    "testing"
)

func TestLookDegrees_ToLookBytes(t *testing.T) {
    type Test struct {
        input    LookDegrees
        expected LookBytes
    }

    var tests = []Test{
        {LookDegrees{0, 0}, LookBytes{0, 0}},
        {LookDegrees{0, 90}, LookBytes{0, 64}},
        {LookDegrees{0, 180}, LookBytes{0, 128}},
        {LookDegrees{0, -90}, LookBytes{0, 192}},
        {LookDegrees{0, 270}, LookBytes{0, 192}},
        {LookDegrees{90, 0}, LookBytes{64, 0}},
        {LookDegrees{180, 0}, LookBytes{128, 0}},
        {LookDegrees{-90, 0}, LookBytes{192, 0}},
        {LookDegrees{270, 0}, LookBytes{192, 0}},
    }

    for _, r := range tests {
        result := r.input.ToLookBytes()
        if r.expected.Yaw != result.Yaw || r.expected.Pitch != result.Pitch {
            t.Errorf("LookDegrees%v expected LookBytes%v got LookBytes%v",
                r.input, r.expected, result)
        }
    }
}

func TestCoordDivMod(t *testing.T) {
    type CoordDivModTest struct {
        expected_div, expected_mod int32
        num, denom                 int32
    }

    var CoordDivModTests = []CoordDivModTest{
        // Simple +ve numerator cases
        CoordDivModTest{0, 0, 0, 16},
        CoordDivModTest{0, 1, 1, 16},
        CoordDivModTest{0, 15, 15, 16},
        CoordDivModTest{1, 0, 16, 16},
        CoordDivModTest{1, 15, 31, 16},

        // -ve numerator cases
        CoordDivModTest{-1, 15, -1, 16},
        CoordDivModTest{-1, 0, -16, 16},
        CoordDivModTest{-2, 15, -17, 16},
        CoordDivModTest{-2, 0, -32, 16},
    }

    for _, r := range CoordDivModTests {
        div, mod := coordDivMod(r.num, r.denom)
        if r.expected_div != div || r.expected_mod != mod {
            t.Errorf("coordDivMod(%d, %d) expected (%d, %d) got (%d, %d)",
                r.num, r.denom, r.expected_div, r.expected_mod, div, mod)
        }
    }
}

func TestChunkXZ_GetChunkCornerBlockXY(t *testing.T) {
    type Test struct {
        input    ChunkXZ
        expected BlockXYZ
    }

    var tests = []Test{
        {ChunkXZ{0, 0}, BlockXYZ{0, 0, 0}},
        {ChunkXZ{0, 1}, BlockXYZ{0, 0, 16}},
        {ChunkXZ{1, 0}, BlockXYZ{16, 0, 0}},
        {ChunkXZ{0, -1}, BlockXYZ{0, 0, -16}},
        {ChunkXZ{-1, 0}, BlockXYZ{-16, 0, 0}},
    }

    for _, r := range tests {
        result := r.input.GetChunkCornerBlockXY()
        if r.expected.X != result.X || r.expected.Y != result.Y || r.expected.Z != result.Z {
            t.Errorf("ChunkXZ%v expected BlockXYZ%v got BlockXYZ%v",
                r.input, r.expected, result)
        }
    }
}

func TestBlockXYZ_ToAbsIntXYZ(t *testing.T) {
    type Test struct {
        input    BlockXYZ
        expected AbsIntXYZ
    }

    var tests = []Test{
        {BlockXYZ{0, 0, 0}, AbsIntXYZ{0, 0, 0}},
        {BlockXYZ{0, 0, 1}, AbsIntXYZ{0, 0, 32}},
        {BlockXYZ{0, 0, -1}, AbsIntXYZ{0, 0, -32}},
        {BlockXYZ{1, 0, 0}, AbsIntXYZ{32, 0, 0}},
        {BlockXYZ{-1, 0, 0}, AbsIntXYZ{-32, 0, 0}},
    }

    for _, r := range tests {
        result := r.input.ToAbsIntXYZ()
        if r.expected.X != result.X || r.expected.Y != result.Y || r.expected.Z != result.Z {
            t.Errorf("BlockXYZ%v expected AbsIntXYZ%v got AbsIntXYZ%v",
                r.input, r.expected, result)
        }
    }
}

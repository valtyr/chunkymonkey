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

func TestAbsXyz_UpdateChunkXz(t *testing.T) {
	type Test struct {
		input    AbsXyz
		expected ChunkXz
	}
	var tests = []Test{
		{AbsXyz{0, 0, 0}, ChunkXz{0, 0}},
		{AbsXyz{0, 0, 16}, ChunkXz{0, 1}},
		{AbsXyz{16, 0, 0}, ChunkXz{1, 0}},
		{AbsXyz{0, 0, -16}, ChunkXz{0, -1}},
		{AbsXyz{-16, 0, 0}, ChunkXz{-1, 0}},
		{AbsXyz{-1, 0, -1}, ChunkXz{-1, -1}},
	}

	for _, test := range tests {
		input, expected := test.input, test.expected
		var result ChunkXz
		input.UpdateChunkXz(&result)
		if expected.X != result.X || expected.Z != result.Z {
			t.Errorf("AbsXyz%+v.UpdateChunkXz() expected ChunkXz%+v got ChunkXz%+v",
				input, expected, result)
		}
	}
}

func TestAbsXyz_ToBlockXyz(t *testing.T) {
	type Test struct {
		pos AbsXyz
		exp BlockXyz
	}

	var tests = []Test{
		// Simple positive tests
		{AbsXyz{0.0, 0.0, 0.0}, BlockXyz{0, 0, 0}},
		{AbsXyz{0.1, 0.2, 0.3}, BlockXyz{0, 0, 0}},
		{AbsXyz{1.0, 2.0, 3.0}, BlockXyz{1, 2, 3}},

		// Negative tests
		{AbsXyz{-0.1, -0.2, -0.3}, BlockXyz{-1, -1, -1}},
		{AbsXyz{-1.0, -2.0, -3.0}, BlockXyz{-1, -2, -3}},
		{AbsXyz{-1.5, -2.5, -3.5}, BlockXyz{-2, -3, -4}},
	}

	for _, r := range tests {
		result := r.pos.ToBlockXyz()
		if r.exp.X != result.X || r.exp.Y != result.Y || r.exp.Z != result.Z {
			t.Errorf("AbsXyz%v.ToBlockXyz() expected BlockXyz%v got BlockXyz%v",
				r.pos, r.exp, result)
		}
	}
}

func TestAbsIntXyz_ToChunkXz(t *testing.T) {
	type Test struct {
		input    AbsIntXyz
		expected ChunkXz
	}

	var tests = []Test{
		{AbsIntXyz{0, 0, 0}, ChunkXz{0, 0}},
		{AbsIntXyz{8 * 32, 0, 8 * 32}, ChunkXz{0, 0}},
		{AbsIntXyz{15 * 32, 0, 15 * 32}, ChunkXz{0, 0}},
		{AbsIntXyz{16 * 32, 0, 16 * 32}, ChunkXz{1, 1}},
		{AbsIntXyz{31*32 + 31, 0, 31*32 + 31}, ChunkXz{1, 1}},
		{AbsIntXyz{32 * 32, 0, 32 * 32}, ChunkXz{2, 2}},
		{AbsIntXyz{0, 0, 32 * 32}, ChunkXz{0, 2}},
		{AbsIntXyz{0, 0, -16 * 32}, ChunkXz{0, -1}},
		{AbsIntXyz{0, 0, -1 * 32}, ChunkXz{0, -1}},
		{AbsIntXyz{0, 0, -1}, ChunkXz{0, -1}},
	}

	for _, r := range tests {
		result := r.input.ToChunkXz()
		if r.expected.X != result.X || r.expected.Z != result.Z {
			t.Errorf("AbsIntXyz%v expected ChunkXz%v got ChunkXz%v",
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

func TestChunkXz_GetChunkCornerBlockXY(t *testing.T) {
	type Test struct {
		input    ChunkXz
		expected BlockXyz
	}

	var tests = []Test{
		{ChunkXz{0, 0}, BlockXyz{0, 0, 0}},
		{ChunkXz{0, 1}, BlockXyz{0, 0, 16}},
		{ChunkXz{1, 0}, BlockXyz{16, 0, 0}},
		{ChunkXz{0, -1}, BlockXyz{0, 0, -16}},
		{ChunkXz{-1, 0}, BlockXyz{-16, 0, 0}},
	}

	for _, r := range tests {
		result := r.input.GetChunkCornerBlockXY()
		if r.expected.X != result.X || r.expected.Y != result.Y || r.expected.Z != result.Z {
			t.Errorf("ChunkXz%v expected BlockXyz%v got BlockXyz%v",
				r.input, r.expected, result)
		}
	}
}

func TestChunkXz_ChunkKey(t *testing.T) {
	type Test struct {
		input    ChunkXz
		expected uint64
	}

	var tests = []Test{
		{ChunkXz{0, 0}, 0},
		{ChunkXz{0, 1}, 0x0000000000000001},
		{ChunkXz{1, 0}, 0x0000000100000000},
		{ChunkXz{0, -1}, 0x00000000ffffffff},
		{ChunkXz{-1, 0}, 0xffffffff00000000},
		{ChunkXz{0, 10}, 0x000000000000000a},
		{ChunkXz{10, 0}, 0x0000000a00000000},
		{ChunkXz{10, 11}, 0x0000000a0000000b},
	}

	for _, r := range tests {
		result := r.input.ChunkKey()
		if r.expected != result {
			t.Errorf("ChunkXz%+v.ChunkKey() expected %d got %d",
				r.input, r.expected, result)
		}
	}
}

func TestBlockXyz_ToAbsIntXyz(t *testing.T) {
	type Test struct {
		input    BlockXyz
		expected AbsIntXyz
	}

	var tests = []Test{
		{BlockXyz{0, 0, 0}, AbsIntXyz{0, 0, 0}},
		{BlockXyz{0, 0, 1}, AbsIntXyz{0, 0, 32}},
		{BlockXyz{0, 0, -1}, AbsIntXyz{0, 0, -32}},
		{BlockXyz{1, 0, 0}, AbsIntXyz{32, 0, 0}},
		{BlockXyz{-1, 0, 0}, AbsIntXyz{-32, 0, 0}},
		{BlockXyz{0, 1, 0}, AbsIntXyz{0, 32, 0}},
		{BlockXyz{0, 10, 0}, AbsIntXyz{0, 320, 0}},
		{BlockXyz{0, 63, 0}, AbsIntXyz{0, 2016, 0}},
		{BlockXyz{0, 64, 0}, AbsIntXyz{0, 2048, 0}},
	}

	for _, r := range tests {
		result := r.input.ToAbsIntXyz()
		if r.expected.X != result.X || r.expected.Y != result.Y || r.expected.Z != result.Z {
			t.Errorf("BlockXyz%v expected AbsIntXyz%v got AbsIntXyz%v",
				r.input, r.expected, result)
		}
	}
}

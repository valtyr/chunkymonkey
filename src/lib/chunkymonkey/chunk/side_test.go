package chunk

import (
    "testing"

    . "chunkymonkey/types"
)

func TestblockIndex(t *testing.T) {
    type Test struct {
        input     SubChunkXyz
        exp_index int32
        exp_shift byte
        exp_ok    bool
    }

    var tests = []Test{
        Test{SubChunkXyz{0, 0, 0}, 0, 0, true},
        Test{SubChunkXyz{0, 1, 0}, 1, 4, true},
        Test{SubChunkXyz{0, 2, 0}, 2, 0, true},
        Test{SubChunkXyz{0, 3, 0}, 3, 4, true},

        Test{SubChunkXyz{0, 127, 0}, 127, 4, true},
        Test{SubChunkXyz{0, 0, 1}, 128, 0, true},

        Test{SubChunkXyz{0, 127, 1}, 255, 4, true},
        Test{SubChunkXyz{0, 0, 2}, 256, 0, true},

        // Invalid locations
        Test{SubChunkXyz{16, 0, 0}, 0, 0, false},
        Test{SubChunkXyz{0, 128, 0}, 0, 0, false},
        Test{SubChunkXyz{0, 0, 16}, 0, 0, false},
    }

    for _, r := range tests {
        index, shift, ok := blockIndex(&r.input)
        if r.exp_index != index || r.exp_shift != shift || r.exp_ok != ok {
            t.Errorf("blockIndex(SubChunkXyz%v) expected (%d, %d, %v) got (%d, %d, %v)",
                r.input,
                r.exp_index, r.exp_shift, r.exp_ok,
                index, shift, ok)
        }
    }
}

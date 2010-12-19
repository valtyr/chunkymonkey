package chunkymonkey

import (
    "testing"
)

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

func TestCoordDivMod(t *testing.T) {
    for _, r := range CoordDivModTests {
        div, mod := coordDivMod(r.num, r.denom)
        if r.expected_div != div || r.expected_mod != mod {
            t.Errorf("coordDivMod(%d, %d) expected (%d, %d) got (%d, %d)",
                r.num, r.denom, r.expected_div, r.expected_mod, div, mod)
        }
    }
}

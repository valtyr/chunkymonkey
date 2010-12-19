package chunkymonkey

import (
    "testing"
)

func testCoordDivModCase(t *testing.T, expected_div, expected_mod, num, denom int32) {
    div, mod := coordDivMod(num, denom)
    if expected_div != div || expected_mod != mod {
        t.Errorf("coordDivMod(%d, %d) expected (%d, %d) got (%d, %d)",
            0, 16, expected_div, expected_mod, div, mod)
    }
}

func TestCoordDivMod(t *testing.T) {
    // Simple +ve numerator cases
    testCoordDivModCase(t, 0, 0, 0, 16)
    testCoordDivModCase(t, 0, 1, 1, 16)
    testCoordDivModCase(t, 0, 15, 15, 16)
    testCoordDivModCase(t, 1, 0, 16, 16)
    testCoordDivModCase(t, 1, 15, 31, 16)

    // -ve numerator cases
    testCoordDivModCase(t, -1, 15, -1, 16)
    testCoordDivModCase(t, -1, 0, -16, 16)
    testCoordDivModCase(t, -2, 15, -17, 16)
    testCoordDivModCase(t, -2, 0, -32, 16)
}

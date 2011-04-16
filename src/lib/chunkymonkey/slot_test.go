package slot

import (
    "testing"

    .   "chunkymonkey/types"
)

func slotEq(s1, s2 *Slot) bool {
    return (s1.ItemTypeId == s2.ItemTypeId &&
        s1.Count == s2.Count &&
        s1.Data == s2.Data)
}

func TestSlot_Add(t *testing.T) {
    type Test struct {
        desc                 string
        initialA, initialB   Slot
        expectedA, expectedB Slot
        expectedChange       bool
    }

    tests := []Test{
        {
            "one empty slot added to another",
            Slot{ItemTypeIdNull, 0, 0}, Slot{ItemTypeIdNull, 0, 0},
            Slot{ItemTypeIdNull, 0, 0}, Slot{ItemTypeIdNull, 0, 0},
            false,
        },
        // Tests involving the same item types: (or empty plus an item)
        {
            "1 + 0 => 1 + 0",
            Slot{1, 1, 0}, Slot{ItemTypeIdNull, 0, 0},
            Slot{1, 1, 0}, Slot{ItemTypeIdNull, 0, 0},
            false,
        },
        {
            "1 + 1 => 2 + 0",
            Slot{1, 1, 0}, Slot{1, 1, 0},
            Slot{1, 2, 0}, Slot{ItemTypeIdNull, 0, 0},
            true,
        },
        {
            "0 + 20 => 20 + 0",
            Slot{ItemTypeIdNull, 0, 0}, Slot{1, 20, 0},
            Slot{1, 20, 0}, Slot{ItemTypeIdNull, 0, 0},
            true,
        },
        {
            "0 + 64 => 64 + 0",
            Slot{ItemTypeIdNull, 0, 0}, Slot{1, 64, 0},
            Slot{1, 64, 0}, Slot{ItemTypeIdNull, 0, 0},
            true,
        },
        {
            "32 + 33 => 64 + 1 (hitting max count)",
            Slot{1, 32, 0}, Slot{1, 33, 0},
            Slot{1, 64, 0}, Slot{1, 1, 0},
            true,
        },
        {
            "65 + 1 => 65 + 1 (already above max count)",
            Slot{1, 65, 0}, Slot{1, 1, 0},
            Slot{1, 65, 0}, Slot{1, 1, 0},
            false,
        },
        {
            "64 + 64 => 64 + 64",
            Slot{1, 64, 0}, Slot{1, 64, 0},
            Slot{1, 64, 0}, Slot{1, 64, 0},
            false,
        },
        {
            "1 + 1 => 1 + 1 where items' \"Data\" value differs",
            Slot{1, 1, 5}, Slot{1, 1, 6},
            Slot{1, 1, 5}, Slot{1, 1, 6},
            false,
        },
        {
            "1 + 1 => 2 + 0 where items' \"Data\" value is the same",
            Slot{1, 1, 5}, Slot{1, 1, 5},
            Slot{1, 2, 5}, Slot{ItemTypeIdNull, 0, 0},
            true,
        },
        {
            "0 + 1 => 1 + 0 - carrying the \"use\" value",
            Slot{ItemTypeIdNull, 0, 0}, Slot{1, 1, 5},
            Slot{1, 1, 5}, Slot{ItemTypeIdNull, 0, 0},
            true,
        },
        // Tests involving different item types:
        {
            "different item types don't mingle",
            Slot{1, 5, 0}, Slot{2, 5, 0},
            Slot{1, 5, 0}, Slot{2, 5, 0},
            false,
        },
    }

    var a, b Slot
    for _, test := range tests {
        t.Logf(
            "Test %s: initial a=%+v, b=%+v - expecting a=%+v, b=%+v, changed=%t",
            test.desc,
            test.initialA, test.initialB,
            test.expectedA, test.expectedB,
            test.expectedChange)
        // Sanity check the test itself. Sum of inputs must equal sum of
        // outputs.
        sumInput := test.initialA.Count + test.initialB.Count
        sumExpectedOutput := test.expectedA.Count + test.expectedB.Count
        if sumInput != sumExpectedOutput {
            t.Errorf(
                "    Test incorrect: sum of inputs %d != sum of expected outputs %d",
                sumInput, sumExpectedOutput)
            continue
        }

        a = test.initialA
        b = test.initialB
        changed := a.Add(&b)
        if !slotEq(&test.expectedA, &a) || !slotEq(&test.expectedB, &b) {
            t.Errorf("    Fail: got a=%+v, b=%+v", a, b)
        }
        if test.expectedChange != changed {
            t.Errorf("    Fail: got changed=%t", changed)
        }
    }
}

func TestSlot_Split(t *testing.T) {
    // TODO
}

func TestSlot_Take(t *testing.T) {
    // TODO
}

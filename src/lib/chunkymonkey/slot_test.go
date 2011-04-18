package slot

import (
    "testing"

    . "chunkymonkey/types"
)

func slotEq(s1, s2 *Slot) bool {
    return (s1.ItemTypeId == s2.ItemTypeId &&
        s1.Count == s2.Count &&
        s1.Data == s2.Data)
}

type slotTest struct {
    desc                 string
    initialA, initialB   Slot
    expectedA, expectedB Slot
    expectedChange       bool
}

func (test *slotTest) run(t *testing.T, op func(a, b *Slot) bool) {
    t.Logf("Test %s:", test.desc)
    t.Logf("    initial   a=%+v, b=%+v", test.initialA, test.initialB)
    t.Logf("    expecting a=%+v, b=%+v, changed=%t",
        test.expectedA, test.expectedB,
        test.expectedChange)

    // Sanity check the test itself. Sum of inputs must equal sum of
    // outputs.
    sumInput := test.initialA.Count + test.initialB.Count
    sumExpectedOutput := test.expectedA.Count + test.expectedB.Count
    if sumInput != sumExpectedOutput {
        t.Errorf(
            "    Test incorrect: sum of inputs %d != sum of expected outputs %d",
            sumInput, sumExpectedOutput,
        )
        return
    }

    a := test.initialA
    b := test.initialB
    changed := op(&a, &b)
    if !slotEq(&test.expectedA, &a) || !slotEq(&test.expectedB, &b) {
        t.Errorf("    Fail: got a=%+v, b=%+v", a, b)
    }
    if test.expectedChange != changed {
        t.Errorf("    Fail: got changed=%t", changed)
    }
}

func runTests(t *testing.T, tests []slotTest, op func(a, b *Slot) bool) {
    for i := range tests {
        tests[i].run(t, op)
    }
}

func TestSlot_Add(t *testing.T) {

    tests := []slotTest{
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
            "0 + 1 => 1 + 0 - carrying the \"Data\" value",
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

    runTests(
        t, tests,
        func(a, b *Slot) bool {
            return a.Add(b)
        },
    )
}

func TestSlot_Swap(t *testing.T) {
    tests := []slotTest{
        {
            "swapping unequal slots",
            Slot{1, 2, 3}, Slot{4, 5, 6},
            Slot{4, 5, 6}, Slot{1, 2, 3},
            true,
        },
        {
            "swapping equal slots",
            Slot{1, 2, 3}, Slot{1, 2, 3},
            Slot{1, 2, 3}, Slot{1, 2, 3},
            false,
        },
    }

    runTests(
        t, tests,
        func(a, b *Slot) bool {
            return a.Swap(b)
        },
    )
}

func TestSlot_Split(t *testing.T) {
    tests := []slotTest{
        // No-op tests.
        {
            "splitting an empty slot",
            Slot{ItemTypeIdNull, 0, 0}, Slot{ItemTypeIdNull, 0, 0},
            Slot{ItemTypeIdNull, 0, 0}, Slot{ItemTypeIdNull, 0, 0},
            false,
        },
        {
            "splitting from a non-empty slot to another non-empty",
            Slot{1, 2, 0}, Slot{1, 3, 0},
            Slot{1, 2, 0}, Slot{1, 3, 0},
            false,
        },
        {
            "splitting from an empty slot to a non-empty",
            Slot{ItemTypeIdNull, 0, 0}, Slot{1, 3, 0},
            Slot{ItemTypeIdNull, 0, 0}, Slot{1, 3, 0},
            false,
        },
        {
            "splitting from a non-empty slot to another non-empty with" +
                " incompatible types",
            Slot{1, 2, 0}, Slot{2, 3, 0},
            Slot{1, 2, 0}, Slot{2, 3, 0},
            false,
        },
        // Remaining tests should all result in changes. They all take from a
        // non-empty subject slot and into the src empty slot.
        {
            "splitting even-numbered stack",
            Slot{1, 64, 0}, Slot{ItemTypeIdNull, 0, 0},
            Slot{1, 32, 0}, Slot{1, 32, 0},
            true,
        },
        {
            "splitting odd-numbered stack",
            Slot{1, 5, 0}, Slot{ItemTypeIdNull, 0, 0},
            Slot{1, 2, 0}, Slot{1, 3, 0},
            true,
        },
        {
            "splitting single-item stack",
            Slot{1, 1, 0}, Slot{ItemTypeIdNull, 0, 0},
            Slot{ItemTypeIdNull, 0, 0}, Slot{1, 1, 0},
            true,
        },
        {
            "item type and data copy",
            Slot{1, 64, 5}, Slot{ItemTypeIdNull, 0, 0},
            Slot{1, 32, 5}, Slot{1, 32, 5},
            true,
        },
        {
            "item type and data move",
            Slot{1, 1, 5}, Slot{ItemTypeIdNull, 0, 0},
            Slot{ItemTypeIdNull, 0, 0}, Slot{1, 1, 5},
            true,
        },
    }

    runTests(
        t, tests,
        func(a, b *Slot) bool {
            return a.Split(b)
        },
    )
}

func TestSlot_AddOne(t *testing.T) {
    tests := []slotTest{
        // No-op tests.
        {
            "adding from empty to empty",
            Slot{ItemTypeIdNull, 0, 0}, Slot{ItemTypeIdNull, 0, 0},
            Slot{ItemTypeIdNull, 0, 0}, Slot{ItemTypeIdNull, 0, 0},
            false,
        },
        {
            "adding from empty to non-empty",
            Slot{1, 1, 0}, Slot{ItemTypeIdNull, 0, 0},
            Slot{1, 1, 0}, Slot{ItemTypeIdNull, 0, 0},
            false,
        },
        {
            "adding incompatible types",
            Slot{1, 1, 0}, Slot{2, 4, 0},
            Slot{1, 1, 0}, Slot{2, 4, 0},
            false,
        },
        {
            "adding incompatible data values",
            Slot{1, 1, 0}, Slot{1, 1, 2},
            Slot{1, 1, 0}, Slot{1, 1, 2},
            false,
        },
        {
            "adding to an already full stack",
            Slot{1, SlotCountMax, 0}, Slot{1, 10, 0},
            Slot{1, SlotCountMax, 0}, Slot{1, 10, 0},
            false,
        },
        // Remaining tests should all result in changes. They all take from a
        // non-empty src slot into a compatible subject slot.
        {
            "adding item to empty, copies type and data",
            Slot{ItemTypeIdNull, 0, 0}, Slot{1, 3, 0},
            Slot{1, 1, 0}, Slot{1, 2, 0},
            true,
        },
        {
            "adding item to non-empty",
            Slot{1, 5, 0}, Slot{1, 3, 0},
            Slot{1, 6, 0}, Slot{1, 2, 0},
            true,
        },
        {
            "adding item to non-empty, empties src",
            Slot{1, 5, 2}, Slot{1, 1, 2},
            Slot{1, 6, 2}, Slot{ItemTypeIdNull, 0, 0},
            true,
        },
    }

    runTests(
        t, tests,
        func(a, b *Slot) bool {
            return a.AddOne(b)
        },
    )
}

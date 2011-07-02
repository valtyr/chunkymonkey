package recipe

import (
	"reflect"
	"strings"
	"testing"

	"chunkymonkey/itemtype"
	"chunkymonkey/slot"
	. "chunkymonkey/types"
)

// Borrows some test code from loader_test.go

// Helper for defining multiple input slots with less syntactic boilerplate.
func Slots(slots ...slot.Slot) []slot.Slot {
	return slots
}

// Helper for defining slots with less syntactic boilerplate.
func Slot(itemTypes itemtype.ItemTypeMap, itemTypeId ItemTypeId, count ItemCount, data ItemData) (out *slot.Slot) {
	out = &slot.Slot{
		ItemType: nil,
		Count:    count,
		Data:     data,
	}
	if itemTypeId != ItemTypeIdNull {
		out.ItemType = itemTypes[itemTypeId]
	}
	return
}

func assertSlotEq(t *testing.T, expected, result *slot.Slot) {
	if !reflect.DeepEqual(expected, result) {
		t.Error("Slots did not match:")
		t.Logf("Expected: %#v", expected)
		t.Logf("Result:   %#v", result)
	}
}

func TestRecipeSet_Match(t *testing.T) {
	itemTypes := createItemTypes()

	reader := strings.NewReader(threeRecipes)
	recipes, err := LoadRecipes(reader, itemTypes)
	if err != nil {
		t.Fatal("Failed to load recipes for match test")
	}

	empty := *Slot(itemTypes, ItemTypeIdNull, 0, 0)
	plank := *Slot(itemTypes, 5, 1, 0)
	log := *Slot(itemTypes, 17, 1, 0)
	flintAndSteel := *Slot(itemTypes, 259, 1, 0)
	iron := *Slot(itemTypes, 265, 1, 0)
	flint := *Slot(itemTypes, 318, 1, 0)

	tests := []struct {
		comment string
		width   int
		height  int
		input   []slot.Slot
		expect  *slot.Slot
	}{
		// A plank in any of 4 slots should produce nothing.
		{
			"P.\n..",
			2, 2,
			Slots(plank, empty, empty, empty),
			&empty,
		},
		{
			".P\n..",
			2, 2,
			Slots(empty, plank, empty, empty),
			&empty,
		},
		{
			"..\nP.",
			2, 2,
			Slots(empty, empty, plank, empty),
			&empty,
		},
		{
			"..\n.P",
			2, 2,
			Slots(empty, empty, empty, plank),
			&empty,
		},
		// A log in any of 4 slots should produce 4 planks.
		{
			"L.\n..",
			2, 2,
			Slots(log, empty, empty, empty),
			Slot(itemTypes, 5, 4, 0),
		},
		{
			".L\n..",
			2, 2,
			Slots(empty, log, empty, empty),
			Slot(itemTypes, 5, 4, 0),
		},
		{
			"..\nL.",
			2, 2,
			Slots(empty, empty, log, empty),
			Slot(itemTypes, 5, 4, 0),
		},
		{
			"..\n.L",
			2, 2,
			Slots(empty, empty, empty, log),
			Slot(itemTypes, 5, 4, 0),
		},
		// Flint and steel
		{
			"F.\n.I",
			2, 2,
			Slots(flint, empty, empty, iron),
			&flintAndSteel,
		},
		{
			"F..\n.I.\n...",
			3, 3,
			Slots(flint, empty, empty, empty, iron, empty, empty, empty, empty),
			&flintAndSteel,
		},
		{
			"...\n.F.\n..I",
			3, 3,
			Slots(empty, empty, empty, empty, flint, empty, empty, empty, iron),
			&flintAndSteel,
		},
	}

	for i := range tests {
		test := &tests[i]
		t.Logf("Test #%d:\n%s", i, test.comment)
		output := recipes.Match(test.width, test.height, test.input)
		assertSlotEq(t, test.expect, &output)
	}

	// TODO test things other than square or 1x1 recipes
	// TODO test recipes with gaps in
}

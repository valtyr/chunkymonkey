package gamerules

import (
	"fmt"
	"runtime"
	"testing"

	. "chunkymonkey/types"
)

var (
	oneCoal = Slot{263, 1, 0}
	oneIronOre = Slot{15, 1, 0}
	oneIronIngot = Slot{265, 1, 0}
	emptySlot = Slot{}
)

func callerDesc(skip int) string {
	if _, file, line, ok := runtime.Caller(skip+1); ok {
		return fmt.Sprintf("%s:%d", file, line)
	}
	return "<unknown code>"
}

func checkLit(t *testing.T, furnace *FurnaceInventory, expectLit bool) {
	if expectLit != furnace.IsLit() {
		var msg string
		if expectLit {
			msg = "lit";
		} else {
			msg = "unlit";
		}
		t.Errorf("%s: Expected furnace to be %s", callerDesc(1), msg)
	}
}

func checkTx(t *testing.T, expected, txState TxState) {
	if expected != txState {
		t.Errorf("%s: Expected tx state=%v, got %v", callerDesc(1), expected, txState)
	}
}

func checkSlot(t *testing.T, expected, result Slot) {
	if !expected.Equals(&result) {
		t.Errorf("%s: Expected slot=%v, got %v", callerDesc(1), expected, result)
	}
}

// Creates a furnace, and loads with one coal and one iron ore, performing
// common tests along the way.
func loadedFurnace(t *testing.T) (furnace *FurnaceInventory) {
	var txState TxState
	var click Click
	furnace = NewFurnaceInventory()

	checkLit(t, furnace, false)

	// Put a coal into the fuel slot.
	click = Click{
		furnaceSlotFuel,
		oneCoal,
		false, false,
		0,
		emptySlot,
	}
	txState = furnace.Click(&click)
	checkTx(t, TxStateAccepted, txState)
	checkSlot(t, emptySlot, click.Cursor)

	checkLit(t, furnace, false)

	// Put iron ore into the reagent slot.
	click = Click{
		furnaceSlotReagent,
		oneIronOre,
		false, false,
		0,
		emptySlot,
	}
	txState = furnace.Click(&click)
	checkTx(t, TxStateAccepted, txState)
	checkSlot(t, emptySlot, click.Cursor)

	// Should now be lit
	checkLit(t, furnace, true)

	// The fuel from the fuel slot should have been consumed.
	checkSlot(t, emptySlot, furnace.slots[furnaceSlotFuel])

	// The output should be empty.
	checkSlot(t, emptySlot, furnace.slots[furnaceSlotOutput])

	return
}

func Test_FurnaceProducesIronIngot(t *testing.T) {
	furnace := loadedFurnace(t)

	// reactionDuration-1 ticks later the output should still be empty.
	for i := Ticks(0); i < reactionDuration-1; i++ {
		furnace.Tick()
	}
	checkSlot(t, emptySlot, furnace.slots[furnaceSlotOutput])

	// One tick later there should be an iron ingot present, and the furnace
	// should still be lit.
	furnace.Tick()
	checkSlot(t, oneIronIngot, furnace.slots[furnaceSlotOutput])
}

func Test_FurnaceFinishBurning(t *testing.T) {
	furnace := loadedFurnace(t)

	// 1599 ticks later there should be an iron ingot present, and the furnace
	// should still be lit.
	for i := Ticks(0); i < 1599; i++ {
		furnace.Tick()
	}
	checkSlot(t, oneIronIngot, furnace.slots[furnaceSlotOutput])
	checkLit(t, furnace, true)

	// One more tick later and the furnace should be unlit.
	furnace.Tick()
	checkLit(t, furnace, false)
}

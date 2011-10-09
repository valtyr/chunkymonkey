package gamerules

import (
	"fmt"
	"io"
	"json"
	"os"

	"chunkymonkey/types"
)

// FurnaceData contains data on furnace reactions.
type FurnaceData struct {
	// FuelDuration contains a map of fuel types to number of ticks that the fuel
	// lasts for.
	Fuels map[types.ItemTypeId]types.Ticks

	// Reactions contains a map of input item type to output item type and data.
	Reactions map[types.ItemTypeId]Reaction
}

// Reaction describes the output of a furnace reaction.
type Reaction struct {
	Output     types.ItemTypeId
	OutputData types.ItemData
}

// furnaceDataDef is used in unmarshalling data from the JSON definition of
// FurnaceData.
type furnaceDataDef struct {
	Fuels []struct {
		Id        types.ItemTypeId
		FuelTicks types.Ticks
	}
	Reactions []struct {
		Comment    string
		Input      types.ItemTypeId
		Output     types.ItemTypeId
		OutputData types.ItemData
	}
}

// LoadFurnaceData reads FurnaceData from the reader.
func LoadFurnaceData(reader io.Reader) (furnaceData FurnaceData, err os.Error) {
	decoder := json.NewDecoder(reader)

	var dataDef furnaceDataDef

	err = decoder.Decode(&dataDef)
	if err != nil {
		return
	}

	furnaceData.Fuels = make(map[types.ItemTypeId]types.Ticks)
	for _, fuelDef := range dataDef.Fuels {
		if _, ok := Items[fuelDef.Id]; !ok {
			err = fmt.Errorf("Furnace fuel type %d is unknown item type ID", fuelDef.Id)
			return
		}
		furnaceData.Fuels[fuelDef.Id] = fuelDef.FuelTicks
	}

	furnaceData.Reactions = make(map[types.ItemTypeId]Reaction)
	for _, reactionDef := range dataDef.Reactions {
		if _, ok := Items[reactionDef.Input]; !ok {
			err = fmt.Errorf(
				"Furnace reaction %q has unknown input item type ID %d",
				reactionDef.Comment, reactionDef.Input)
			return
		}
		if _, ok := Items[reactionDef.Output]; !ok {
			err = fmt.Errorf(
				"Furnace reaction %q has unknown output item type ID %d",
				reactionDef.Comment, reactionDef.Output)
			return
		}
		furnaceData.Reactions[reactionDef.Input] = Reaction{
			Output:     reactionDef.Output,
			OutputData: reactionDef.OutputData,
		}
	}

	return
}

// LoadFurnaceDataFromFile reads FurnaceData from the named file.
func LoadFurnaceDataFromFile(filename string) (furnaceData FurnaceData, err os.Error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	return LoadFurnaceData(file)
}

package gamerules

import (
	"os"

	"chunkymonkey/command"
)

// GameRules is a container type for block, item and recipe definitions.
type GameRules struct {
	BlockTypes       BlockTypeList
	ItemTypes        ItemTypeMap
	Recipes          *RecipeSet
	FurnaceData      FurnaceData
	CommandFramework *command.CommandFramework
}

func LoadGameRules(blocksDefFile, itemsDefFile, recipesDefFile, furnaceDefFile string) (rules *GameRules, err os.Error) {
	blockTypes, err := LoadBlocksFromFile(blocksDefFile)
	if err != nil {
		return
	}

	itemTypes, err := LoadItemTypesFromFile(itemsDefFile)
	if err != nil {
		return
	}

	blockTypes.CreateBlockItemTypes(itemTypes)

	recipes, err := LoadRecipesFromFile(recipesDefFile, itemTypes)
	if err != nil {
		return
	}

	furnaceData, err := LoadFurnaceDataFromFile(furnaceDefFile)
	if err != nil {
		return
	}

	// TODO: Load the prefix from a config file
	cmdFramework := command.NewCommandFramework("/")

	rules = &GameRules{
		BlockTypes:       blockTypes,
		ItemTypes:        itemTypes,
		Recipes:          recipes,
		FurnaceData:      furnaceData,
		CommandFramework: cmdFramework,
	}

	return
}

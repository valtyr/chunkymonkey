// Utility to perform basic checks on supplied data files for blocks, items and
// recipes.
package main

import (
	"flag"
	"fmt"
	"os"

	"chunkymonkey/gamerules"
)

var blockDefs = flag.String(
	"blocks", "blocks.json",
	"The JSON file containing block type definitions.")

var itemDefs = flag.String(
	"items", "items.json",
	"The JSON file containing item type definitions.")

var recipeDefs = flag.String(
	"recipes", "recipes.json",
	"The JSON file containing recipe definitions.")

var furnaceDefs = flag.String(
	"furnace", "furnace.json",
	"The JSON file containing furnace fuel and reaction definitions.")

func main() {
	var err os.Error

	blockTypes, err := gamerules.LoadBlocksFromFile(*blockDefs)
	if err != nil {
		fmt.Fprintf(os.Stdout, "Error loading block definitions: %v\n", err)
		os.Exit(1)
	}

	itemTypes, err := gamerules.LoadItemTypesFromFile(*itemDefs)
	if err != nil {
		fmt.Fprintf(os.Stdout, "Error loading item definitions: %v\n", err)
		os.Exit(1)
	}

	blockTypes.CreateBlockItemTypes(itemTypes)

	_, err = gamerules.LoadRecipesFromFile(*recipeDefs, itemTypes)
	if err != nil {
		fmt.Fprintf(os.Stdout, "Error loading recipe definitions: %v\n", err)
		os.Exit(1)
	}

	_, err = gamerules.LoadFurnaceDataFromFile(*furnaceDefs)
	if err != nil {
		fmt.Fprintf(os.Stdout, "Error loading furnace data definitions: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("PASS")
}

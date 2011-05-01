// Utility to perform basic checks on supplied data files for blocks, items and
// recipes.
package main

import (
	"flag"
	"log"
	"os"

	"chunkymonkey/block"
	"chunkymonkey/itemtype"
	"chunkymonkey/recipe"
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

func main() {
	var err os.Error

	blockTypes, err := block.LoadBlocksFromFile(*blockDefs)
	if err != nil {
		log.Print("Error loading block definitions: ", err)
		os.Exit(1)
	}

	itemTypes, err := itemtype.LoadItemTypesFromFile(*itemDefs)
	if err != nil {
		log.Print("Error loading item definitions: ", err)
		os.Exit(1)
	}

	blockTypes.CreateBlockItemTypes(itemTypes)

	recipes, err := recipe.LoadRecipesFromFile(*recipeDefs, itemTypes)
	if err != nil {
		log.Print("Error loading recipe definitions: ", err)
		os.Exit(1)
	}

	log.Printf(
		"Successfully loaded %d block types, %d item types and %d recipes.",
		len(blockTypes), len(itemTypes), len(recipes.Recipes))
}

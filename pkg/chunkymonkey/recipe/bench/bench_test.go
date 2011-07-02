// Benchmarks for chunkymonkey/recipe. These need to be in here rather than in
// chunkymonkey/recipe to avoid a build dependency loop.
package bench

import (
	"os"
	"testing"

	"chunkymonkey/block"
	"chunkymonkey/itemtype"
	"chunkymonkey/recipe"
	. "chunkymonkey/types"
)

func loadRecipesAndItems() (recipes *recipe.RecipeSet, itemTypes itemtype.ItemTypeMap, err os.Error) {
	blockTypes, err := block.LoadBlocksFromFile("blocks.json")
	if err != nil {
		return
	}

	itemTypes, err = itemtype.LoadItemTypesFromFile("items.json")
	if err != nil {
		return
	}

	blockTypes.CreateBlockItemTypes(itemTypes)

	recipes, err = recipe.LoadRecipesFromFile("recipes.json", itemTypes)
	if err != nil {
		return
	}

	return
}

func Benchmark_RecipeSet_Match_Simple2x2(b *testing.B) {
	recipes, itemTypes, err := loadRecipesAndItems()
	if err != nil {
		panic(err)
	}

	empty := *recipe.Slot(itemTypes, ItemTypeIdNull, 0, 0)
	log := *recipe.Slot(itemTypes, 17, 1, 0)

	inputs := recipe.Slots(log, empty, empty, empty)

	var matcher recipe.RecipeSetMatcher
	matcher.Init(recipes)

	b.ResetTimer()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		matcher.Match(2, 2, inputs)
	}
}

func Benchmark_RecipeSet_Match_Nothing2x2(b *testing.B) {
	recipes, itemTypes, err := loadRecipesAndItems()
	if err != nil {
		panic(err)
	}

	log := *recipe.Slot(itemTypes, 17, 1, 0)

	inputs := recipe.Slots(log, log, log, log)

	var matcher recipe.RecipeSetMatcher
	matcher.Init(recipes)

	b.ResetTimer()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		matcher.Match(2, 2, inputs)
	}
}

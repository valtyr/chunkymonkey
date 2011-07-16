package gamerules

func init() {
	if err := LoadGameRules("blocks.json", "items.json", "recipes.json", "furnace.json"); err != nil {
		panic(err)
	}
}

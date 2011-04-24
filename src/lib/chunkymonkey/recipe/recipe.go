// The recipe package covers the logic and data for crafting items from other
// items in the inventory or workbench.
package recipe

import (
    "chunkymonkey/itemtype"
    . "chunkymonkey/types"
)

const (
    maxRecipeWidth  = 3
    maxRecipeHeight = 3
)

type RecipeSlot struct {
    Type *itemtype.ItemType
    Data ItemData
}

type Recipe struct {
    Comment     string
    Width       byte
    Height      byte
    Input       []RecipeSlot
    Output      RecipeSlot
    OutputCount ItemCount
}

type RecipeSet struct {
    Recipes []Recipe
    // TODO Recipe index data for fast lookup by recipe dimensions, or maybe by
    // some sort of recipe input hash?
}

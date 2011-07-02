// The recipe package covers the logic and data for crafting items from other
// items in the inventory or workbench.
package recipe

import (
	"chunkymonkey/slot"
)

const (
	fnv1_32_offset = 2166136261
	fnv1_32_prime  = 16777619
)

const (
	maxRecipeWidth  = 3
	maxRecipeHeight = 3
)

type Recipe struct {
	Comment string
	Width   byte
	Height  byte
	Input   []slot.Slot
	Output  slot.Slot
}

func (r *Recipe) match(width, height byte, slots []*slot.Slot) (isMatch bool) {
	isMatch = false
	if width != r.Width || height != r.Height {
		return
	}
	for i := range r.Input {
		inSlot := slots[i]
		rSlot := &r.Input[i]
		if inSlot.ItemType != rSlot.ItemType || inSlot.Data != rSlot.Data {
			return
		}

	}
	isMatch = true
	return
}

func (r *Recipe) hash() (hash uint32) {
	slots := make([]*slot.Slot, len(r.Input))
	for i := range r.Input {
		slots[i] = &r.Input[i]
	}
	return inputHash(slots)
}

func inputHash(slots []*slot.Slot) (hash uint32) {
	// Hash based on FNV-1a.
	hash = fnv1_32_offset
	for _, slot := range slots {
		// Hash the lower 16 bits of the item type ID.
		itemTypeId := slot.ItemTypeId()
		hash ^= uint32(itemTypeId & 0xff)
		hash *= fnv1_32_prime
		hash ^= uint32((itemTypeId >> 8) & 0xff)
		hash *= fnv1_32_prime

		// Hash the data value
		hash ^= uint32(slot.Data & 0xff)
		hash *= fnv1_32_prime
	}

	return
}

type RecipeSet struct {
	recipes []Recipe

	// TODO Recipe index data for fast lookup by recipe dimensions, or maybe by
	// some sort of recipe input hash?
	// Recipe by inputs hash.
	recipeHash map[uint32][]*Recipe
}

func (r *RecipeSet) Init() {
	r.recipeHash = make(map[uint32][]*Recipe)
	for i := range r.recipes {
		recipe := &r.recipes[i]
		hash := recipe.hash()
		bucket := r.recipeHash[hash]
		bucket = append(bucket, recipe)
		r.recipeHash[hash] = bucket
	}
}

// Match looks for a matching recipe for the input slots, and returns a Slot
// with the result of any matching recipe. output.ItemType==nil and
// output.Count==0 if nothing matched).
//
// The order of slots is left to right, then top to bottom.
//
// Precondition: len(slots) == width * height
func (r *RecipeSet) Match(width, height int, slots []slot.Slot) (output slot.Slot) {

	// Precondition check.
	if width*height != len(slots) || width > maxRecipeWidth || height > maxRecipeHeight {
		return
	}

	minX, minY := int(width), int(height)
	maxX, maxY := 0, 0
	curIndex := 0
	// Find the position and size of the smallest rectangle that contains all
	// non-empty slots.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			hasItem := slots[curIndex].ItemType != nil
			if hasItem {
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
				if x > maxX {
					maxX = x
				}
				if y > maxY {
					maxY = y
				}
			}
			curIndex++
		}
	}

	widthUsed := 1 + maxY - minY
	heightUsed := 1 + maxX - minX
	// Empty grid.
	if widthUsed <= 0 || heightUsed <= 0 {
		return
	}

	// Make used rectangle into a linear array of []*Slot.
	rectSlots := make([]*slot.Slot, widthUsed*heightUsed)
	outIndex := 0
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			inIndex := y*width + x
			rectSlots[outIndex] = &slots[inIndex]
			outIndex++
		}
	}

	hash := inputHash(rectSlots)

	bucket, ok := r.recipeHash[hash]

	if !ok {
		return
	}

	// Find the matching recipe, if any.
	for i := range bucket {
		recipe := bucket[i]
		if recipe.match(byte(widthUsed), byte(heightUsed), rectSlots) {
			// Found matching recipe.
			output = recipe.Output
		}
	}

	return
}

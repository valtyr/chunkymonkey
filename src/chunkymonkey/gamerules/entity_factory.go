package gamerules

import (
	. "chunkymonkey/types"
)

// NewEntityByTypeName creates the appropriate entity type based on the input
// string, e.g "Item" or "Zombie". Returns nil if typeName is unknown.
func NewEntityByTypeName(typeName string) (entity INonPlayerEntity) {
	switch typeName {
	case "Item":
		entity = new(Item)

		// Mobs
	case "Chicken":
		entity = NewHen()
	case "Cow":
		entity = NewCow()
	case "Creeper":
		entity = NewCreeper()
	case "Pig":
		entity = NewPig()
	case "Sheep":
		entity = NewSheep()
	case "Skeleton":
		entity = NewSkeleton()
	case "Squid":
		entity = NewSquid()
	case "Spider":
		entity = NewSpider()
	case "Wolf":
		entity = NewWolf()
	case "Zombie":
		entity = NewZombie()

		// Objects
	default:
		// Handle all other objects
		if objType, ok := ObjTypeMap[typeName]; ok {
			entity = NewObject(objType)
		} else {
			entity = nil
		}
	}

	return
}

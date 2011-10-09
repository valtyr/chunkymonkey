package gamerules

import (
	"chunkymonkey/types"
)

type MobType struct {
	Id   types.EntityMobType
	Name string
}

type MobTypeMap map[types.EntityMobType]*MobType

// Used for protocol parsing.
var Mobs = MobTypeMap{
	types.MobTypeIdCreeper:      &CreeperType,
	types.MobTypeIdSkeleton:     &SkeletonType,
	types.MobTypeIdSpider:       &SpiderType,
	types.MobTypeIdGiantZombie:  &GiantZombieType,
	types.MobTypeIdZombie:       &ZombieType,
	types.MobTypeIdSlime:        &SlimeType,
	types.MobTypeIdGhast:        &GhastType,
	types.MobTypeIdZombiePigman: &ZombiePigmanType,
	types.MobTypeIdPig:          &PigType,
	types.MobTypeIdSheep:        &SheepType,
	types.MobTypeIdCow:          &CowType,
	types.MobTypeIdHen:          &HenType,
	types.MobTypeIdSquid:        &SquidType,
	types.MobTypeIdWolf:         &WolfType,
}

var CreeperType = MobType{types.MobTypeIdCreeper, "creeper"}
var SkeletonType = MobType{types.MobTypeIdSkeleton, "skeleton"}
var SpiderType = MobType{types.MobTypeIdSpider, "spider"}
var GiantZombieType = MobType{types.MobTypeIdGiantZombie, "giantzombie"}
var ZombieType = MobType{types.MobTypeIdZombie, "zombie"}
var SlimeType = MobType{types.MobTypeIdSlime, "slime"}
var GhastType = MobType{types.MobTypeIdGhast, "ghast"}
var ZombiePigmanType = MobType{types.MobTypeIdZombiePigman, "zombiepigman"}
var PigType = MobType{types.MobTypeIdPig, "pig"}
var SheepType = MobType{types.MobTypeIdSheep, "sheep"}
var CowType = MobType{types.MobTypeIdCow, "cow"}
var HenType = MobType{types.MobTypeIdHen, "hen"}
var SquidType = MobType{types.MobTypeIdSquid, "squid"}
var WolfType = MobType{types.MobTypeIdWolf, "wolf"}

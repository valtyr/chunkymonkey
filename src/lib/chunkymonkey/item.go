package chunkymonkey

import (
    "io"
    "os"

    "chunkymonkey/proto"
    .   "chunkymonkey/types"
)

const (
    // TODO add in new items from Beta 1.2
    ItemIDIronSpade           = ItemID(256)
    ItemIDIronPickaxe         = ItemID(257)
    ItemIDIronAxe             = ItemID(258)
    ItemIDFlintAndSteel       = ItemID(259)
    ItemIDApple               = ItemID(260)
    ItemIDBow                 = ItemID(261)
    ItemIDArrow               = ItemID(262)
    ItemIDCoal                = ItemID(263)
    ItemIDDiamond             = ItemID(264)
    ItemIDIronIngot           = ItemID(265)
    ItemIDGoldIngot           = ItemID(266)
    ItemIDIronSword           = ItemID(267)
    ItemIDWoodenSword         = ItemID(268)
    ItemIDWoodenSpade         = ItemID(269)
    ItemIDWoodenPickaxe       = ItemID(270)
    ItemIDWoodenAxe           = ItemID(271)
    ItemIDStoneSword          = ItemID(272)
    ItemIDStoneSpade          = ItemID(273)
    ItemIDStonePickaxe        = ItemID(274)
    ItemIDStoneAxe            = ItemID(275)
    ItemIDDiamondSword        = ItemID(276)
    ItemIDDiamondSpade        = ItemID(277)
    ItemIDDiamondPickaxe      = ItemID(278)
    ItemIDDiamondAxe          = ItemID(279)
    ItemIDStick               = ItemID(280)
    ItemIDBowl                = ItemID(281)
    ItemIDMushroomSoup        = ItemID(282)
    ItemIDGoldSword           = ItemID(283)
    ItemIDGoldSpade           = ItemID(284)
    ItemIDGoldPickaxe         = ItemID(285)
    ItemIDGoldAxe             = ItemID(286)
    ItemIDString              = ItemID(287)
    ItemIDFeather             = ItemID(288)
    ItemIDGunpowder           = ItemID(289)
    ItemIDWoodenHoe           = ItemID(290)
    ItemIDStoneHoe            = ItemID(291)
    ItemIDIronHoe             = ItemID(292)
    ItemIDDiamondHoe          = ItemID(293)
    ItemIDGoldHoe             = ItemID(294)
    ItemIDSeeds               = ItemID(295)
    ItemIDWheat               = ItemID(296)
    ItemIDBread               = ItemID(297)
    ItemIDLeatherHelmet       = ItemID(298)
    ItemIDLeatherChestplate   = ItemID(299)
    ItemIDLeatherLeggings     = ItemID(300)
    ItemIDLeatherBoots        = ItemID(301)
    ItemIDChainmailHelmet     = ItemID(302)
    ItemIDChainmailChestplate = ItemID(303)
    ItemIDChainmailLeggings   = ItemID(304)
    ItemIDChainmailBoots      = ItemID(305)
    ItemIDIronHelmet          = ItemID(306)
    ItemIDIronChestplate      = ItemID(307)
    ItemIDIronLeggings        = ItemID(308)
    ItemIDIronBoots           = ItemID(309)
    ItemIDDiamondHelmet       = ItemID(310)
    ItemIDDiamondChestplate   = ItemID(311)
    ItemIDDiamondLeggings     = ItemID(312)
    ItemIDDiamondBoots        = ItemID(313)
    ItemIDGoldHelmet          = ItemID(314)
    ItemIDGoldChestplate      = ItemID(315)
    ItemIDGoldLeggings        = ItemID(316)
    ItemIDGoldBoots           = ItemID(317)
    ItemIDFlint               = ItemID(318)
    ItemIDPork                = ItemID(319)
    ItemIDGrilledPork         = ItemID(320)
    ItemIDPaintings           = ItemID(321)
    ItemIDGoldenapple         = ItemID(322)
    ItemIDSign                = ItemID(323)
    ItemIDWoodendoor          = ItemID(324)
    ItemIDBucket              = ItemID(325)
    ItemIDWaterbucket         = ItemID(326)
    ItemIDLavabucket          = ItemID(327)
    ItemIDMinecart            = ItemID(328)
    ItemIDSaddle              = ItemID(329)
    ItemIDIronDoor            = ItemID(330)
    ItemIDRedstone            = ItemID(331)
    ItemIDSnowball            = ItemID(332)
    ItemIDBoat                = ItemID(333)
    ItemIDLeather             = ItemID(334)
    ItemIDMilkBucket          = ItemID(335)
    ItemIDClayBrick           = ItemID(336)
    ItemIDClayBalls           = ItemID(337)
    ItemIDSugarCane           = ItemID(338)
    ItemIDPaper               = ItemID(339)
    ItemIDBook                = ItemID(340)
    ItemIDSlimeBall           = ItemID(341)
    ItemIDStorageMinecart     = ItemID(342)
    ItemIDPoweredMinecart     = ItemID(343)
    ItemIDEgg                 = ItemID(344)
    ItemIDCompass             = ItemID(345)
    ItemIDFishingRod          = ItemID(346)
    ItemIDWatch               = ItemID(347)
    ItemIDGlowstoneDust       = ItemID(348)
    ItemIDRawFish             = ItemID(349)
    ItemIDCookedFish          = ItemID(350)
    ItemIDGoldRecord          = ItemID(2256)
)

type Item struct {
    Entity
    itemType    ItemID
    count       ItemCount
    position    AbsIntXYZ
    velocity    Velocity
    orientation OrientationBytes
}

func NewItem(game *Game, itemType ItemID, count ItemCount, position *AbsIntXYZ, velocity *Velocity) {
    item := &Item{
        itemType: itemType,
        count:    count,
        position: *position,
        velocity: *velocity,
        // TODO proper orientation
        orientation: OrientationBytes{0, 0, 0},
    }

    game.Enqueue(func(game *Game) {
        game.AddItem(item)
    })
}

func (item *Item) SendSpawn(writer io.Writer) (err os.Error) {
    // TODO pass uses value instead of 0
    err = proto.WriteItemSpawn(writer, item.EntityID, item.itemType, item.count, 0, &item.position, &item.orientation)
    if err != nil {
        return
    }

    err = proto.WriteEntityVelocity(writer, item.EntityID, &item.velocity)
    if err != nil {
        return
    }

    return
}

func (item *Item) SendUpdate(writer io.Writer) (err os.Error) {
    if err = proto.WriteEntity(writer, item.Entity.EntityID); err != nil {
        return
    }
    // TODO don't send movement/velocity packets when item hasn't moved
    // TODO optimise bandwidth to use WriteEntityRelMove when possible
    if err = proto.WriteEntityTeleport(writer, item.Entity.EntityID, &item.position, &LookBytes{0, 0}); err != nil {
        return
    }
    if err = proto.WriteEntityVelocity(writer, item.Entity.EntityID, &item.velocity); err != nil {
        return
    }

    return
}

const (
    // Guestimated gravity value. Unknown how accurate this is.
    gravityBlocksPerSecond2 = 3.0

    gravityMilliPixelsPerTick2 = (MilliPixelsPerBlock * gravityBlocksPerSecond2) / ticksPerSecond

    // Air resistance, as a denominator of a velocity component
    airResistance = 5
)

func (item *Item) PhysicsTick() (itemDestroyed bool) {
    itemDestroyed = false

    var vx, vy, vz int32
    vx = int32(item.velocity.X)
    vy = int32(item.velocity.Y)
    vz = int32(item.velocity.Z)

    vy -= gravityMilliPixelsPerTick2

    // TODO rethink air resistance. it does pretty much nothing to counter
    // gravity
    vx -= vx / airResistance
    vy -= vy / airResistance
    vz -= vz / airResistance

    // Scrub out any residual slow velocity so that things can come to a stop
    if vx > -airResistance && vx < airResistance {
        vx = 0
    }
    if vy > -airResistance && vy < airResistance {
        vy = 0
    }
    if vz > -airResistance && vz < airResistance {
        vz = 0
    }

    item.velocity.X = VelocityComponentConstrained(vx)
    item.velocity.Y = VelocityComponentConstrained(vy)
    item.velocity.Z = VelocityComponentConstrained(vz)

    // TODO project the velocity in block space to see if we hit anything
    // solid, and stop the item's velocity if so

    item.position.X += AbsIntCoord(item.velocity.X / MilliPixelsPerPixel)
    item.position.Y += AbsIntCoord(item.velocity.Y / MilliPixelsPerPixel)
    item.position.Z += AbsIntCoord(item.velocity.Z / MilliPixelsPerPixel)

    // Destroy item if fallen out the bottom of the world
    if item.position.Y < 0 {
        itemDestroyed = true
    }
    return
}

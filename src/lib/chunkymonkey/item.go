package item

import (
    "io"
    "os"

    "chunkymonkey/entity"
    "chunkymonkey/physics"
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
    entity.Entity
    itemType    ItemID
    count       ItemCount
    physObj     physics.PointObject
    orientation OrientationBytes
}

func NewItem(itemType ItemID, count ItemCount, position *AbsXYZ, velocity *AbsVelocity) (item *Item) {
    item = &Item{
        itemType: itemType,
        count:    count,
        // TODO proper orientation
        orientation: OrientationBytes{0, 0, 0},
    }
    item.physObj.Init(position, velocity)
    return
}

func (item *Item) GetEntity() *entity.Entity {
    return &item.Entity
}

func (item *Item) GetPosition() *AbsXYZ {
    return &item.physObj.Position
}

func (item *Item) SendSpawn(writer io.Writer) (err os.Error) {
    // TODO pass uses value instead of 0
    err = proto.WriteItemSpawn(
        writer, item.EntityID, item.itemType, item.count, 0,
        &item.physObj.LastSentPosition, &item.orientation)
    if err != nil {
        return
    }

    err = proto.WriteEntityVelocity(writer, item.EntityID, &item.physObj.LastSentVelocity)
    if err != nil {
        return
    }

    return
}

func (item *Item) SendUpdate(writer io.Writer) (err os.Error) {
    if err = proto.WriteEntity(writer, item.Entity.EntityID); err != nil {
        return
    }

    err = item.physObj.SendUpdate(writer, item.Entity.EntityID, &LookBytes{0, 0})

    return
}

func (item *Item) Tick(blockQuery physics.BlockQueryFn) (leftBlock bool) {
    return item.physObj.Tick(blockQuery)
}

package item

import (
    "io"
    "os"

    "chunkymonkey/entity"
    "chunkymonkey/physics"
    "chunkymonkey/proto"
    "chunkymonkey/slot"
    .   "chunkymonkey/types"
)

const (
    // TODO add in new items from Beta 1.2
    ItemIdIronSpade           = ItemId(256)
    ItemIdIronPickaxe         = ItemId(257)
    ItemIdIronAxe             = ItemId(258)
    ItemIdFlintAndSteel       = ItemId(259)
    ItemIdApple               = ItemId(260)
    ItemIdBow                 = ItemId(261)
    ItemIdArrow               = ItemId(262)
    ItemIdCoal                = ItemId(263)
    ItemIdDiamond             = ItemId(264)
    ItemIdIronIngot           = ItemId(265)
    ItemIdGoldIngot           = ItemId(266)
    ItemIdIronSword           = ItemId(267)
    ItemIdWoodenSword         = ItemId(268)
    ItemIdWoodenSpade         = ItemId(269)
    ItemIdWoodenPickaxe       = ItemId(270)
    ItemIdWoodenAxe           = ItemId(271)
    ItemIdStoneSword          = ItemId(272)
    ItemIdStoneSpade          = ItemId(273)
    ItemIdStonePickaxe        = ItemId(274)
    ItemIdStoneAxe            = ItemId(275)
    ItemIdDiamondSword        = ItemId(276)
    ItemIdDiamondSpade        = ItemId(277)
    ItemIdDiamondPickaxe      = ItemId(278)
    ItemIdDiamondAxe          = ItemId(279)
    ItemIdStick               = ItemId(280)
    ItemIdBowl                = ItemId(281)
    ItemIdMushroomSoup        = ItemId(282)
    ItemIdGoldSword           = ItemId(283)
    ItemIdGoldSpade           = ItemId(284)
    ItemIdGoldPickaxe         = ItemId(285)
    ItemIdGoldAxe             = ItemId(286)
    ItemIdString              = ItemId(287)
    ItemIdFeather             = ItemId(288)
    ItemIdGunpowder           = ItemId(289)
    ItemIdWoodenHoe           = ItemId(290)
    ItemIdStoneHoe            = ItemId(291)
    ItemIdIronHoe             = ItemId(292)
    ItemIdDiamondHoe          = ItemId(293)
    ItemIdGoldHoe             = ItemId(294)
    ItemIdSeeds               = ItemId(295)
    ItemIdWheat               = ItemId(296)
    ItemIdBread               = ItemId(297)
    ItemIdLeatherHelmet       = ItemId(298)
    ItemIdLeatherChestplate   = ItemId(299)
    ItemIdLeatherLeggings     = ItemId(300)
    ItemIdLeatherBoots        = ItemId(301)
    ItemIdChainmailHelmet     = ItemId(302)
    ItemIdChainmailChestplate = ItemId(303)
    ItemIdChainmailLeggings   = ItemId(304)
    ItemIdChainmailBoots      = ItemId(305)
    ItemIdIronHelmet          = ItemId(306)
    ItemIdIronChestplate      = ItemId(307)
    ItemIdIronLeggings        = ItemId(308)
    ItemIdIronBoots           = ItemId(309)
    ItemIdDiamondHelmet       = ItemId(310)
    ItemIdDiamondChestplate   = ItemId(311)
    ItemIdDiamondLeggings     = ItemId(312)
    ItemIdDiamondBoots        = ItemId(313)
    ItemIdGoldHelmet          = ItemId(314)
    ItemIdGoldChestplate      = ItemId(315)
    ItemIdGoldLeggings        = ItemId(316)
    ItemIdGoldBoots           = ItemId(317)
    ItemIdFlint               = ItemId(318)
    ItemIdPork                = ItemId(319)
    ItemIdGrilledPork         = ItemId(320)
    ItemIdPaintings           = ItemId(321)
    ItemIdGoldenapple         = ItemId(322)
    ItemIdSign                = ItemId(323)
    ItemIdWoodendoor          = ItemId(324)
    ItemIdBucket              = ItemId(325)
    ItemIdWaterbucket         = ItemId(326)
    ItemIdLavabucket          = ItemId(327)
    ItemIdMinecart            = ItemId(328)
    ItemIdSaddle              = ItemId(329)
    ItemIdIronDoor            = ItemId(330)
    ItemIdRedstone            = ItemId(331)
    ItemIdSnowball            = ItemId(332)
    ItemIdBoat                = ItemId(333)
    ItemIdLeather             = ItemId(334)
    ItemIdMilkBucket          = ItemId(335)
    ItemIdClayBrick           = ItemId(336)
    ItemIdClayBalls           = ItemId(337)
    ItemIdSugarCane           = ItemId(338)
    ItemIdPaper               = ItemId(339)
    ItemIdBook                = ItemId(340)
    ItemIdSlimeBall           = ItemId(341)
    ItemIdStorageMinecart     = ItemId(342)
    ItemIdPoweredMinecart     = ItemId(343)
    ItemIdEgg                 = ItemId(344)
    ItemIdCompass             = ItemId(345)
    ItemIdFishingRod          = ItemId(346)
    ItemIdWatch               = ItemId(347)
    ItemIdGlowstoneDust       = ItemId(348)
    ItemIdRawFish             = ItemId(349)
    ItemIdCookedFish          = ItemId(350)
    ItemIdGoldRecord          = ItemId(2256)
)

type Item struct {
    entity.Entity
    slot.Slot
    physObj     physics.PointObject
    orientation OrientationBytes
}

func NewItem(itemType ItemId, count ItemCount, position *AbsXyz, velocity *AbsVelocity) (item *Item) {
    item = &Item{
        // TODO proper orientation
        orientation: OrientationBytes{0, 0, 0},
    }
    item.Slot.ItemType = itemType
    item.Slot.Count = count
    item.physObj.Init(position, velocity)
    return
}

func (item *Item) GetEntity() *entity.Entity {
    return &item.Entity
}

func (item *Item) GetSlot() *slot.Slot {
    return &item.Slot
}

func (item *Item) GetItemType() ItemId {
    return item.ItemType
}

func (item *Item) GetCount() ItemCount {
    return item.Count
}

func (item *Item) SetCount(count ItemCount) {
    item.Count = count
}

func (item *Item) GetPosition() *AbsXyz {
    return &item.physObj.Position
}

func (item *Item) SendSpawn(writer io.Writer) (err os.Error) {
    // TODO pass uses value instead of 0
    err = proto.WriteItemSpawn(
        writer, item.EntityId, item.ItemType, item.Count, 0,
        &item.physObj.LastSentPosition, &item.orientation)
    if err != nil {
        return
    }

    err = proto.WriteEntityVelocity(writer, item.EntityId, &item.physObj.LastSentVelocity)
    if err != nil {
        return
    }

    return
}

func (item *Item) SendUpdate(writer io.Writer) (err os.Error) {
    if err = proto.WriteEntity(writer, item.Entity.EntityId); err != nil {
        return
    }

    err = item.physObj.SendUpdate(writer, item.Entity.EntityId, &LookBytes{0, 0})

    return
}

func (item *Item) Tick(blockQuery physics.BlockQueryFn) (leftBlock bool) {
    return item.physObj.Tick(blockQuery)
}

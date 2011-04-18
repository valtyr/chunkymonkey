package item

import (
    "io"
    "os"

    "chunkymonkey/entity"
    "chunkymonkey/physics"
    "chunkymonkey/proto"
    "chunkymonkey/slot"
    . "chunkymonkey/types"
)

const (
    // TODO add in new items from Beta 1.2
    ItemIdIronSpade           = ItemTypeId(256)
    ItemIdIronPickaxe         = ItemTypeId(257)
    ItemIdIronAxe             = ItemTypeId(258)
    ItemIdFlintAndSteel       = ItemTypeId(259)
    ItemIdApple               = ItemTypeId(260)
    ItemIdBow                 = ItemTypeId(261)
    ItemIdArrow               = ItemTypeId(262)
    ItemIdCoal                = ItemTypeId(263)
    ItemIdDiamond             = ItemTypeId(264)
    ItemIdIronIngot           = ItemTypeId(265)
    ItemIdGoldIngot           = ItemTypeId(266)
    ItemIdIronSword           = ItemTypeId(267)
    ItemIdWoodenSword         = ItemTypeId(268)
    ItemIdWoodenSpade         = ItemTypeId(269)
    ItemIdWoodenPickaxe       = ItemTypeId(270)
    ItemIdWoodenAxe           = ItemTypeId(271)
    ItemIdStoneSword          = ItemTypeId(272)
    ItemIdStoneSpade          = ItemTypeId(273)
    ItemIdStonePickaxe        = ItemTypeId(274)
    ItemIdStoneAxe            = ItemTypeId(275)
    ItemIdDiamondSword        = ItemTypeId(276)
    ItemIdDiamondSpade        = ItemTypeId(277)
    ItemIdDiamondPickaxe      = ItemTypeId(278)
    ItemIdDiamondAxe          = ItemTypeId(279)
    ItemIdStick               = ItemTypeId(280)
    ItemIdBowl                = ItemTypeId(281)
    ItemIdMushroomSoup        = ItemTypeId(282)
    ItemIdGoldSword           = ItemTypeId(283)
    ItemIdGoldSpade           = ItemTypeId(284)
    ItemIdGoldPickaxe         = ItemTypeId(285)
    ItemIdGoldAxe             = ItemTypeId(286)
    ItemIdString              = ItemTypeId(287)
    ItemIdFeather             = ItemTypeId(288)
    ItemIdGunpowder           = ItemTypeId(289)
    ItemIdWoodenHoe           = ItemTypeId(290)
    ItemIdStoneHoe            = ItemTypeId(291)
    ItemIdIronHoe             = ItemTypeId(292)
    ItemIdDiamondHoe          = ItemTypeId(293)
    ItemIdGoldHoe             = ItemTypeId(294)
    ItemIdSeeds               = ItemTypeId(295)
    ItemIdWheat               = ItemTypeId(296)
    ItemIdBread               = ItemTypeId(297)
    ItemIdLeatherHelmet       = ItemTypeId(298)
    ItemIdLeatherChestplate   = ItemTypeId(299)
    ItemIdLeatherLeggings     = ItemTypeId(300)
    ItemIdLeatherBoots        = ItemTypeId(301)
    ItemIdChainmailHelmet     = ItemTypeId(302)
    ItemIdChainmailChestplate = ItemTypeId(303)
    ItemIdChainmailLeggings   = ItemTypeId(304)
    ItemIdChainmailBoots      = ItemTypeId(305)
    ItemIdIronHelmet          = ItemTypeId(306)
    ItemIdIronChestplate      = ItemTypeId(307)
    ItemIdIronLeggings        = ItemTypeId(308)
    ItemIdIronBoots           = ItemTypeId(309)
    ItemIdDiamondHelmet       = ItemTypeId(310)
    ItemIdDiamondChestplate   = ItemTypeId(311)
    ItemIdDiamondLeggings     = ItemTypeId(312)
    ItemIdDiamondBoots        = ItemTypeId(313)
    ItemIdGoldHelmet          = ItemTypeId(314)
    ItemIdGoldChestplate      = ItemTypeId(315)
    ItemIdGoldLeggings        = ItemTypeId(316)
    ItemIdGoldBoots           = ItemTypeId(317)
    ItemIdFlint               = ItemTypeId(318)
    ItemIdPork                = ItemTypeId(319)
    ItemIdGrilledPork         = ItemTypeId(320)
    ItemIdPaintings           = ItemTypeId(321)
    ItemIdGoldenapple         = ItemTypeId(322)
    ItemIdSign                = ItemTypeId(323)
    ItemIdWoodendoor          = ItemTypeId(324)
    ItemIdBucket              = ItemTypeId(325)
    ItemIdWaterbucket         = ItemTypeId(326)
    ItemIdLavabucket          = ItemTypeId(327)
    ItemIdMinecart            = ItemTypeId(328)
    ItemIdSaddle              = ItemTypeId(329)
    ItemIdIronDoor            = ItemTypeId(330)
    ItemIdRedstone            = ItemTypeId(331)
    ItemIdSnowball            = ItemTypeId(332)
    ItemIdBoat                = ItemTypeId(333)
    ItemIdLeather             = ItemTypeId(334)
    ItemIdMilkBucket          = ItemTypeId(335)
    ItemIdClayBrick           = ItemTypeId(336)
    ItemIdClayBalls           = ItemTypeId(337)
    ItemIdSugarCane           = ItemTypeId(338)
    ItemIdPaper               = ItemTypeId(339)
    ItemIdBook                = ItemTypeId(340)
    ItemIdSlimeBall           = ItemTypeId(341)
    ItemIdStorageMinecart     = ItemTypeId(342)
    ItemIdPoweredMinecart     = ItemTypeId(343)
    ItemIdEgg                 = ItemTypeId(344)
    ItemIdCompass             = ItemTypeId(345)
    ItemIdFishingRod          = ItemTypeId(346)
    ItemIdWatch               = ItemTypeId(347)
    ItemIdGlowstoneDust       = ItemTypeId(348)
    ItemIdRawFish             = ItemTypeId(349)
    ItemIdCookedFish          = ItemTypeId(350)
    ItemIdGoldRecord          = ItemTypeId(2256)
)

type Item struct {
    entity.Entity
    slot.Slot
    physObj     physics.PointObject
    orientation OrientationBytes
}

func NewItem(itemTypeId ItemTypeId, count ItemCount, position *AbsXyz, velocity *AbsVelocity) (item *Item) {
    item = &Item{
        // TODO proper orientation
        orientation: OrientationBytes{0, 0, 0},
    }
    item.Slot.ItemTypeId = itemTypeId
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

func (item *Item) GetItemTypeId() ItemTypeId {
    return item.ItemTypeId
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
        writer, item.EntityId, item.ItemTypeId, item.Count, 0,
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

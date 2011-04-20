package block

import (
    "rand"

    "chunkymonkey/item"
    . "chunkymonkey/types"
)

const (
    // TODO add in new blocks from Beta 1.2
    BlockIdMin = BlockId(0)

    BlockIdAir                 = BlockId(0)
    BlockIdStone               = BlockId(1)
    BlockIdGrass               = BlockId(2)
    BlockIdDirt                = BlockId(3)
    BlockIdCobblestone         = BlockId(4)
    BlockIdPlank               = BlockId(5)
    BlockIdSapling             = BlockId(6)
    BlockIdBedrock             = BlockId(7)
    BlockIdWater               = BlockId(8)
    BlockIdStationaryWater     = BlockId(9)
    BlockIdLava                = BlockId(10)
    BlockIdStationaryLava      = BlockId(11)
    BlockIdSand                = BlockId(12)
    BlockIdGravel              = BlockId(13)
    BlockIdGoldOre             = BlockId(14)
    BlockIdIronOre             = BlockId(15)
    BlockIdCoalOre             = BlockId(16)
    BlockIdLog                 = BlockId(17)
    BlockIdLeaves              = BlockId(18)
    BlockIdSponge              = BlockId(19)
    BlockIdGlass               = BlockId(20)
    BlockIdWool                = BlockId(35)
    BlockIdYellowFlower        = BlockId(37)
    BlockIdRedRose             = BlockId(38)
    BlockIdBrownMushroom       = BlockId(39)
    BlockIdRedMushroom         = BlockId(40)
    BlockIdGoldBlock           = BlockId(41)
    BlockIdIronBlock           = BlockId(42)
    BlockIdDoubleStoneSlab     = BlockId(43)
    BlockIdStoneSlab           = BlockId(44)
    BlockIdBrick               = BlockId(45)
    BlockIdTnt                 = BlockId(46)
    BlockIdBookshelf           = BlockId(47)
    BlockIdMossStone           = BlockId(48)
    BlockIdObsidian            = BlockId(49)
    BlockIdTorch               = BlockId(50)
    BlockIdFire                = BlockId(51)
    BlockIdMobSpawner          = BlockId(52)
    BlockIdWoodenStairs        = BlockId(53)
    BlockIdChest               = BlockId(54)
    BlockIdRedstoneWire        = BlockId(55)
    BlockIdDiamondOre          = BlockId(56)
    BlockIdDiamondBlock        = BlockId(57)
    BlockIdWorkbench           = BlockId(58)
    BlockIdCrops               = BlockId(59)
    BlockIdFarmland            = BlockId(60)
    BlockIdFurnace             = BlockId(61)
    BlockIdBurningFurnace      = BlockId(62)
    BlockIdSignPost            = BlockId(63)
    BlockIdWoodenDoor          = BlockId(64)
    BlockIdLadder              = BlockId(65)
    BlockIdMinecartTracks      = BlockId(66)
    BlockIdCobblestoneStairs   = BlockId(67)
    BlockIdWallSign            = BlockId(68)
    BlockIdLever               = BlockId(69)
    BlockIdStonePressurePlate  = BlockId(70)
    BlockIdIronDoor            = BlockId(71)
    BlockIdWoodenPressurePlate = BlockId(72)
    BlockIdRedstoneOre         = BlockId(73)
    BlockIdGlowingRedstoneOre  = BlockId(74)
    BlockIdRedstoneTorchOff    = BlockId(75)
    BlockIdRedstoneTorchOn     = BlockId(76)
    BlockIdStoneButton         = BlockId(77)
    BlockIdSnow                = BlockId(78)
    BlockIdIce                 = BlockId(79)
    BlockIdSnowBlock           = BlockId(80)
    BlockIdCactus              = BlockId(81)
    BlockIdClay                = BlockId(82)
    BlockIdSugarCane           = BlockId(83)
    BlockIdJukebox             = BlockId(84)
    BlockIdFence               = BlockId(85)
    BlockIdPumpkin             = BlockId(86)
    BlockIdNetherrack          = BlockId(87)
    BlockIdSoulSand            = BlockId(88)
    BlockIdGlowstone           = BlockId(89)
    BlockIdPortal              = BlockId(90)
    BlockIdJackOLantern        = BlockId(91)

    BlockIdMax = BlockId(91)
)

// Defines the behaviour of a block.
type IBlockAspect interface {
    Dig(chunk IChunkBlock, blockLoc *BlockXyz, digStatus DigStatus) (destroyed bool)
}

// The core information about any block type.
type BlockType struct {
    Aspect       *StandardAspect
    Name         string
    Opacity      int8
    Destructable bool
    Solid        bool
    Replaceable  bool
    Attachable   bool
}

// The interface required of a chunk by block behaviour.
type IChunkBlock interface {
    GetRand() *rand.Rand
    AddItem(item *item.Item)
}

// The distance from the edge of a block that items spawn at in fractional
// blocks.
const blockItemSpawnFromEdge = 4.0 / PixelsPerBlock

// Returns true if the block should be destroyed
// This must be called within the Chunk's goroutine.

func (blockType *BlockType) IsSolid() bool {
    return blockType.Solid
}

// Returns true if item can be "replaced" by block placement.
func (blockType *BlockType) IsReplaceable() bool {
    return blockType.Replaceable
}

// Returns true if a block can be placed onto a face of the block.
func (blockType *BlockType) IsAttachable() bool {
    return blockType.Attachable
}

func (blockType *BlockType) GetName() string {
    return blockType.Name
}

func (blockType *BlockType) GetOpacity() int8 {
    return blockType.Opacity
}

func LoadStandardBlockTypes() map[BlockId]*BlockType {
    b := make(map[BlockId]*BlockType)

    newBlock := func(id BlockId, name string) {
        b[id] = &BlockType{
            Aspect: &StandardAspect{
                DroppedItems: nil,
                BreakOn:      DigBlockBroke,
            },
            Name:         name,
            Opacity:      15,
            Destructable: true,
            Solid:        true,
            Replaceable:  false,
            Attachable:   true,
        }
    }

    newBlock(BlockIdAir, "air")
    newBlock(BlockIdStone, "stone")
    newBlock(BlockIdGrass, "grass")
    newBlock(BlockIdDirt, "dirt")
    newBlock(BlockIdCobblestone, "cobblestone")
    newBlock(BlockIdPlank, "wood")
    newBlock(BlockIdSapling, "sapling")
    newBlock(BlockIdBedrock, "bedrock")
    newBlock(BlockIdWater, "water")
    newBlock(BlockIdStationaryWater, "stationary water")
    newBlock(BlockIdLava, "lava")
    newBlock(BlockIdStationaryLava, "stationary lava")
    newBlock(BlockIdSand, "sand")
    newBlock(BlockIdGravel, "gravel")
    newBlock(BlockIdGoldOre, "gold ore")
    newBlock(BlockIdIronOre, "iron ore")
    newBlock(BlockIdCoalOre, "coal ore")
    newBlock(BlockIdLog, "log")
    newBlock(BlockIdLeaves, "leaves")
    newBlock(BlockIdSponge, "sponge")
    newBlock(BlockIdGlass, "glass")
    newBlock(BlockIdWool, "wool")
    newBlock(BlockIdYellowFlower, "yellow flower")
    newBlock(BlockIdRedRose, "red rose")
    newBlock(BlockIdBrownMushroom, "brown mushroom")
    newBlock(BlockIdRedMushroom, "red mushroom")
    newBlock(BlockIdGoldBlock, "gold block")
    newBlock(BlockIdIronBlock, "iron block")
    newBlock(BlockIdDoubleStoneSlab, "double stone slab")
    newBlock(BlockIdStoneSlab, "stone slab")
    newBlock(BlockIdBrick, "brick")
    newBlock(BlockIdTnt, "TNT")
    newBlock(BlockIdBookshelf, "bookshelf")
    newBlock(BlockIdMossStone, "moss stone")
    newBlock(BlockIdObsidian, "obsidian")
    newBlock(BlockIdTorch, "torch")
    newBlock(BlockIdFire, "fire")
    newBlock(BlockIdMobSpawner, "mob spawner")
    newBlock(BlockIdWoodenStairs, "wooden stairs")
    newBlock(BlockIdChest, "chest")
    newBlock(BlockIdRedstoneWire, "redstone wire")
    newBlock(BlockIdDiamondOre, "diamond ore")
    newBlock(BlockIdDiamondBlock, "diamond block")
    newBlock(BlockIdWorkbench, "workbench")
    newBlock(BlockIdCrops, "crops")
    newBlock(BlockIdFarmland, "soil")
    newBlock(BlockIdFurnace, "furnace")
    newBlock(BlockIdBurningFurnace, "burning furnace")
    newBlock(BlockIdSignPost, "sign post")
    newBlock(BlockIdWoodenDoor, "wooden door")
    newBlock(BlockIdLadder, "ladder")
    newBlock(BlockIdMinecartTracks, "minecart tracks")
    newBlock(BlockIdCobblestoneStairs, "cobblestone stairs")
    newBlock(BlockIdWallSign, "wall sign")
    newBlock(BlockIdLever, "lever")
    newBlock(BlockIdStonePressurePlate, "stone pressure plate")
    newBlock(BlockIdIronDoor, "irondoor")
    newBlock(BlockIdWoodenPressurePlate, "wooden pressure plate")
    newBlock(BlockIdRedstoneOre, "redstone ore")
    newBlock(BlockIdGlowingRedstoneOre, "glowing redstone ore")
    newBlock(BlockIdRedstoneTorchOff, "redstone torch off")
    newBlock(BlockIdRedstoneTorchOn, "redstone torch on")
    newBlock(BlockIdStoneButton, "stone button")
    newBlock(BlockIdSnow, "snow")
    newBlock(BlockIdIce, "ice")
    newBlock(BlockIdSnowBlock, "snow block")
    newBlock(BlockIdCactus, "cactus")
    newBlock(BlockIdClay, "clay")
    newBlock(BlockIdSugarCane, "sugar cane")
    newBlock(BlockIdJukebox, "jukebox")
    newBlock(BlockIdFence, "fence")
    newBlock(BlockIdPumpkin, "pumpkin")
    newBlock(BlockIdNetherrack, "netherrack")
    newBlock(BlockIdSoulSand, "soul sand")
    newBlock(BlockIdGlowstone, "glowstone")
    newBlock(BlockIdPortal, "portal")
    newBlock(BlockIdJackOLantern, "jack o lantern")

    setTrans := func(opacity int8, blockTypes []BlockId) {
        for _, blockType := range blockTypes {
            b[blockType].Opacity = opacity
        }
    }
    // Setup transparent blocks
    setTrans(0, []BlockId{BlockIdAir, BlockIdSapling, BlockIdGlass,
        BlockIdYellowFlower, BlockIdRedRose, BlockIdBrownMushroom,
        BlockIdRedMushroom, BlockIdFire, BlockIdMobSpawner, BlockIdWoodenStairs,
        BlockIdRedstoneWire, BlockIdCrops, BlockIdSignPost, BlockIdLadder,
        BlockIdMinecartTracks, BlockIdCobblestoneStairs, BlockIdWallSign,
        BlockIdLever, BlockIdIronDoor, BlockIdRedstoneTorchOff,
        BlockIdRedstoneTorchOn, BlockIdStoneButton, BlockIdSnow, BlockIdCactus,
        BlockIdSugarCane, BlockIdFence, BlockIdPortal})

    // Setup semi-transparent blocks
    setTrans(1, []BlockId{BlockIdLeaves})
    setTrans(3, []BlockId{BlockIdWater, BlockIdStationaryWater, BlockIdIce})

    // Setup non-solid blocks
    nonSolid := []BlockId{
        BlockIdAir, BlockIdSapling, BlockIdWater, BlockIdStationaryWater,
        BlockIdLava, BlockIdStationaryLava, BlockIdYellowFlower,
        BlockIdRedRose, BlockIdBrownMushroom, BlockIdRedMushroom, BlockIdTorch,
        BlockIdFire, BlockIdRedstoneWire, BlockIdCrops, BlockIdSignPost,
        BlockIdLadder, BlockIdMinecartTracks, BlockIdWallSign, BlockIdLever,
        BlockIdStonePressurePlate, BlockIdIronDoor, BlockIdWoodenPressurePlate,
        BlockIdRedstoneTorchOff, BlockIdRedstoneTorchOn, BlockIdStoneButton,
        BlockIdSugarCane, BlockIdPortal,
    }
    for _, blockId := range nonSolid {
        b[blockId].Solid = false
    }

    // Setup replaceable blocks
    setReplaceable := func(blockIds ...BlockId) {
        for _, blockId := range blockIds {
            b[blockId].Replaceable = true
        }
    }
    setReplaceable(
        BlockIdAir,
        BlockIdWater, BlockIdStationaryWater,
        BlockIdLava, BlockIdStationaryLava,
        BlockIdFire,
        BlockIdSnow,
    )

    // Setup non-attachable blocks
    setNonAttachable := func(blockIds ...BlockId) {
        for _, blockId := range blockIds {
            b[blockId].Attachable = false
        }
    }
    setNonAttachable(
        BlockIdAir, BlockIdSapling, BlockIdWater, BlockIdStationaryWater,
        BlockIdLava, BlockIdStationaryLava, BlockIdYellowFlower,
        BlockIdRedRose, BlockIdBrownMushroom, BlockIdRedMushroom,
        BlockIdStoneSlab, BlockIdTorch, BlockIdFire, BlockIdWoodenStairs,
        BlockIdChest, BlockIdRedstoneWire, BlockIdWorkbench, BlockIdCrops,
        BlockIdFurnace, BlockIdBurningFurnace, BlockIdSignPost,
        BlockIdWoodenDoor, BlockIdLadder, BlockIdMinecartTracks,
        BlockIdCobblestoneStairs, BlockIdWallSign, BlockIdLever,
        BlockIdStonePressurePlate, BlockIdIronDoor, BlockIdWoodenPressurePlate,
        BlockIdRedstoneTorchOff, BlockIdRedstoneTorchOn, BlockIdStoneButton,
        BlockIdSnow, BlockIdSugarCane, BlockIdJukebox, BlockIdFence,
        BlockIdPortal,
    )

    // Setup behaviour of blocks when destroyed
    setMinedDropsSameItem := func(blockTypes []BlockId) {
        for _, blockType := range blockTypes {
            b[blockType].Aspect.DroppedItems = append(
                b[blockType].Aspect.DroppedItems,
                BlockDropItem{
                    ItemTypeId(blockType),
                    100,
                    1,
                })
        }
    }

    type Drop struct {
        minedBlockType  BlockId
        droppedItemType ItemTypeId
    }
    setMinedDropBlock := func(drops []Drop) {
        for _, drop := range drops {
            b[drop.minedBlockType].Aspect.DroppedItems = append(
                b[drop.minedBlockType].Aspect.DroppedItems,
                BlockDropItem{
                    drop.droppedItemType,
                    100,
                    1,
                })
        }
    }

    b[BlockIdBedrock].Destructable = false

    // TODO crops are more complicated, and need code to look at their metadata
    // to decide what to drop.
    // TODO ice blocks are more complicated as to what they do when destroyed

    // TODO data about tool usage

    // TODO what item ID drops for redstone torches (on vs off state)

    // Blocks that drop the same ItemTypeId as BlockId 100% of the time
    setMinedDropsSameItem([]BlockId{
        BlockIdDirt, BlockIdCobblestone, BlockIdPlank, BlockIdSapling,
        BlockIdSand, BlockIdGoldOre, BlockIdIronOre, BlockIdLog, BlockIdSponge,
        BlockIdWool, BlockIdYellowFlower, BlockIdRedRose, BlockIdBrownMushroom,
        BlockIdRedMushroom, BlockIdGoldBlock, BlockIdIronBlock,
        BlockIdStoneSlab, BlockIdBrick, BlockIdMossStone, BlockIdObsidian,
        BlockIdTorch, BlockIdWoodenStairs, BlockIdChest, BlockIdDiamondBlock,
        BlockIdWorkbench, BlockIdLadder, BlockIdMinecartTracks,
        BlockIdCobblestoneStairs, BlockIdLever, BlockIdStonePressurePlate,
        BlockIdWoodenPressurePlate, BlockIdStoneButton, BlockIdCactus,
        BlockIdClay, BlockIdJukebox, BlockIdFence, BlockIdPumpkin,
        BlockIdNetherrack, BlockIdSoulSand, BlockIdGlowstone,
        BlockIdJackOLantern,
    })
    // Blocks that drop a single different item 100% of the time
    setMinedDropBlock([]Drop{
        Drop{BlockIdStone, ItemTypeId(BlockIdCobblestone)},
        Drop{BlockIdGrass, ItemTypeId(BlockIdDirt)},
        Drop{BlockIdCoalOre, item.ItemIdCoal},
        Drop{BlockIdDoubleStoneSlab, ItemTypeId(BlockIdStoneSlab)},
        Drop{BlockIdDiamondOre, item.ItemIdDiamond},
        Drop{BlockIdFarmland, ItemTypeId(BlockIdDirt)},
        Drop{BlockIdSignPost, item.ItemIdSign},
        Drop{BlockIdWoodenDoor, item.ItemIdWoodendoor},
        Drop{BlockIdWallSign, item.ItemIdSign},
        Drop{BlockIdIronDoor, item.ItemIdIronDoor},
        Drop{BlockIdSnow, ItemTypeId(BlockIdDirt)},
        Drop{BlockIdSugarCane, item.ItemIdSugarCane},
        Drop{BlockIdGlowstone, item.ItemIdGlowstoneDust},
    })
    // Blocks that drop things with varying probability (or one of several
    // items)
    b[BlockIdGravel].Aspect.DroppedItems = []BlockDropItem{
        BlockDropItem{item.ItemIdFlint, 10, 1},
        BlockDropItem{ItemTypeId(BlockIdGravel), 90, 1},
    }
    b[BlockIdLeaves].Aspect.DroppedItems = []BlockDropItem{
        // TODO get more accurate probability of sapling drop
        BlockDropItem{ItemTypeId(BlockIdSapling), 5, 1},
    }
    b[BlockIdRedstoneOre].Aspect.DroppedItems = []BlockDropItem{
        // TODO find probabilities of dropping 4 vs 5 items
        BlockDropItem{item.ItemIdRedstone, 50, 4},
        BlockDropItem{item.ItemIdRedstone, 50, 5},
    }
    b[BlockIdGlowingRedstoneOre].Aspect.DroppedItems = b[BlockIdRedstoneOre].Aspect.DroppedItems
    b[BlockIdSnowBlock].Aspect.DroppedItems = []BlockDropItem{
        BlockDropItem{item.ItemIdSnowball, 100, 4},
    }

    // Blocks that break on DigStarted
    setBlockBreakOn := func(digStatus DigStatus, blockIds ...BlockId) {
        for _, blockId := range blockIds {
            b[blockId].Aspect.BreakOn = digStatus
        }
    }
    setBlockBreakOn(
        DigStarted,
        BlockIdSapling, BlockIdYellowFlower, BlockIdRedRose,
        BlockIdBrownMushroom, BlockIdRedMushroom, BlockIdTorch, BlockIdCrops,
    )

    retval := make(map[BlockId]*BlockType)
    for k, block := range b {
        retval[k] = block
    }
    return retval
}

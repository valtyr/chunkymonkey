package block

import (
    .      "chunkymonkey/interfaces"
    cmitem "chunkymonkey/item"
    .      "chunkymonkey/types"
)

const (
    // TODO add in new blocks from Beta 1.2
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
)

type BlockDropItem struct {
    droppedItem ItemId
    probability byte // Probabilities specified as a percentage
    quantity    ItemCount
}

type BlockType struct {
    name         string
    transparency int8
    destructable bool
    // Items, up to one of which will potentially spawn when block destroyed
    droppedItems []BlockDropItem
    isSolid      bool
}

// The distance from the edge of a block that items spawn at in fractional
// blocks.
const blockItemSpawnFromEdge = 4.0 / PixelsPerBlock

// Returns true if the block should be destroyed
// This must be called within the Chunk's goroutine.
func (blockType *BlockType) Destroy(chunk IChunk, blockLoc *BlockXyz) bool {
    if len(blockType.droppedItems) > 0 {
        rand := chunk.GetRand()
        // Possibly drop item(s)
        r := byte(rand.Intn(100))
        for _, dropItem := range blockType.droppedItems {
            if dropItem.probability > r {
                for i := dropItem.quantity; i > 0; i-- {
                    position := blockLoc.ToAbsXyz()
                    position.X += AbsCoord(blockItemSpawnFromEdge + rand.Float64()*(1-2*blockItemSpawnFromEdge))
                    position.Y += AbsCoord(blockItemSpawnFromEdge)
                    position.Z += AbsCoord(blockItemSpawnFromEdge + rand.Float64()*(1-2*blockItemSpawnFromEdge))
                    chunk.AddItem(
                        cmitem.NewItem(
                            dropItem.droppedItem, 1,
                            position,
                            &AbsVelocity{0, 0, 0}))
                }
                break
            }
            r -= dropItem.probability
        }
    }

    return blockType.destructable
}

func (blockType *BlockType) IsSolid() bool {
    return blockType.isSolid
}

func (blockType *BlockType) GetName() string {
    return blockType.name
}

func (blockType *BlockType) GetTransparency() int8 {
    return blockType.transparency
}

func LoadStandardBlockTypes() map[BlockId]IBlockType {
    b := make(map[BlockId]*BlockType)

    newBlock := func(id BlockId, name string) {
        b[id] = &BlockType{
            name:         name,
            transparency: -1,
            destructable: true,
            isSolid:      true,
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
    newBlock(BlockIdTnt, "Tnt")
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

    setTrans := func(transparency int8, blockTypes []BlockId) {
        for _, blockType := range blockTypes {
            b[blockType].transparency = transparency
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
        b[blockId].isSolid = false
    }

    // Setup behaviour of blocks when destroyed
    setMinedDropsSameItem := func(blockTypes []BlockId) {
        for _, blockType := range blockTypes {
            b[blockType].droppedItems = append(
                b[blockType].droppedItems,
                BlockDropItem{
                    ItemId(blockType),
                    100,
                    1,
                })
        }
    }

    type Drop struct {
        minedBlockType  BlockId
        droppedItemType ItemId
    }
    setMinedDropBlock := func(drops []Drop) {
        for _, drop := range drops {
            b[drop.minedBlockType].droppedItems = append(
                b[drop.minedBlockType].droppedItems,
                BlockDropItem{
                    drop.droppedItemType,
                    100,
                    1,
                })
        }
    }

    b[BlockIdBedrock].destructable = false

    // TODO crops are more complicated, and need code to look at their metadata
    // to decide what to drop.
    // TODO ice blocks are more complicated as to what they do when destroyed

    // TODO data about tool usage

    // TODO what item ID drops for redstone torches (on vs off state)

    // Blocks that drop the same ItemId as BlockId 100% of the time
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
        Drop{BlockIdStone, ItemId(BlockIdCobblestone)},
        Drop{BlockIdGrass, ItemId(BlockIdDirt)},
        Drop{BlockIdCoalOre, cmitem.ItemIdCoal},
        Drop{BlockIdDoubleStoneSlab, ItemId(BlockIdStoneSlab)},
        Drop{BlockIdDiamondOre, cmitem.ItemIdDiamond},
        Drop{BlockIdFarmland, ItemId(BlockIdDirt)},
        Drop{BlockIdSignPost, cmitem.ItemIdSign},
        Drop{BlockIdWoodenDoor, cmitem.ItemIdWoodendoor},
        Drop{BlockIdWallSign, cmitem.ItemIdSign},
        Drop{BlockIdIronDoor, cmitem.ItemIdIronDoor},
        Drop{BlockIdSnow, ItemId(BlockIdDirt)},
        Drop{BlockIdSugarCane, cmitem.ItemIdSugarCane},
        Drop{BlockIdGlowstone, cmitem.ItemIdGlowstoneDust},
    })
    // Blocks that drop things with varying probability (or one of several
    // items)
    b[BlockIdGravel].droppedItems = []BlockDropItem{
        BlockDropItem{cmitem.ItemIdFlint, 10, 1},
        BlockDropItem{ItemId(BlockIdGravel), 90, 1},
    }
    b[BlockIdLeaves].droppedItems = []BlockDropItem{
        // TODO get more accurate probability of sapling drop
        BlockDropItem{ItemId(BlockIdSapling), 5, 1},
    }
    b[BlockIdRedstoneOre].droppedItems = []BlockDropItem{
        // TODO find probabilities of dropping 4 vs 5 items
        BlockDropItem{cmitem.ItemIdRedstone, 50, 4},
        BlockDropItem{cmitem.ItemIdRedstone, 50, 5},
    }
    b[BlockIdGlowingRedstoneOre].droppedItems = b[BlockIdRedstoneOre].droppedItems
    b[BlockIdSnowBlock].droppedItems = []BlockDropItem{
        BlockDropItem{cmitem.ItemIdSnowball, 100, 4},
    }

    retval := make(map[BlockId]IBlockType)
    for k, block := range b {
        retval[k] = block
    }
    return retval
}

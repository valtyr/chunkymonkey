package block

import (
    .      "chunkymonkey/interfaces"
    cmitem "chunkymonkey/item"
    .      "chunkymonkey/types"
)

const (
    // TODO add in new blocks from Beta 1.2
    BlockIDAir                 = BlockID(0)
    BlockIDStone               = BlockID(1)
    BlockIDGrass               = BlockID(2)
    BlockIDDirt                = BlockID(3)
    BlockIDCobblestone         = BlockID(4)
    BlockIDPlank               = BlockID(5)
    BlockIDSapling             = BlockID(6)
    BlockIDBedrock             = BlockID(7)
    BlockIDWater               = BlockID(8)
    BlockIDStationaryWater     = BlockID(9)
    BlockIDLava                = BlockID(10)
    BlockIDStationaryLava      = BlockID(11)
    BlockIDSand                = BlockID(12)
    BlockIDGravel              = BlockID(13)
    BlockIDGoldOre             = BlockID(14)
    BlockIDIronOre             = BlockID(15)
    BlockIDCoalOre             = BlockID(16)
    BlockIDLog                 = BlockID(17)
    BlockIDLeaves              = BlockID(18)
    BlockIDSponge              = BlockID(19)
    BlockIDGlass               = BlockID(20)
    BlockIDWool                = BlockID(35)
    BlockIDYellowFlower        = BlockID(37)
    BlockIDRedRose             = BlockID(38)
    BlockIDBrownMushroom       = BlockID(39)
    BlockIDRedMushroom         = BlockID(40)
    BlockIDGoldBlock           = BlockID(41)
    BlockIDIronBlock           = BlockID(42)
    BlockIDDoubleStoneSlab     = BlockID(43)
    BlockIDStoneSlab           = BlockID(44)
    BlockIDBrick               = BlockID(45)
    BlockIDTNT                 = BlockID(46)
    BlockIDBookshelf           = BlockID(47)
    BlockIDMossStone           = BlockID(48)
    BlockIDObsidian            = BlockID(49)
    BlockIDTorch               = BlockID(50)
    BlockIDFire                = BlockID(51)
    BlockIDMobSpawner          = BlockID(52)
    BlockIDWoodenStairs        = BlockID(53)
    BlockIDChest               = BlockID(54)
    BlockIDRedstoneWire        = BlockID(55)
    BlockIDDiamondOre          = BlockID(56)
    BlockIDDiamondBlock        = BlockID(57)
    BlockIDWorkbench           = BlockID(58)
    BlockIDCrops               = BlockID(59)
    BlockIDFarmland            = BlockID(60)
    BlockIDFurnace             = BlockID(61)
    BlockIDBurningFurnace      = BlockID(62)
    BlockIDSignPost            = BlockID(63)
    BlockIDWoodenDoor          = BlockID(64)
    BlockIDLadder              = BlockID(65)
    BlockIDMinecartTracks      = BlockID(66)
    BlockIDCobblestoneStairs   = BlockID(67)
    BlockIDWallSign            = BlockID(68)
    BlockIDLever               = BlockID(69)
    BlockIDStonePressurePlate  = BlockID(70)
    BlockIDIronDoor            = BlockID(71)
    BlockIDWoodenPressurePlate = BlockID(72)
    BlockIDRedstoneOre         = BlockID(73)
    BlockIDGlowingRedstoneOre  = BlockID(74)
    BlockIDRedstoneTorchOff    = BlockID(75)
    BlockIDRedstoneTorchOn     = BlockID(76)
    BlockIDStoneButton         = BlockID(77)
    BlockIDSnow                = BlockID(78)
    BlockIDIce                 = BlockID(79)
    BlockIDSnowBlock           = BlockID(80)
    BlockIDCactus              = BlockID(81)
    BlockIDClay                = BlockID(82)
    BlockIDSugarCane           = BlockID(83)
    BlockIDJukebox             = BlockID(84)
    BlockIDFence               = BlockID(85)
    BlockIDPumpkin             = BlockID(86)
    BlockIDNetherrack          = BlockID(87)
    BlockIDSoulSand            = BlockID(88)
    BlockIDGlowstone           = BlockID(89)
    BlockIDPortal              = BlockID(90)
    BlockIDJackOLantern        = BlockID(91)
)

type BlockDropItem struct {
    droppedItem ItemID
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
func (blockType *BlockType) Destroy(chunk IChunk, blockLoc *BlockXYZ) bool {
    if len(blockType.droppedItems) > 0 {
        rand := chunk.GetRand()
        // Possibly drop item(s)
        r := byte(rand.Intn(100))
        for _, dropItem := range blockType.droppedItems {
            if dropItem.probability > r {
                for i := dropItem.quantity; i > 0; i-- {
                    position := blockLoc.ToAbsXYZ()
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

func LoadStandardBlockTypes() map[BlockID]IBlockType {
    b := make(map[BlockID]*BlockType)

    newBlock := func(id BlockID, name string) {
        b[id] = &BlockType{
            name:         name,
            transparency: -1,
            destructable: true,
            isSolid:      true,
        }
    }

    newBlock(BlockIDAir, "air")
    newBlock(BlockIDStone, "stone")
    newBlock(BlockIDGrass, "grass")
    newBlock(BlockIDDirt, "dirt")
    newBlock(BlockIDCobblestone, "cobblestone")
    newBlock(BlockIDPlank, "wood")
    newBlock(BlockIDSapling, "sapling")
    newBlock(BlockIDBedrock, "bedrock")
    newBlock(BlockIDWater, "water")
    newBlock(BlockIDStationaryWater, "stationary water")
    newBlock(BlockIDLava, "lava")
    newBlock(BlockIDStationaryLava, "stationary lava")
    newBlock(BlockIDSand, "sand")
    newBlock(BlockIDGravel, "gravel")
    newBlock(BlockIDGoldOre, "gold ore")
    newBlock(BlockIDIronOre, "iron ore")
    newBlock(BlockIDCoalOre, "coal ore")
    newBlock(BlockIDLog, "log")
    newBlock(BlockIDLeaves, "leaves")
    newBlock(BlockIDSponge, "sponge")
    newBlock(BlockIDGlass, "glass")
    newBlock(BlockIDWool, "wool")
    newBlock(BlockIDYellowFlower, "yellow flower")
    newBlock(BlockIDRedRose, "red rose")
    newBlock(BlockIDBrownMushroom, "brown mushroom")
    newBlock(BlockIDRedMushroom, "red mushroom")
    newBlock(BlockIDGoldBlock, "gold block")
    newBlock(BlockIDIronBlock, "iron block")
    newBlock(BlockIDDoubleStoneSlab, "double stone slab")
    newBlock(BlockIDStoneSlab, "stone slab")
    newBlock(BlockIDBrick, "brick")
    newBlock(BlockIDTNT, "TNT")
    newBlock(BlockIDBookshelf, "bookshelf")
    newBlock(BlockIDMossStone, "moss stone")
    newBlock(BlockIDObsidian, "obsidian")
    newBlock(BlockIDTorch, "torch")
    newBlock(BlockIDFire, "fire")
    newBlock(BlockIDMobSpawner, "mob spawner")
    newBlock(BlockIDWoodenStairs, "wooden stairs")
    newBlock(BlockIDChest, "chest")
    newBlock(BlockIDRedstoneWire, "redstone wire")
    newBlock(BlockIDDiamondOre, "diamond ore")
    newBlock(BlockIDDiamondBlock, "diamond block")
    newBlock(BlockIDWorkbench, "workbench")
    newBlock(BlockIDCrops, "crops")
    newBlock(BlockIDFarmland, "soil")
    newBlock(BlockIDFurnace, "furnace")
    newBlock(BlockIDBurningFurnace, "burning furnace")
    newBlock(BlockIDSignPost, "sign post")
    newBlock(BlockIDWoodenDoor, "wooden door")
    newBlock(BlockIDLadder, "ladder")
    newBlock(BlockIDMinecartTracks, "minecart tracks")
    newBlock(BlockIDCobblestoneStairs, "cobblestone stairs")
    newBlock(BlockIDWallSign, "wall sign")
    newBlock(BlockIDLever, "lever")
    newBlock(BlockIDStonePressurePlate, "stone pressure plate")
    newBlock(BlockIDIronDoor, "irondoor")
    newBlock(BlockIDWoodenPressurePlate, "wooden pressure plate")
    newBlock(BlockIDRedstoneOre, "redstone ore")
    newBlock(BlockIDGlowingRedstoneOre, "glowing redstone ore")
    newBlock(BlockIDRedstoneTorchOff, "redstone torch off")
    newBlock(BlockIDRedstoneTorchOn, "redstone torch on")
    newBlock(BlockIDStoneButton, "stone button")
    newBlock(BlockIDSnow, "snow")
    newBlock(BlockIDIce, "ice")
    newBlock(BlockIDSnowBlock, "snow block")
    newBlock(BlockIDCactus, "cactus")
    newBlock(BlockIDClay, "clay")
    newBlock(BlockIDSugarCane, "sugar cane")
    newBlock(BlockIDJukebox, "jukebox")
    newBlock(BlockIDFence, "fence")
    newBlock(BlockIDPumpkin, "pumpkin")
    newBlock(BlockIDNetherrack, "netherrack")
    newBlock(BlockIDSoulSand, "soul sand")
    newBlock(BlockIDGlowstone, "glowstone")
    newBlock(BlockIDPortal, "portal")
    newBlock(BlockIDJackOLantern, "jack o lantern")

    setTrans := func(transparency int8, blockTypes []BlockID) {
        for _, blockType := range blockTypes {
            b[blockType].transparency = transparency
        }
    }
    // Setup transparent blocks
    setTrans(0, []BlockID{BlockIDAir, BlockIDSapling, BlockIDGlass,
        BlockIDYellowFlower, BlockIDRedRose, BlockIDBrownMushroom,
        BlockIDRedMushroom, BlockIDFire, BlockIDMobSpawner, BlockIDWoodenStairs,
        BlockIDRedstoneWire, BlockIDCrops, BlockIDSignPost, BlockIDLadder,
        BlockIDMinecartTracks, BlockIDCobblestoneStairs, BlockIDWallSign,
        BlockIDLever, BlockIDIronDoor, BlockIDRedstoneTorchOff,
        BlockIDRedstoneTorchOn, BlockIDStoneButton, BlockIDSnow, BlockIDCactus,
        BlockIDSugarCane, BlockIDFence, BlockIDPortal})

    // Setup semi-transparent blocks
    setTrans(1, []BlockID{BlockIDLeaves})
    setTrans(3, []BlockID{BlockIDWater, BlockIDStationaryWater, BlockIDIce})

    // Setup non-solid blocks
    nonSolid := []BlockID{
        BlockIDAir, BlockIDSapling, BlockIDWater, BlockIDStationaryWater,
        BlockIDLava, BlockIDStationaryLava, BlockIDYellowFlower,
        BlockIDRedRose, BlockIDBrownMushroom, BlockIDRedMushroom, BlockIDTorch,
        BlockIDFire, BlockIDRedstoneWire, BlockIDCrops, BlockIDSignPost,
        BlockIDLadder, BlockIDMinecartTracks, BlockIDWallSign, BlockIDLever,
        BlockIDStonePressurePlate, BlockIDIronDoor, BlockIDWoodenPressurePlate,
        BlockIDRedstoneTorchOff, BlockIDRedstoneTorchOn, BlockIDStoneButton,
        BlockIDSugarCane, BlockIDPortal,
    }
    for _, blockID := range nonSolid {
        b[blockID].isSolid = false
    }

    // Setup behaviour of blocks when destroyed
    setMinedDropsSameItem := func(blockTypes []BlockID) {
        for _, blockType := range blockTypes {
            b[blockType].droppedItems = append(
                b[blockType].droppedItems,
                BlockDropItem{
                    ItemID(blockType),
                    100,
                    1,
                })
        }
    }

    type Drop struct {
        minedBlockType  BlockID
        droppedItemType ItemID
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

    b[BlockIDBedrock].destructable = false

    // TODO crops are more complicated, and need code to look at their metadata
    // to decide what to drop.
    // TODO ice blocks are more complicated as to what they do when destroyed

    // TODO data about tool usage

    // TODO what item ID drops for redstone torches (on vs off state)

    // Blocks that drop the same ItemID as BlockID 100% of the time
    setMinedDropsSameItem([]BlockID{
        BlockIDDirt, BlockIDCobblestone, BlockIDPlank, BlockIDSapling,
        BlockIDSand, BlockIDGoldOre, BlockIDIronOre, BlockIDLog, BlockIDSponge,
        BlockIDWool, BlockIDYellowFlower, BlockIDRedRose, BlockIDBrownMushroom,
        BlockIDRedMushroom, BlockIDGoldBlock, BlockIDIronBlock,
        BlockIDStoneSlab, BlockIDBrick, BlockIDMossStone, BlockIDObsidian,
        BlockIDTorch, BlockIDWoodenStairs, BlockIDChest, BlockIDDiamondBlock,
        BlockIDWorkbench, BlockIDLadder, BlockIDMinecartTracks,
        BlockIDCobblestoneStairs, BlockIDLever, BlockIDStonePressurePlate,
        BlockIDWoodenPressurePlate, BlockIDStoneButton, BlockIDCactus,
        BlockIDClay, BlockIDJukebox, BlockIDFence, BlockIDPumpkin,
        BlockIDNetherrack, BlockIDSoulSand, BlockIDGlowstone,
        BlockIDJackOLantern,
    })
    // Blocks that drop a single different item 100% of the time
    setMinedDropBlock([]Drop{
        Drop{BlockIDStone, ItemID(BlockIDCobblestone)},
        Drop{BlockIDGrass, ItemID(BlockIDDirt)},
        Drop{BlockIDCoalOre, cmitem.ItemIDCoal},
        Drop{BlockIDDoubleStoneSlab, ItemID(BlockIDStoneSlab)},
        Drop{BlockIDDiamondOre, cmitem.ItemIDDiamond},
        Drop{BlockIDFarmland, ItemID(BlockIDDirt)},
        Drop{BlockIDSignPost, cmitem.ItemIDSign},
        Drop{BlockIDWoodenDoor, cmitem.ItemIDWoodendoor},
        Drop{BlockIDWallSign, cmitem.ItemIDSign},
        Drop{BlockIDIronDoor, cmitem.ItemIDIronDoor},
        Drop{BlockIDSnow, ItemID(BlockIDDirt)},
        Drop{BlockIDSugarCane, cmitem.ItemIDSugarCane},
        Drop{BlockIDGlowstone, cmitem.ItemIDGlowstoneDust},
    })
    // Blocks that drop things with varying probability (or one of several
    // items)
    b[BlockIDGravel].droppedItems = []BlockDropItem{
        BlockDropItem{cmitem.ItemIDFlint, 10, 1},
        BlockDropItem{ItemID(BlockIDGravel), 90, 1},
    }
    b[BlockIDLeaves].droppedItems = []BlockDropItem{
        // TODO get more accurate probability of sapling drop
        BlockDropItem{ItemID(BlockIDSapling), 5, 1},
    }
    b[BlockIDRedstoneOre].droppedItems = []BlockDropItem{
        // TODO find probabilities of dropping 4 vs 5 items
        BlockDropItem{cmitem.ItemIDRedstone, 50, 4},
        BlockDropItem{cmitem.ItemIDRedstone, 50, 5},
    }
    b[BlockIDGlowingRedstoneOre].droppedItems = b[BlockIDRedstoneOre].droppedItems
    b[BlockIDSnowBlock].droppedItems = []BlockDropItem{
        BlockDropItem{cmitem.ItemIDSnowball, 100, 4},
    }

    retval := make(map[BlockID]IBlockType)
    for k, block := range b {
        retval[k] = block
    }
    return retval
}

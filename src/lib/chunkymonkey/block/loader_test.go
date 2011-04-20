package block

import (
    "reflect"
    "strings"
    "testing"

    . "chunkymonkey/types"
)

const twoBlocks = ("{\n" +
    "  \"0\": {\n" +
    "    \"Aspect\": \"Standard\",\n" +
    "    \"AspectArgs\": {\n" +
    "      \"DroppedItems\": [],\n" +
    "      \"BreakOn\": 2\n" +
    "    },\n" +
    "    \"Name\": \"air\",\n" +
    "    \"Opacity\": 0,\n" +
    "    \"Destructable\": true,\n" +
    "    \"Solid\": false,\n" +
    "    \"Replaceable\": true,\n" +
    "    \"Attachable\": false\n" +
    "  },\n" +
    "  \"1\": {\n" +
    "    \"Aspect\": \"Standard\",\n" +
    "    \"AspectArgs\": {\n" +
    "      \"DroppedItems\": [\n" +
    "        {\n" +
    "          \"DroppedItem\": 4,\n" +
    "          \"Probability\": 100,\n" +
    "          \"Count\": 1\n" +
    "        }\n" +
    "      ],\n" +
    "      \"BreakOn\": 2\n" +
    "    },\n" +
    "    \"Name\": \"stone\",\n" +
    "    \"Opacity\": 15,\n" +
    "    \"Destructable\": true,\n" +
    "    \"Solid\": true,\n" +
    "    \"Replaceable\": false,\n" +
    "    \"Attachable\": true\n" +
    "  }\n" +
    "}")

func assertBlockTypeEq(t *testing.T, expected, result *BlockType) {
    if !reflect.DeepEqual(expected, result) {
        t.Error("BlockTypes differed")
        t.Logf("    expected %#v", expected)
        t.Logf("    result   %#v", result)
        t.Logf("    expected aspect %#v", expected.Aspect)
        t.Logf("    result aspect   %#v", result.Aspect)
    }
}


func TestLoadBlockDefs(t *testing.T) {
    reader := strings.NewReader(twoBlocks)
    t.Log(twoBlocks)

    blocks, err := LoadBlockDefs(reader)
    if err != nil {
        t.Fatalf("Expected no error but got %v", err)
    }
    if len(blocks) != 2 {
        t.Fatalf("Expected 2 block defs, but got %d", len(blocks))
    }

    assertBlockTypeEq(
        t,
        &BlockType{
            BlockAttrs: BlockAttrs{
                Name: "air",
                Opacity: 0,
                Destructable: true,
                Solid: false,
                Replaceable: true,
                Attachable: false,
            },
            Aspect: &StandardAspect{
                // TODO air will likely change to some specialised block in
                // future.
                DroppedItems: []BlockDropItem{},
                BreakOn: DigBlockBroke,
            },
        },
        blocks[0],
    )

    assertBlockTypeEq(
        t,
        &BlockType{
            BlockAttrs: BlockAttrs{
                Name: "stone",
                Opacity: 15,
                Destructable: true,
                Solid: true,
                Replaceable: false,
                Attachable: true,
            },
            Aspect: &StandardAspect{
                DroppedItems: []BlockDropItem{
                    BlockDropItem{
                        DroppedItem: 4,
                        Probability: 100,
                        Count: 1,
                    },
                },
                BreakOn: DigBlockBroke,
            },
        },
        blocks[1],
    )
}

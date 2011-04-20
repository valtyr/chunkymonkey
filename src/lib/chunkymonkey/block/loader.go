package block

import (
    "fmt"
    "io"
    "json"
    "os"

    . "chunkymonkey/types"
)

type aspectMakerFn func() (aspect IBlockAspect)

var aspectMakers map[string]aspectMakerFn

// Used specifically for json unmarshalling of block definitions.
type blockDef struct {
    BlockAttrs
    Aspect string
    AspectArgs *aspectArgs
}

type aspectArgs struct {
    Raw []byte
}

func (a *aspectArgs)UnmarshalJSON(raw []byte) (err os.Error) {
    // Copy raw into a.Raw - the JSON library will destroy the content of the
    // argument after this function returns.
    a.Raw = make([]byte, len(raw))
    for i := range raw {
        a.Raw[i] = raw[i]
    }
    return
}

func LoadBlockDefs(reader io.Reader) (blocks map[BlockId]*BlockType, err os.Error) {
    blocksStr := make(map[string]blockDef)
    decoder := json.NewDecoder(reader)
    err = decoder.Decode(&blocksStr)

    // Convert map string keys to ints.
    blocks = make(map[BlockId]*BlockType)
    for idStr, blockDef := range blocksStr {
        var id BlockId
        fmt.Sscanf(idStr, "%d", &id)
        block := &BlockType{
            BlockAttrs: blockDef.BlockAttrs,
        }

        // Create the Aspect attribute of the block.
        aspectMakerFn, ok := aspectMakers[blockDef.Aspect]
        if !ok {
            err = os.NewError(fmt.Sprintf("Unknown aspect type %q", blockDef.Aspect))
            return
        }
        aspect := aspectMakerFn()
        err = json.Unmarshal(blockDef.AspectArgs.Raw, aspect)
        if err != nil {
            return
        }
        block.Aspect = aspect

        blocks[id] = block
    }

    return
}

func init() {
    aspectMakers = map[string]aspectMakerFn{
        "Standard": makeStandardAspect,
    }
}

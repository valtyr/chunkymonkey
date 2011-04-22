package itemtype

import (
    "io"
    "json"
    "os"
    "strconv"

    . "chunkymonkey/types"
)

type ToolTypeId byte

type ItemType struct {
    Name     string
    MaxStack ItemCount
    ToolType ToolTypeId
    ToolUses ItemData
}

type ItemTypeMap map[ItemTypeId]ItemType

func LoadItemDefs(reader io.Reader) (items ItemTypeMap, err os.Error) {
    itemsStr := make(map[string]ItemType)
    decoder := json.NewDecoder(reader)
    err = decoder.Decode(&itemsStr)

    // Convert map string keys to ints.
    items = make(ItemTypeMap)
    for idStr := range itemsStr {
        var id int
        id, err = strconv.Atoi(idStr)

        if err != nil {
            return
        }

        items[ItemTypeId(id)] = itemsStr[idStr]
    }

    return
}

package main

import (
    "log"
    "os"

    "chunkymonkey/block"
)

func dumpBlocks(filename string) (err os.Error) {
    outputFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0640)
    if err != nil {
        return
    }
    defer outputFile.Close()

    blocks := block.LoadStandardBlockTypes()
    err = block.SaveBlockDefs(outputFile, blocks)

    return
}

func main() {
    err := dumpBlocks("blocks.json")
    if err != nil {
        log.Printf("Error: %v", err)
    }
}

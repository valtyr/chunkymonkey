package main

import (
    _ "expvar"
    "flag"
    "http"
    "log"
    "net"
    "os"

    "chunkymonkey"
    "chunkymonkey/block"
)

var addr = flag.String(
    "addr", ":25565",
    "Serves on the given address:port.")

var httpAddr = flag.String(
    "http_addr", ":25566",
    "Serves HTTP diagnostics on the given address:port.")

var blockDefs = flag.String(
    "blocks", "blocks.json",
    "The JSON file containing block type definitions.")

func usage() {
    os.Stderr.WriteString("usage: " + os.Args[0] + " [flags] <world>\n")
    flag.PrintDefaults()
}

func startHttpServer(addr string) (err os.Error) {
    httpPort, err := net.Listen("tcp", addr)
    if err != nil {
        return
    }
    go http.Serve(httpPort, nil)
    return
}

func loadBlocks() (blocks block.BlockTypeMap, err os.Error) {
    file, err := os.Open(*blockDefs)
    if err != nil {
        return
    }
    defer file.Close()

    return block.LoadBlockDefs(file)
}

func main() {
    var err os.Error

    flag.Usage = usage
    flag.Parse()

    if flag.NArg() != 1 {
        flag.Usage()
        os.Exit(1)
    }

    blocks, err := loadBlocks()
    if err != nil {
        log.Print("Error loading block definitions: ", err)
        os.Exit(1)
    }

    worldPath := flag.Arg(0)
    game, err := chunkymonkey.NewGame(worldPath, blocks)
    if err != nil {
        log.Panic(err)
    }

    err = startHttpServer(*httpAddr)
    if err != nil {
        log.Panic(err)
    }

    game.Serve(*addr)
}

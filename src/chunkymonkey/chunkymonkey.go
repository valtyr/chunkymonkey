package main

import (
    "chunkymonkey/chunkymonkey"
    "flag"
    "os"
)

var addr = flag.String("addr", ":25565", "Serves on the given address:port.")

func usage() {
    os.Stderr.WriteString("usage: " + os.Args[0] + " [flags] <world>\n")
    flag.PrintDefaults()
}

func main() {
    flag.Usage = usage
    flag.Parse()

    if flag.NArg() != 1 {
        flag.Usage()
        os.Exit(1)
    }

    worldPath := flag.Arg(0)

    game := chunkymonkey.NewGame(worldPath)
    game.Serve(*addr)
}

package main

import (
    "chunkymonkey/chunkymonkey"
    "flag"
    "os"
)

func usage() {
    os.Stderr.WriteString("usage: " + os.Args[0] + " <world>\n")
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
    game.Serve(":25565")
}

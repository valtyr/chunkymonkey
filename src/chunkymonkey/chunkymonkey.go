package main

import (
    _   "expvar"
    "flag"
    "http"
    "log"
    "net"
    "os"

    "chunkymonkey/chunkymonkey"
)

var addr = flag.String(
    "addr", ":25565",
    "Serves on the given address:port.")

var httpAddr = flag.String(
    "http_addr", ":25566",
    "Serves HTTP diagnostics on the given address:port.")

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

func main() {
    var err os.Error

    flag.Usage = usage
    flag.Parse()

    if flag.NArg() != 1 {
        flag.Usage()
        os.Exit(1)
    }

    worldPath := flag.Arg(0)
    game := chunkymonkey.NewGame(worldPath)

    err = startHttpServer(*httpAddr)
    if err != nil {
        log.Panic(err)
    }

    game.Serve(*addr)
}

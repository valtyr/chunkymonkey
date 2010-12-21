package main

import (
    "flag"
    "io"
    "log"
    "net"
    "os"
    parser "intercept_parse"
)

type RelayReport struct {
    written int64
    err     os.Error
}

// Sends a report on reportChan when it completes
func spliceParser(reportChan chan RelayReport,
parser parser.Parser,
dst io.Writer, src io.Reader) {

    parserReader, parserWriter := io.Pipe()
    wrappedDst := io.MultiWriter(dst, parserWriter)

    go parser.Parse(parserReader)

    written, err := io.Copy(wrappedDst, src)

    reportChan <- RelayReport{written, err}
}

func serveConn(clientConn net.Conn, remoteaddr string) {
    defer clientConn.Close()

    clientAddr := clientConn.RemoteAddr().String()

    log.Printf("(%s) Client connected ", clientAddr)

    log.Printf("(%s) Creating relay to server %s", clientAddr, remoteaddr)
    serverConn, err := net.Dial("tcp", "", remoteaddr)

    if err != nil {
        log.Printf("(%s) Error connecting to %s: %v", clientAddr, remoteaddr, err)
        return
    }

    defer serverConn.Close()

    log.Printf("(%s) Connected to server %s", clientAddr, remoteaddr)

    serverToClientReportChan := make(chan RelayReport)
    clientToServerReportChan := make(chan RelayReport)

    // Set up for parsing messages from server to client
    go spliceParser(serverToClientReportChan, new(parser.ServerMessageParser), clientConn, serverConn)
    // Set up for parsing messages from client to server
    go spliceParser(clientToServerReportChan, new(parser.ClientMessageParser), serverConn, clientConn)

    // Wait for the both relay/splices to stop, then we let the connections
    // close via deferred calls
    report := <-serverToClientReportChan
    log.Printf("(%s) Server->client relay after %d bytes with error: %v", clientAddr, report.written, report.err)
    report = <-clientToServerReportChan
    log.Printf("(%s) Client->server relay after %d bytes with error: %v", clientAddr, report.written, report.err)

    log.Printf("(%s) Client disconnected", clientAddr)
}

func serve(localaddr, remoteaddr string) (err os.Error) {
    listener, err := net.Listen("tcp", localaddr)
    if err != nil {
        log.Exit("Listen: ", err.String())
        return
    }

    defer listener.Close()
    log.Print("Listening on ", localaddr)

    for {
        clientConn, acceptErr := listener.Accept()
        if acceptErr != nil {
            log.Printf("Accept error: %s", acceptErr.String())
        } else {
            go serveConn(clientConn, remoteaddr)
        }
    }
    return
}

func usage() {
    os.Stderr.WriteString("usage: " + os.Args[0] + " localaddr:port remoteaddr:port\n")
    flag.PrintDefaults()
}

func main() {
    flag.Usage = usage
    flag.Parse()

    if flag.NArg() != 2 {
        flag.Usage()
        os.Exit(1)
    }

    localaddr := flag.Arg(0)
    remoteaddr := flag.Arg(1)

    serve(localaddr, remoteaddr)
}

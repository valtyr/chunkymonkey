package intercept_parse

import (
    "encoding/hex"
    "io"
    "log"
)

type Parser interface {
    Parse(reader io.Reader)
}

// Hex dumps the input to the log
func dumpInput(logPrefix string, reader io.Reader) {
    buf := make([]byte, 16, 16)
    for {
        _, err := io.ReadAtLeast(reader, buf, 1)
        if err != nil {
            return
        }

        hexData := hex.EncodeToString(buf)
        log.Printf("%sUnparsed data: %s", logPrefix, hexData)
    }
}

// Consumes data from reader until an error occurs
func consumeUnrecognizedInput(reader io.Reader) {
    buf := make([]byte, 4096)
    for {
        _, err := io.ReadAtLeast(reader, buf, 1)
        if err != nil {
            return
        }
    }
}

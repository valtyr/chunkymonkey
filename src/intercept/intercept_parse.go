package intercept_parse

import (
    "encoding/hex"
    "io"
    "log"
)

type Parser interface {
    Parse(reader io.Reader)
}

func hexDump(data []byte) {
    hexData := hex.EncodeToString(data)
    log.Printf("Unparsed data: %s", hexData)
}

// Hex dumps the input to the log
func dumpInput(reader io.Reader) {
    buf := make([]byte, 16, 16)
    for {
        _, err := io.ReadAtLeast(reader, buf, 1)
        if err != nil {
            return
        }

        hexDump(buf)
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

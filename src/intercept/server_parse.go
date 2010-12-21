package intercept_parse

import (
    "io"
)

type ServerMessageParser struct {}

// Parses messages from the server
func (p *ServerMessageParser)Parse(reader io.Reader) {
    dumpInput("(S->C) ", reader)
}

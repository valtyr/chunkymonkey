package serverAuth

import (
    "http"
    "os"
)

func CheckUserAuth(serverId, user string) (authenticated bool, err os.Error) {
    authenticated = false

    url := "http://www.minecraft.net/game/checkserver.jsp?" + http.EncodeQuery(
        map[string][]string {
            "serverId": {serverId},
            "user": {user},
        },
    )

    response, _, err := http.Get(url)
    if err != nil {
        return
    }

    if response.StatusCode == 200 {
        // We only need to read up to 3 bytes for "YES" or "NO"
        buf := make([]byte, 3)
        bufferPos := 0
        var numBytesRead int

        for err == nil && bufferPos < 3 {
            numBytesRead, err = response.Body.Read(buf[bufferPos:])
            if err != nil && err != os.EOF{
                return
            }
            bufferPos += numBytesRead
        }

        result := string(buf[0:bufferPos])
        authenticated = (result == "YES")
    } else {
        return
    }

    return
}

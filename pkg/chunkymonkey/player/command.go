package player

import (
	"flag"
	"os"
	"bytes"
	"runtime/pprof"
	"strconv"
	"sync"
	"strings"

	"chunkymonkey/proto"
	"chunkymonkey/slot"
	. "chunkymonkey/types"
)

var profileCmdsEnabled = flag.Bool("profile_cmds", false, "Enable profiling commands")

var profiling bool
var profilingMutex sync.Mutex

const cpuprofileCmd = "cpuprofile"
const cpuprofileUsage = ""
const cpuprofileDesc = ""

func (player *Player) cmdCpuProfile(message string) {
	if !*profileCmdsEnabled {
		return
	}

	profilingMutex.Lock()
	defer profilingMutex.Unlock()

	filename := "/tmp/chunkymonkey.cpu.pprof"

	if !profiling {
		w, err := os.Create(filename)
		if err != nil {
			buf := new(bytes.Buffer)
			proto.WriteChatMessage(buf, err.String())
			player.TransmitPacket(buf.Bytes())
			return
		}
		pprof.StartCPUProfile(w)
		profiling = true
		buf := new(bytes.Buffer)
		proto.WriteChatMessage(buf, "CPU profiling started and writing to "+filename)
		player.TransmitPacket(buf.Bytes())
	} else {
		pprof.StopCPUProfile()
		profiling = false
		buf := new(bytes.Buffer)
		proto.WriteChatMessage(buf, "CPU profiling stopped")
		player.TransmitPacket(buf.Bytes())
	}
}

const memprofileCmd = "memprofile"
const memprofileUsage = ""
const memprofileDesc = ""

func (player *Player) cmdMemProfile(message string) {
	if !*profileCmdsEnabled {
		return
	}

	filename := "/tmp/chunkymonkey.heap.pprof"

	w, err := os.Create(filename)
	if err != nil {
		buf := new(bytes.Buffer)
		proto.WriteChatMessage(buf, err.String())
		player.TransmitPacket(buf.Bytes())
		return
	}
	defer w.Close()

	pprof.WriteHeapProfile(w)

	buf := new(bytes.Buffer)
	proto.WriteChatMessage(buf, "Heap profile written to "+filename)
	player.TransmitPacket(buf.Bytes())
}

const giveCmd = "give"
const giveUsage = "/give <item ID> [<quantity> [<data>]]"
const giveDesc = "Gives x amount of y items to player."

func (player *Player) cmdGive(message string) {
	cmdParts := strings.Split(message, " ", -1)
	if len(cmdParts) < 2 || len(cmdParts) > 4 {
		buf := new(bytes.Buffer)
		proto.WriteChatMessage(buf, giveUsage)
		player.TransmitPacket(buf.Bytes())
		return
	}
	cmdParts = cmdParts[1:]

	// TODO First argument should be player to receive item. Right now it just
	// gives it to the current player.
	itemId, err := strconv.Atoi(cmdParts[0])
	if err != nil {
		buf := new(bytes.Buffer)
		proto.WriteChatMessage(buf, giveUsage)
		player.TransmitPacket(buf.Bytes())
		return
	}

	quantity := 1
	if len(cmdParts) >= 2 {
		quantity, err = strconv.Atoi(cmdParts[1])
		if err != nil {
			buf := new(bytes.Buffer)
			proto.WriteChatMessage(buf, giveUsage)
			player.TransmitPacket(buf.Bytes())
			return
		}
	}

	data := 0
	if len(cmdParts) >= 3 {
		data, err = strconv.Atoi(cmdParts[2])
		if err != nil {
			buf := new(bytes.Buffer)
			proto.WriteChatMessage(buf, giveUsage)
			player.TransmitPacket(buf.Bytes())
			return
		}
	}

	itemType, ok := player.gameRules.ItemTypes[ItemTypeId(itemId)]
	if !ok {
		buf := new(bytes.Buffer)
		proto.WriteChatMessage(buf, "Unknown item ID")
		player.TransmitPacket(buf.Bytes())
		return
	}

	item := slot.Slot{
		ItemType: itemType,
		Count:    ItemCount(quantity),
		Data:     ItemData(data),
	}

	player.reqGiveItem(&player.position, &item)
}

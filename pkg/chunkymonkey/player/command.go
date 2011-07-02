package player

import (
	"bytes"
	"strconv"
	"strings"

	"chunkymonkey/proto"
	"chunkymonkey/slot"
	. "chunkymonkey/types"
)

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

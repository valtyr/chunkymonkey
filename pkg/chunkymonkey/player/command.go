package player

import (
	"bytes"
	"strconv"
	"strings"

	"chunkymonkey/command"
	"chunkymonkey/proto"
	"chunkymonkey/slot"
	. "chunkymonkey/types"
)

// /tell player message 
const tellCmd = "tell"
const tellUsage = "tell <player> <message>"
const tellDesc = "Tells a player a message."

func (player *Player) cmdTell(message string) {
	cmdParts := strings.Split(message, " ", -1)
	if len(cmdParts) < 3 {
		buf := new(bytes.Buffer)
		proto.WriteChatMessage(buf, tellUsage)
		player.TransmitPacket(buf.Bytes())
		return
	}
	/* TODO Get player to send message, too
	player := cmdParts[1]
	message := strings.Join(cmdParts[2:], " ")
	*/
}

const helpShortCmd = "?"
const helpCmd = "help"
const helpUsage = "help|?"
const helpDesc = "Shows a list of all commands."
// TODO: Implement help <command> to show the description and usage of a command
func (player *Player) cmdHelp(message string, cmdFramework *command.CommandFramework) {
	var resp string
	cmds := cmdFramework.Commands()
	if len(cmds) == 0 {
		resp = "No commands available."
	} else {
		resp = "Commands:"
		for trigger, _ := range cmds {
			resp += " " + trigger + ","
		}
		resp = resp[:len(resp)-1]
	}
	buf := new(bytes.Buffer)
	proto.WriteChatMessage(buf, resp)
	player.TransmitPacket(buf.Bytes())
}

const giveCmd = "give"
const giveUsage = "give <item ID> [<quantity> [<data>]]"
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

	// TODO Check for item IDs which could break the client
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
	player.Enqueue(func(player *Player) {
		player.reqGiveItem(&player.position, &item)
	})
}

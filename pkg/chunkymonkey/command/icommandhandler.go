package command

import (
	"os"
)

type ICommandHandler interface {
	GiveItem(itemTypeId int, quantity int, data int) os.Error
	SendMessageToPlayer(msg string)
	BroadcastMessage(msg string, self bool)
}

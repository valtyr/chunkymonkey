package command

type ICommandHandler interface {
	GiveItem(itemTypeId int, quantity int, data int)
	SendMessageToPlayer(msg string)
	BroadcastMessage(msg string, self bool)
}

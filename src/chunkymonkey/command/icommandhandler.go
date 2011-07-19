package command

type ICommandHandler interface {
	// Give a player 'quantity' of 'itemTypeId' with data value 'data'
	GiveItem(player string, itemTypeId, quantity, data int)
	// Sends a message from the server to the player
	SendMessageToPlayer(player, msg string)
	// Send a message to all users connected to the game
	BroadcastMessage(msg string)
}

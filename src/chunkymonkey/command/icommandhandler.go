package command

// ICommandHandler contains methods that can perform actions that impact the
// game in some manner, such as giving an item to a player or sending
// messages. In addition there are several utility functions that can be used
// to validate the input to these commands.
type ICommandHandler interface {
	// Give a player 'quantity' of 'itemTypeId' with data value 'data'
	GiveItem(player string, itemTypeId, quantity, data int)
	// Sends a message from the server to the player
	SendMessageToPlayer(player, msg string)
	// Send a message to all users connected to the game
	BroadcastMessage(msg string)
	// Teleport one player to another
	TeleportToPlayer(teleportee, destination string)

	// Return whether or not a player name is valid (i.e. player is logged in)
	IsValidPlayerName(name string) bool
	// Return whether or not a given itemId is valid
	IsValidItemId(id int) bool
}

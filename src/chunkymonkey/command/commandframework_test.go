package command

import (
	"testing"

	"gomock.googlecode.com/hg/gomock"

	"testmatcher"
)

func TestCommandFramework(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockPlayer := NewMockICommandHandler(mockCtrl)

	cf := NewCommandFramework("/")

	mockPlayer.EXPECT().BroadcastMessage("§dthis is a broadcast", true)
	cf.Process("/say this is a broadcast", mockPlayer)

	mockPlayer.EXPECT().GiveItem(1, 64, 0)
	cf.Process("/give 1 64", mockPlayer)

	mockPlayer.EXPECT().SendMessageToPlayer(&testmatcher.StringPrefix{"Commands:"})
	cf.Process("/help", mockPlayer)

	gomock.InOrder(
		mockPlayer.EXPECT().SendMessageToPlayer("Command: /help"),
		mockPlayer.EXPECT().SendMessageToPlayer("Usage: help|?"),
		mockPlayer.EXPECT().SendMessageToPlayer("Description: Shows a list of all commands."),
	)
	cf.Process("/help help", mockPlayer)
}

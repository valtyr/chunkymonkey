package command

import (
	"testing"

	"gomock.googlecode.com/hg/gomock"

	"testmatcher"
)

func TestCommandFramework(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockGame := NewMockICommandHandler(mockCtrl)

	cf := NewCommandFramework("/")

	mockGame.EXPECT().BroadcastMessage("thePlayer", "§dthis is a broadcast")
	cf.Process("thePlayer", "/say this is a broadcast", mockGame)

	mockGame.EXPECT().GiveItem("thePlayer", 1, 64, 0)
	cf.Process("thePlayer", "/give 1 64", mockGame)

	mockGame.EXPECT().SendMessageToPlayer("thePlayer", &testmatcher.StringPrefix{"Commands:"})
	cf.Process("thePlayer", "/help", mockGame)

	gomock.InOrder(
		mockGame.EXPECT().SendMessageToPlayer("thePlayer", "Command: /help"),
		mockGame.EXPECT().SendMessageToPlayer("thePlayer", "Usage: help|?"),
		mockGame.EXPECT().SendMessageToPlayer("thePlayer", "Description: Shows a list of all commands."),
	)
	cf.Process("thePlayer", "/help help", mockGame)
}

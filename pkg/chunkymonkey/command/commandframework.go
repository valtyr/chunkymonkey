package cmd

import (
	"os"
	"strings"
)

var ErrCmdExists = os.NewError("The command already exists.")

// The CommandFramework handles all message based commands.
// It uses channels to safly handle multiple calls.
type CommandFramework struct {
	prefix     string // The command prefix befor every command.
	cmds       map[string]*Command
	Message    chan string
	modifyCmds chan *Command
}

// Creates a new CommandFramework and starts the update process.
func NewCommandFramework(prefix string) *CommandFramework {
	cf := &CommandFramework{prefix: prefix, cmds: make(map[string]*Command), Message: make(chan string, 100), modifyCmds: make(chan *Command, 10)}
	go cf.update()
	return cf
}

// Adds the command to the framework if the command already exists it will be overwritten.
// Commands without a CommandHandler in Func will be ignored.
func (cf *CommandFramework) AddCommand(command *Command) {
	cf.modifyCmds <- command
}

// Removes the command from the framework.
func (cf *CommandFramework) RemoveCommand(trigger string) {
	cf.modifyCmds <- &Command{Trigger: trigger}
}

func (cf *CommandFramework) update() {
	for {
		select {
		case command := <-cf.modifyCmds:
			if command == nil {
				continue
			}
			if len(command.Trigger) == 0 {
				continue
			}
			if command.Func == nil { // Remove
				cf.cmds[command.Trigger] = nil
			} else { // Add
				cf.cmds[command.Trigger] = command
			}
		case msg := <-cf.Message:
			if len(msg) < 2 || msg[0:len(cf.prefix)] != cf.prefix {
				continue
			}
			attr := strings.Split(msg, " ", -1)
			trigger := attr[0][1:]
			if command, ok := cf.cmds[trigger]; ok {
				command.Func(msg)
			}
		}
	}
}
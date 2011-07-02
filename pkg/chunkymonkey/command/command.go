package cmd

type CommandHandler func(string)

type Command struct {
	Trigger     string         // The initial text eg. "give".
	Description string         // A description of what the command does.
	Usage       string         // A usage string for the command.
	Func        CommandHandler // This function will be called if a Message begins with the CommandPrefix and the Trigger.
}

func NewCommand(trigger, desc, usage string, cmdHandler CommandHandler) *Command {
	return &Command{Trigger: trigger, Description: desc, Usage: usage, Func: cmdHandler}
}

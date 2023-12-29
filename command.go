package smtpServerProtocol

import (
	"strings"
)

// CommandName represent standard command name (sting) described in rfc5321
type CommandName string

// CommandEndSymbol contains separator to understand where command is ended and need generate reply.
const CommandEndSymbol = "\r\n"

// List predefined command names
const (
	CommandHelo = CommandName("HELO")
	CommandEhlo = CommandName("EHLO")
	CommandAuth = CommandName("AUTH")
	CommandMail = CommandName("MAIL")
	CommandRset = CommandName("RSET")
	CommandRcpt = CommandName("RCPT")
	CommandData = CommandName("DATA")
	CommandQuit = CommandName("QUIT")
)

// Command is a struct representing an SMTP command (verb + arguments)
type Command struct {
	verb CommandName
	args string
}

// CommandFromLine creates Command object form line string
func CommandFromLine(line string) *Command {
	parts := strings.SplitN(line, " ", 2)
	args := ""
	if len(parts) > 1 {
		args = parts[1]
	}
	return &Command{
		verb: CommandName(strings.ToUpper(parts[0])),
		args: args,
	}
}

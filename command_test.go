package smtpServerProtocol

import (
	"github.com/mailhedgehog/gounit"
	"testing"
)

func TestCommandFromLine(t *testing.T) {
	command := CommandFromLine("")
	(*gounit.T)(t).AssertEqualsString("", string(command.verb))
	(*gounit.T)(t).AssertEqualsString("", command.args)

	command = CommandFromLine(string(CommandQuit))
	(*gounit.T)(t).AssertEqualsString(string(CommandQuit), string(command.verb))
	(*gounit.T)(t).AssertEqualsString("", command.args)

	command = CommandFromLine(string(CommandAuth) + " PLAIN")
	(*gounit.T)(t).AssertEqualsString(string(CommandAuth), string(command.verb))
	(*gounit.T)(t).AssertEqualsString("PLAIN", command.args)

	command = CommandFromLine("foo bar baz")
	(*gounit.T)(t).AssertEqualsString("FOO", string(command.verb))
	(*gounit.T)(t).AssertEqualsString("bar baz", command.args)
}

func TestCommandStrings(t *testing.T) {
	(*gounit.T)(t).AssertEqualsString(string(CommandHelo), "HELO")
	(*gounit.T)(t).AssertEqualsString(string(CommandEhlo), "EHLO")
	(*gounit.T)(t).AssertEqualsString(string(CommandAuth), "AUTH")
	(*gounit.T)(t).AssertEqualsString(string(CommandMail), "MAIL")
	(*gounit.T)(t).AssertEqualsString(string(CommandRcpt), "RCPT")
	(*gounit.T)(t).AssertEqualsString(string(CommandData), "DATA")
	(*gounit.T)(t).AssertEqualsString(string(CommandQuit), "QUIT")
}

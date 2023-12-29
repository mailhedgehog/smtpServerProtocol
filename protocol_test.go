package smtpServerProtocol

import (
	"github.com/mailhedgehog/gounit"
	"github.com/mailhedgehog/smtpMessage"
	"testing"
)

func TestResetState(t *testing.T) {
	protocol := CreateProtocol("", nil, nil)

	protocol.state = StateData
	protocol.message.From, _ = smtpMessage.MessagePathFromString("<foo@bar.com>")

	(*gounit.T)(t).AssertEqualsString(string(StateData), string(protocol.state))
	(*gounit.T)(t).AssertEqualsString("foo@bar.com", protocol.message.From.Address())

	protocol.resetState()

	(*gounit.T)(t).AssertEqualsString(string(StateCommandsExchange), string(protocol.state))
	(*gounit.T)(t).AssertNil(protocol.message.From)
}

func TestSayHi(t *testing.T) {
	protocol := CreateProtocol("", nil, nil)
	reply := protocol.SayWelcome("   foo bar    ")

	(*gounit.T)(t).AssertEqualsInt(CODE_SERVICE_READY, reply.Status)
	(*gounit.T)(t).AssertEqualsString("foo bar Service ready", reply.lines[0])
}

func TestHandleReceivedLine(t *testing.T) {
	logManager().Warning("TODO: add test for HandleReceivedLine") // TODO
}

func TestHandleMailContent(t *testing.T) {
	protocol := CreateProtocol("", nil, nil)
	(*gounit.T)(t).AssertNil(protocol.handleMailContent("content foo bar baz"))
	(*gounit.T)(t).AssertNil(protocol.handleMailContent("content foo bar"))
	(*gounit.T)(t).AssertNil(protocol.handleMailContent("foo bar baz"))
}

func TestHandleCommandQUIT(t *testing.T) {
	protocol := CreateProtocol("", nil, nil)
	reply := protocol.handleCommand("QUIT")

	(*gounit.T)(t).AssertEqualsInt(CODE_SERVICE_CLOSING, reply.Status)
	(*gounit.T)(t).AssertEqualsInt(1, len(reply.lines))
	(*gounit.T)(t).AssertEqualsString("Bye", reply.lines[0])
}

func TestHandleCommandCommandFake(t *testing.T) {
	protocol := CreateProtocol("", nil, nil)
	reply := protocol.handleCommand("FAKE :)")

	(*gounit.T)(t).AssertEqualsInt(CODE_COMMAND_SYNTAX_ERROR, reply.Status)
}

func TestHELO(t *testing.T) {
	command := CommandFromLine("HELO foo.host.bar")
	protocol := CreateProtocol("", nil, nil)
	reply := protocol.HELO(command)

	(*gounit.T)(t).AssertEqualsInt(CODE_ACTION_OK, reply.Status)
	(*gounit.T)(t).AssertEqualsInt(1, len(reply.lines))
	(*gounit.T)(t).AssertEqualsString("Hello foo.host.bar", reply.lines[0])

	(*gounit.T)(t).AssertEqualsString("foo.host.bar", protocol.message.Helo)
}

func TestEHLO(t *testing.T) {
	command := CommandFromLine("EHLO foo.host.bar")
	protocol := CreateProtocol("", nil, nil)
	reply := protocol.EHLO(command)

	(*gounit.T)(t).AssertEqualsInt(CODE_ACTION_OK, reply.Status)
	(*gounit.T)(t).AssertEqualsInt(2, len(reply.lines))
	(*gounit.T)(t).AssertEqualsString("Hello foo.host.bar", reply.lines[0])
	(*gounit.T)(t).AssertEqualsString("PIPELINING", reply.lines[1])

	(*gounit.T)(t).AssertEqualsString("foo.host.bar", protocol.message.Helo)
}

func TestMAIL(t *testing.T) {
	command := CommandFromLine("MAIL FROM:<userx@y.foo.org>")
	protocol := CreateProtocol("", nil, nil)

	(*gounit.T)(t).AssertNil(protocol.message.From)

	reply := protocol.MAIL(command)

	(*gounit.T)(t).AssertEqualsInt(CODE_ACTION_OK, reply.Status)
	(*gounit.T)(t).AssertEqualsString("Sender userx@y.foo.org ok", reply.lines[0])

	(*gounit.T)(t).AssertEqualsString("userx@y.foo.org", protocol.message.From.Address())
}

func TestMAILFails(t *testing.T) {
	command := CommandFromLine("MAIL fake data")
	protocol := CreateProtocol("", nil, nil)

	reply := protocol.MAIL(command)

	(*gounit.T)(t).AssertEqualsInt(CODE_MAILBOX_404, reply.Status)
	(*gounit.T)(t).AssertEqualsString("Invalid syntax in MAIL command", reply.lines[0])
}

func TestRCPT(t *testing.T) {
	protocol := CreateProtocol("", nil, nil)

	(*gounit.T)(t).AssertEqualsInt(0, len(protocol.message.To))

	command := CommandFromLine("RCPT TO:<userx@y.foo.org>")
	reply := protocol.RCPT(command)
	(*gounit.T)(t).AssertEqualsInt(CODE_ACTION_OK, reply.Status)
	(*gounit.T)(t).AssertEqualsString("Receiver userx@y.foo.org ok", reply.lines[0])

	command = CommandFromLine("RCPT TO:<user2@y.foo.org>")
	reply = protocol.RCPT(command)
	(*gounit.T)(t).AssertEqualsInt(CODE_ACTION_OK, reply.Status)
	(*gounit.T)(t).AssertEqualsString("Receiver user2@y.foo.org ok", reply.lines[0])

	(*gounit.T)(t).AssertEqualsInt(2, len(protocol.message.To))
	(*gounit.T)(t).AssertEqualsString("user2@y.foo.org", protocol.message.To[1].Address())
}

func TestRCPTFails(t *testing.T) {
	protocol := CreateProtocol("", nil, nil)
	command := CommandFromLine("RCPT fake")
	reply := protocol.RCPT(command)
	(*gounit.T)(t).AssertEqualsInt(CODE_MAILBOX_404, reply.Status)
	(*gounit.T)(t).AssertEqualsString("Invalid syntax in MAIL command", reply.lines[0])
}

func TestGetAuthMechanisms(t *testing.T) {
	protocol := CreateProtocol("", nil, nil)
	(*gounit.T)(t).AssertEqualsInt(0, len(protocol.supportedAuthMechanisms))

	protocol.supportedAuthMechanisms = []string{"PLAIN", "foo", "BAR"}

	mechanisms := protocol.supportedAuthMechanisms
	(*gounit.T)(t).AssertEqualsInt(3, len(mechanisms))
	(*gounit.T)(t).AssertEqualsString("PLAIN", mechanisms[0])
	(*gounit.T)(t).AssertEqualsString("foo", mechanisms[1])
	(*gounit.T)(t).AssertEqualsString("BAR", mechanisms[2])
}

func TestParseAuthMechanism(t *testing.T) {
	protocol := CreateProtocol("", nil, nil)

	(*gounit.T)(t).AssertEqualsString("", protocol.parseAuthMechanism(""))
	(*gounit.T)(t).AssertEqualsString("foo", protocol.parseAuthMechanism("foo"))
	(*gounit.T)(t).AssertEqualsString("BAR", protocol.parseAuthMechanism("BAR"))
	(*gounit.T)(t).AssertEqualsString("foo", protocol.parseAuthMechanism("foo baz"))
}

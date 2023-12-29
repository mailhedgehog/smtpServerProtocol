package smtpServerProtocol

import (
	"errors"
	"fmt"
	"github.com/mailhedgehog/smtpMessage"
	"golang.org/x/exp/slices"
	"net"
	"reflect"
	"regexp"
	"strings"
)

// ConversationState represents on what stage now current Client<->Server conversation.
type ConversationState string

const (
	StateCommandsExchange = ConversationState("commands_exchange")
	StateWaitingAuth      = ConversationState("waiting_auth")
	StateData             = ConversationState("data")
	StateCustomScene      = ConversationState("custom_scene")
)

// Validation allows to send to package custom validation parameters what accepts server
type Validation struct {
	MaximumLineLength int
	MaximumReceivers  int
}

// Protocol represents rfc5321 described protocol conversation
type Protocol struct {
	Hostname   string
	Ip         *net.TCPAddr
	validation *Validation

	state      ConversationState
	message    *smtpMessage.SmtpMessage
	tempOrigin string

	// supportedAuthMechanisms can be empty, if empty client will not go through auth flow
	supportedAuthMechanisms []string
	messageReceivedCallback func(message *smtpMessage.SmtpMessage) (string, error)

	createCustomSceneCallback func(sceneName string) Scene
	currentScene              Scene
}

func CreateProtocol(hostname string, ip *net.TCPAddr, validation *Validation) *Protocol {
	if validation == nil {
		validation = &Validation{
			MaximumLineLength: 0,
			MaximumReceivers:  0,
		}
	}

	protocol := &Protocol{
		Hostname:   hostname,
		Ip:         ip,
		validation: validation,
	}
	protocol.resetState()

	return protocol
}

func (protocol *Protocol) SetAuthMechanisms(authMechanisms []string) {
	protocol.supportedAuthMechanisms = authMechanisms
}

// OnMessageReceived allow to provide custom success callback.
func (protocol *Protocol) OnMessageReceived(callback func(message *smtpMessage.SmtpMessage) (string, error)) {
	protocol.messageReceivedCallback = callback
}

func (protocol *Protocol) CreateCustomSceneUsing(callback func(sceneName string) Scene) {
	protocol.createCustomSceneCallback = callback
}

func (protocol *Protocol) SetStateCommandsExchange() {
	protocol.state = StateCommandsExchange
}

func (protocol *Protocol) SayWelcome(identification string) *Reply {
	identification = strings.TrimSpace(identification)
	if len(identification) > 0 {
		identification = identification + " "
	}
	hostname := protocol.Hostname
	if len(hostname) > 0 {
		hostname = hostname + " "
	}
	protocol.state = StateCommandsExchange
	return ReplyServiceReady(hostname + identification + "Service ready")
}

func (protocol *Protocol) HandleReceivedLine(receivedLine string) *Reply {
	if protocol.validation.MaximumLineLength > 0 && len(receivedLine) > 0 {
		if len(receivedLine) > protocol.validation.MaximumLineLength {
			return ReplyLineTooLong()
		}
	}

	if protocol.state == StateCustomScene {
		if protocol.currentScene != nil {
			return protocol.currentScene.ReadAndWriteReply(receivedLine)
		}
		return ReplyCommandNotImplemented()
	}

	if protocol.state == StateData {
		return protocol.handleMailContent(receivedLine)
	}

	return protocol.handleCommand(receivedLine)
}

func (protocol *Protocol) resetState() {
	protocol.message = &smtpMessage.SmtpMessage{
		ID: smtpMessage.NewMessageID(),
	}
	protocol.tempOrigin = ""
	protocol.SetStateCommandsExchange()
}

func (protocol *Protocol) handleMailContent(receivedLine string) *Reply {
	protocol.tempOrigin += receivedLine + "\r\n"

	// Check is this is end
	if strings.HasSuffix(protocol.tempOrigin, "\r\n.\r\n") {
		protocol.tempOrigin = strings.ReplaceAll(protocol.tempOrigin, "\r\n..", "\r\n.")

		logManager().Debug("Got EOF, storing message and reset state.")
		protocol.tempOrigin = strings.TrimSuffix(protocol.tempOrigin, "\r\n.\r\n")
		protocol.state = StateCommandsExchange

		defer protocol.resetState()

		if protocol.messageReceivedCallback == nil {
			logManager().Error("No receive callback processed")
			return ReplyExceededStorage("No storage backend")
		}

		var err error
		var messageId string

		err = protocol.message.SetOrigin(protocol.tempOrigin)
		if err != nil {
			logManager().Error(fmt.Sprintf("Error storing message origin: %s", err.Error()))
			return ReplyExceededStorage("Unable to store message")
		}

		messageId, err = protocol.messageReceivedCallback(protocol.message)
		if err != nil {
			logManager().Error(fmt.Sprintf("Error storing message: %s", err.Error()))
			return ReplyExceededStorage("Unable to store message")
		}

		logManager().Debug("Message processed and returns success.")
		return ReplyOk("Ok: queued as " + messageId)
	}

	return nil
}

func (protocol *Protocol) handleCommand(receivedLine string) *Reply {
	receivedLine = strings.Trim(receivedLine, "\r\n")
	command := CommandFromLine(receivedLine)

	logManager().Debug(fmt.Sprintf("Handle command: '%s', with args: '%s'", command.verb, command.args))

	if protocol.state == StateWaitingAuth && command.verb != CommandAuth {
		return ReplyAuthFailed("")
	}

	st := reflect.ValueOf(protocol)
	m := st.MethodByName("command" + string(command.verb))
	if m.IsValid() {
		m.Call([]reflect.Value{reflect.ValueOf(command)})
	}

	switch command.verb {
	case CommandHelo:
		return protocol.HELO(command)
	case CommandEhlo:
		return protocol.EHLO(command)
	case CommandAuth:
		logManager().Debug(fmt.Sprintf("Got %s command", command.verb))

		authMechanism := protocol.parseAuthMechanism(command.args)
		if slices.Contains(protocol.supportedAuthMechanisms, authMechanism) && protocol.createCustomSceneCallback != nil {
			reply, err := protocol.startCustomScene(string(command.verb)+"_"+authMechanism, receivedLine)
			if err == nil {
				return reply
			}
		}
		return ReplyCommandNotImplemented()
	case CommandRset:
		return protocol.RSET(command)
	case CommandMail:
		return protocol.MAIL(command)
	case CommandRcpt:
		return protocol.RCPT(command)
	case CommandData:
		protocol.state = StateData
		return ReplyMailData()
	case CommandQuit:
		return ReplyBye()
	default:
		return ReplyUnrecognisedCommand()
	}
}

func (protocol *Protocol) startCustomScene(customSceneName string, receivedLine string) (*Reply, error) {
	protocol.currentScene = protocol.createCustomSceneCallback(customSceneName)
	if protocol.currentScene != nil {
		protocol.state = StateCustomScene
		return protocol.currentScene.Start(receivedLine, protocol), nil
	}

	return nil, errors.New(fmt.Sprintf("custom scene not provided [%s]", customSceneName))
}

func (protocol *Protocol) HELO(command *Command) *Reply {
	protocol.message.Helo = command.args

	if len(protocol.supportedAuthMechanisms) > 0 {
		protocol.state = StateWaitingAuth
	}

	return ReplyOk("Hello " + command.args)
}

func (protocol *Protocol) EHLO(command *Command) *Reply {
	protocol.message.Helo = command.args
	replyArgs := []string{"Hello " + command.args, "PIPELINING"}

	logManager().Warning("TODO: add tls support") // TODO

	if len(protocol.supportedAuthMechanisms) > 0 {
		protocol.state = StateWaitingAuth
		replyArgs = append(replyArgs, string(CommandAuth)+" "+strings.Join(protocol.supportedAuthMechanisms, " "))
	}
	return ReplyOk(replyArgs...)
}

func (protocol *Protocol) RSET(command *Command) *Reply {
	protocol.resetState()

	return ReplyOk("")
}

func (protocol *Protocol) MAIL(command *Command) *Reply {
	match := regexp.MustCompile(`(?i:From):\s*(.+)`).FindStringSubmatch(command.args)

	if len(match) != 2 {
		return ReplyMailbox404("Invalid syntax in MAIL command")
	}

	var err error
	protocol.message.From, err = smtpMessage.MessagePathFromString(match[1])
	if err != nil {
		return ReplyMailbox404(err.Error())
	}

	return ReplyOk("Sender " + protocol.message.From.Address() + " ok")
}

func (protocol *Protocol) RCPT(command *Command) *Reply {
	if protocol.validation.MaximumReceivers > 0 && len(protocol.message.To) >= protocol.validation.MaximumReceivers {
		return ReplyExceededStorage("Maximum receivers extended")
	}
	match := regexp.MustCompile(`(?i:To):\s*(.+)`).FindStringSubmatch(command.args)

	if len(match) != 2 {
		return ReplyMailbox404("Invalid syntax in MAIL command")
	}

	mailPath, err := smtpMessage.MessagePathFromString(match[1])
	if err != nil {
		return ReplyMailbox404(err.Error())
	}

	protocol.message.To = append(protocol.message.To, mailPath)

	return ReplyOk("Receiver " + mailPath.Address() + " ok")
}

func (protocol *Protocol) parseAuthMechanism(args string) string {
	parts := strings.SplitN(args, " ", 2)

	return parts[0]
}

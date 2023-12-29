package smtpServerProtocol

import "strconv"

// Reply is a struct representing an SMTP reply (status code + lines)
type Reply struct {
	Status int
	lines  []string
}

// LIst of predefined by rfc5321 list of status codes.
const (
	CODE_SYSTEM_STATUS             = 211
	CODE_HELP_MESSAGE              = 214
	CODE_SERVICE_READY             = 220
	CODE_SERVICE_CLOSING           = 221
	CODE_AUTHENTICATION_SUCCESS    = 235
	CODE_ACTION_OK                 = 250
	CODE_USER_IS_NOT_LOCAL         = 251 // will forward to <forward-path>
	CODE_USER_NOT_VERIFIED         = 252
	CODE_AUTH_CREDENTIALS          = 334
	CODE_MAIL_DATA                 = 354
	CODE_SERVICE_NOT_AVAILABLE     = 421
	CODE_MAILBOX_UNAVAILABLE       = 450
	CODE_LOCAL_ERROR               = 451
	CODE_EXCEEDED_SYSTEM_STORAGE   = 452
	CODE_COMMAND_SYNTAX_ERROR      = 500
	CODE_PARAMETER_SYNTAX_ERROR    = 501
	CODE_COMMAND_NOT_IMPLEMENTED   = 502
	CODE_COMMANDS_BAD_SEQUENCE     = 503
	CODE_PARAMETER_NOT_IMPLEMENTED = 504
	CODE_AUTH_FAILED               = 535
	CODE_MAILBOX_404               = 550
	CODE_USER_NOT_LOCAL            = 551 // please try <forward-path>
	CODE_EXCEEDED_STORAGE          = 552
	CODE__MAILBOX_NAME_INCORRECT   = 553
	CODE_TRANSACTION_FAILED        = 554
)

// FormattedLines returns the formatted SMTP reply lines.
func (r Reply) FormattedLines() []string {
	var lines []string

	if len(r.lines) == 0 {
		l := strconv.Itoa(r.Status)
		lines = append(lines, l+"\n")
		return lines
	}

	for i, line := range r.lines {
		l := ""
		if i == len(r.lines)-1 {
			l = strconv.Itoa(r.Status) + " " + line + CommandEndSymbol
		} else {
			l = strconv.Itoa(r.Status) + "-" + line + CommandEndSymbol
		}
		lines = append(lines, l)
	}

	return lines
}

// ReplyServiceReady creates a welcome reply.
func ReplyServiceReady(identification string) *Reply {
	return &Reply{CODE_SERVICE_READY, []string{identification}}
}

// ReplyBye used on close connection.
func ReplyBye() *Reply { return &Reply{CODE_SERVICE_CLOSING, []string{"Bye"}} }

// ReplyAuthOk creates a authentication successful reply.
func ReplyAuthOk() *Reply {
	return &Reply{CODE_AUTHENTICATION_SUCCESS, []string{"Authenticate successful"}}
}

// ReplyOk represents generic success response.
func ReplyOk(message ...string) *Reply {
	if len(message) == 0 {
		message = []string{"Ok"}
	}
	return &Reply{CODE_ACTION_OK, message}
}

func ReplyUnrecognisedCommand() *Reply {
	return &Reply{CODE_COMMAND_SYNTAX_ERROR, []string{"Unrecognised command"}}
}

func ReplyCommandNotImplemented() *Reply {
	return &Reply{CODE_COMMAND_NOT_IMPLEMENTED, []string{"Command not implemented"}}
}

// ReplyLineTooLong due to exceeding these limits
func ReplyLineTooLong() *Reply {
	return &Reply{CODE_COMMAND_SYNTAX_ERROR, []string{"Line too long."}}
}

// ReplyAuthCredentials creates reply with a 334 code and requests a username
func ReplyAuthCredentials(response string) *Reply {
	return &Reply{CODE_AUTH_CREDENTIALS, []string{response}}
}

func ReplyAuthFailed(response string) *Reply {
	if len(response) <= 0 {
		response = "Authenticate failed"
	}
	return &Reply{CODE_AUTH_FAILED, []string{response}}
}

func ReplyMailbox404(response string) *Reply {
	return &Reply{CODE_MAILBOX_404, []string{response}}
}

func ReplyExceededStorage(response string) *Reply {
	return &Reply{CODE_EXCEEDED_STORAGE, []string{response}}
}

func ReplyMailData() *Reply {
	return &Reply{CODE_MAIL_DATA, []string{"End data with <CR><LF>.<CR><LF>"}}
}

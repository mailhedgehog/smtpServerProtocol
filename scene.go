package smtpServerProtocol

// Scene represents custom logic flow (scene) for some specific
// set of commands, for example authentication.
type Scene interface {
	// Start scene by send specific message (reply) to client.
	Start(receivedLine string, protocol *Protocol) *Reply
	// ReadAndWriteReply reads client message and write reply
	ReadAndWriteReply(receivedLine string) *Reply
	// Finish scene, by notifying protocol to finish this scene
	Finish()
}

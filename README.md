## SMTP server protocol

RFC spec you can read [here](rfc5321.txt)

#### Authentication

Some informative doc can be found [here](https://mailtrap.io/blog/smtp-auth/)

## Usage

```go
protocol := smtpServerProtocol.CreateProtocol(
    hostname,
    ipAddr,
    &smtpServerProtocol.Validation{
        MaximumLineLength: context.Config.Smtp.Validation.MaximumLineLength,
        MaximumReceivers:  context.Config.Smtp.Validation.MaximumReceivers,
    },
)

protocol.OnMessageReceived(func(message *smtpMessage.SmtpMessage) (string, error) {
    messageId, err := savemessage(message)
    
    return messageId, err
})
```

## Development

```shell
go mod tidy
go mod verify
go mod vendor
go test --cover
```

## Credits

- [![Think Studio](https://yaroslawww.github.io/images/sponsors/packages/logo-think-studio.png)](https://think.studio/)

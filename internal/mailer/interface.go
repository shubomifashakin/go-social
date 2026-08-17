package mailer

type Mail struct {
	From        string
	To          []string
	Subject     string
	Html        string
	Cc          []string
	Attachments []MailContent
}

type MailContent struct {
	Content     []byte
	FileName    string
	ContentType string
}

type MailerService interface {
	SendMail(body Mail) (string, error)
}

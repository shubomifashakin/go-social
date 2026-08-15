package mailer

import "github.com/resend/resend-go/v3"

type Mailer struct {
	client *resend.Client
}

type mail struct {
	From string
	To []string
	Subject string
	Html string
	Attachments []*resend.Attachment
	Cc []string
}

func (m *Mailer) SendMail(body mail)(string,error) {
	res,err:=m.client.Emails.Send(&resend.SendEmailRequest{
		From: body.From,
		To: body.To,
		Subject: body.Subject,
		Html: body.Html,
		Attachments: body.Attachments,
		Cc:body.Cc ,
	})
	
	if err != nil {
		return "",err
	}

	return res.Id,nil
}

func NewMailer(apiKey string) *Mailer {
	return	&Mailer{
		client: resend.NewClient(apiKey),
	}
}
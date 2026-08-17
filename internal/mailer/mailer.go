package mailer

import (
	"github.com/resend/resend-go/v3"
)

type Mailer struct {
	client *resend.Client
}

func (m *Mailer)SendMail(body Mail)(string,error) {
	attachments:= []*resend.Attachment{}

	for _,v:=range body.Attachments {
		attachments=append(attachments, &resend.Attachment{
			Content: v.Content,
			Filename: v.FileName,
			ContentType: v.ContentType,
			Path: "",
			ContentId: "",
		})
	}

	res,err:=m.client.Emails.Send(&resend.SendEmailRequest{
		From: body.From,
		To: body.To,
		Subject: body.Subject,
		Html: body.Html,
		Attachments: attachments,
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
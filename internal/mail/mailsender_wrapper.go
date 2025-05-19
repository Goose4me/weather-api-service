package mail

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mailersend/mailersend-go"
)

type MailSenderWrapper struct {
	apiKey string
	ms     *mailersend.Mailersend
}

func NewMailSenderWrapper(apiKey string) *MailSenderWrapper {
	ms := mailersend.NewMailersend(apiKey)
	return &MailSenderWrapper{apiKey: apiKey, ms: ms}
}

func (msw *MailSenderWrapper) SendMail(subject, html, text string, recipients []mailersend.Recipient) int {
	ms := msw.ms

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	from := mailersend.From{
		Name:  "Weather App",
		Email: os.Getenv("MAILSENDER_EMAIL"),
	}

	message := ms.Email.NewMessage()

	message.SetFrom(from)
	message.SetRecipients(recipients)
	message.SetSubject(subject)
	message.SetHTML(html)
	message.SetText(text)

	res, _ := ms.Email.Send(ctx, message)

	fmt.Println(res.Header.Get("X-Message-Id"))

	return res.StatusCode
}

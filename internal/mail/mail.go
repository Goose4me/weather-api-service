package mail

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mailersend/mailersend-go"
	"gorm.io/gorm"
)

type MailService struct {
	DB *gorm.DB
}

func NewMailService(db *gorm.DB) *MailService {
	return &MailService{DB: db}
}

func (srv *MailService) SendConfirmationMail(email, confirmationUrl, unsubscribeUrl string) error {
	APIKey := os.Getenv("MAILSENDER_API_KEY")

	ms := mailersend.NewMailersend(APIKey)
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	data := ConfirmationData{
		ConfirmURL:     confirmationUrl,
		UnsubscribeURL: unsubscribeUrl,
	}

	subject := "Confirm your subscription"
	text := fmt.Sprintf("Confirm your mail using %s.\n Unsubscribe with %s", data.ConfirmURL, data.UnsubscribeURL)
	html, err := FormConfirmationMail(&data)

	if err != nil {
		return err
	}

	from := mailersend.From{
		Name:  "Weather App",
		Email: os.Getenv("MAILSENDER_EMAIL"),
	}

	recipients := []mailersend.Recipient{
		{
			Email: email,
		},
	}

	message := ms.Email.NewMessage()

	message.SetFrom(from)
	message.SetRecipients(recipients)
	message.SetSubject(subject)
	message.SetHTML(html)
	message.SetText(text)

	res, _ := ms.Email.Send(ctx, message)

	fmt.Println(res.Header.Get("X-Message-Id"))

	return nil
}

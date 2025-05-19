package mail

import (
	"bytes"
	"html/template"
)

const confirmationEmailHTML = `
<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8">
    <title>Confirm your subscription</title>
  </head>
  <body style="font-family: sans-serif; background-color: #f7f7f7; padding: 20px;">
    <div style="max-width: 600px; margin: auto; background: #ffffff; padding: 30px; border-radius: 8px; box-shadow: 0 2px 5px rgba(0,0,0,0.1);">
      <h2 style="color: #333333;">Confirm your subscription</h2>
      <p style="font-size: 16px; color: #555555;">
        Hi there! Please confirm your email address by clicking the button below:
      </p>
      <p style="text-align: center; margin: 30px 0;">
        <a href="{{.ConfirmURL}}" style="background-color: #007BFF; color: white; padding: 12px 20px; text-decoration: none; border-radius: 5px;">
          Confirm Email
        </a>
      </p>
      <p style="font-size: 14px; color: #888888;">
        If you didn’t request this, you can safely ignore this email.
      </p>
      <hr style="margin: 40px 0; border: none; border-top: 1px solid #eeeeee;">
      <p style="font-size: 12px; color: #999999; text-align: center;">
        Not interested? 
        <a href="{{.UnsubscribeURL}}" style="color: #007BFF; text-decoration: none;">Unsubscribe</a>
      </p>
    </div>
  </body>
</html>
`

type ConfirmationData struct {
	ConfirmURL     string
	UnsubscribeURL string
}

func FormConfirmationMail(confirmData *ConfirmationData) (string, error) {
	tmpl, err := template.New("confirm").Parse(confirmationEmailHTML)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, confirmData); err != nil {
		return "", err
	}

	return buf.String(), nil
}

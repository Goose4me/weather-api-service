package mail_templates

import (
	"bytes"
	"html/template"
)

const weatherUpdateEmailHTML = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Weather Update</title>
</head>
<body style="font-family: Arial, sans-serif; background-color: #f7f7f7; padding: 20px;">
  <div style="max-width: 600px; margin: auto; background-color: #ffffff; padding: 30px; border-radius: 8px; box-shadow: 0 2px 5px rgba(0,0,0,0.1);">
    
    <h2 style="color: #333333;">Your Weather Update for {{.City}}</h2>

    <p style="font-size: 16px; color: #555555;">
      Here's your latest forecast:
    </p>

    <ul style="font-size: 16px; color: #444444;">
      <li><strong>Temperature:</strong> {{.Temperature}}°C</li>
      <li><strong>Humidity:</strong> {{.Humidity}}%</li>
      <li><strong>Condition:</strong> {{.Description}}</li>
    </ul>

    <p style="margin-top: 30px; font-size: 14px; color: #888888;">
      Stay safe and dress appropriately for today's weather!
    </p>

    <hr style="margin: 40px 0; border: none; border-top: 1px solid #eeeeee;">

    <p style="font-size: 12px; color: #999999; text-align: center;">
      Don’t want to receive updates? 
      <a href="{{.UnsubscribeURL}}" style="color: #007BFF; text-decoration: none;">Unsubscribe here</a>.
    </p>

  </div>
</body>
</html>
`

type WeatherUpdateData struct {
	City           string
	Temperature    float64
	Humidity       int
	Description    string
	UnsubscribeURL string
}

func FormWeatherUpdateMail(confirmData *WeatherUpdateData) (string, error) {
	tmpl, err := template.New("weather").Parse(weatherUpdateEmailHTML)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, confirmData); err != nil {
		return "", err
	}

	return buf.String(), nil
}

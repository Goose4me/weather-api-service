package mail

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"
	"weather-app/internal/database"
	"weather-app/internal/mail/mail_templates"
	"weather-app/internal/weather"

	"github.com/mailersend/mailersend-go"
	"gorm.io/gorm"
)

type MailService struct {
	DB *gorm.DB
}

func NewMailService(db *gorm.DB) *MailService {
	return &MailService{DB: db}
}

type UpdateType int

const (
	Hourly UpdateType = iota
	Daily
)

var updateTypeName = map[UpdateType]string{
	Hourly: "hourly",
	Daily:  "daily",
}

func sendMail(subject, html, text string, recipients []mailersend.Recipient) {
	APIKey := os.Getenv("MAILSENDER_API_KEY")

	ms := mailersend.NewMailersend(APIKey)
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
}

// TODO: Move to other place. Should be common
func BuildTokenURL(base, apiPath, token string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	// Join the path and token properly
	u.Path = path.Join(u.Path, apiPath, token)

	return u.String(), nil
}

func (srv *MailService) SendConfirmationMail(email, confirmationUrl, unsubscribeUrl string) error {

	data := mail_templates.ConfirmationData{
		ConfirmURL:     confirmationUrl,
		UnsubscribeURL: unsubscribeUrl,
	}

	subject := "Confirm your subscription"
	text := fmt.Sprintf("Confirm your mail using %s.\n Unsubscribe with %s", data.ConfirmURL, data.UnsubscribeURL)
	html, err := mail_templates.FormConfirmationMail(&data)

	if err != nil {
		return err
	}

	recipients := []mailersend.Recipient{
		{
			Email: email,
		},
	}

	sendMail(subject, html, text, recipients)

	return nil
}

func buildWeatherURL(baseURL, city string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	u.Path = "/api/weather"

	// Add query parameters
	q := u.Query()
	q.Set("city", city)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func callWeatherAPI(city string) (*weather.WeatherData, error) {
	var result weather.WeatherData

	url, err := buildWeatherURL(os.Getenv("WEATHER_APP_BASE_URL"), city)

	if err != nil {
		return nil, err
	}

	resp, err := http.Get(url)
	if err != nil {
		fetchErr := fmt.Errorf("failed to fetch weather: %s", err.Error())

		return nil, fetchErr
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode == http.StatusNotFound {
			return nil, weather.ErrCityNotFound
		} else {
			return nil, fmt.Errorf("API error %s", string(body))
		}
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		invalidResponceErr := fmt.Errorf("invalid response JSON: %s", err.Error())

		return nil, invalidResponceErr
	}

	return &result, nil
}

func (srv *MailService) SendWeatherUpdate(updateType UpdateType) error {

	offset := 0
	limit := 100
	var globalError error

	globalError = nil

	for {
		batch, err := database.GetNextUserEmailInfoBatch(limit, offset, updateTypeName[updateType], srv.DB)

		if err != nil {
			return fmt.Errorf("failed to load batch: %v", err)
		}
		if len(batch) == 0 {
			break
		}

		for _, entry := range batch {
			log.Printf("Send %s to %s for city %s with token %s\n", updateTypeName[updateType], entry.Email, entry.City, entry.TokenValue)
			data, err := callWeatherAPI(entry.City)

			if err != nil {
				log.Printf("call weather API error: %s\n", err.Error())
				globalError = err

				continue
			}

			log.Printf("Temperature: %1.f\nHumidity:%d\nDescription:%s", data.Temperature, data.Humidity, data.Description)

			subject := fmt.Sprintf("Weather update for %s", entry.City)

			unsubscribeUrl, _ := BuildTokenURL(os.Getenv("BASE_URL"), "/api/unsubscribe/", entry.TokenValue)

			weatherData := mail_templates.WeatherUpdateData{
				City:           entry.City,
				Temperature:    data.Temperature,
				Humidity:       data.Humidity,
				Description:    data.Description,
				UnsubscribeURL: unsubscribeUrl,
			}

			text := fmt.Sprintf("Weather update for %s", unsubscribeUrl)
			html, _ := mail_templates.FormWeatherUpdateMail(&weatherData)

			// TODO: Send not one by one, but group by city
			recipients := []mailersend.Recipient{
				{
					Email: entry.Email,
				},
			}

			sendMail(subject, html, text, recipients)
		}

		offset += limit
	}

	return globalError
}

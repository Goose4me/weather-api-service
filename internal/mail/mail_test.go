package mail_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"weather-app/internal/database/repository"
	"weather-app/internal/mail"
	"weather-app/internal/weather"

	"github.com/mailersend/mailersend-go"
)

type mockSender struct {
	Called      bool
	LastSubject string
}

func (m *mockSender) SendMail(subject, html, text string, recipients []mailersend.Recipient) int {
	m.Called = true
	m.LastSubject = subject
	return 1
}

type mockUserRepo struct {
	batch []repository.UserEmailInfo
	err   error
}

func (m *mockUserRepo) GetUserEmailInfoBatch(limit, offset int, subscriptionFrequency string) ([]repository.UserEmailInfo, error) {
	if offset > 0 {
		return make([]repository.UserEmailInfo, 0), m.err
	}

	return m.batch, m.err
}

func TestSendConfirmationMail_Success(t *testing.T) {
	sender := &mockSender{}
	svc := mail.NewMailService(nil, sender)

	err := svc.SendConfirmationMail("user@example.com", "http://confirm", "http://unsubscribe")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !sender.Called {
		t.Error("expected SendMail to be called")
	}
}

func TestSendWeatherUpdate_Success(t *testing.T) {
	// Start mock weather API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := weather.WeatherData{
			Temperature: 23.5,
			Humidity:    60,
			Description: "sunny",
		}
		json.NewEncoder(w).Encode(data)
	}))
	defer server.Close()

	os.Setenv("WEATHER_APP_BASE_URL", server.URL)
	os.Setenv("BASE_URL", "http://localhost:8080")

	userRepo := &mockUserRepo{
		batch: []repository.UserEmailInfo{
			{Email: "test@example.com", City: "Kyiv", TokenValue: "abc123"},
		},
	}

	sender := &mockSender{}
	svc := mail.NewMailService(userRepo, sender)

	err := svc.SendWeatherUpdate(mail.Daily)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !sender.Called {
		t.Error("expected SendMail to be called")
	}
}

func TestSendWeatherUpdate_DBError(t *testing.T) {
	userRepo := &mockUserRepo{
		err: errors.New("DB failure"),
	}

	svc := mail.NewMailService(userRepo, &mockSender{})

	err := svc.SendWeatherUpdate(mail.Hourly)
	if err == nil || err.Error() != "failed to load batch: DB failure" {
		t.Errorf("expected DB error, got %v", err)
	}
}

func TestSendWeatherUpdate_ErrCityNotFound(t *testing.T) {
	// Start mock weather API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"city not found"}`, http.StatusNotFound)
	}))
	defer server.Close()

	os.Setenv("WEATHER_APP_BASE_URL", server.URL)
	os.Setenv("BASE_URL", "http://localhost:8080")

	userRepo := &mockUserRepo{
		batch: []repository.UserEmailInfo{
			{Email: "test@example.com", City: "Kyiv", TokenValue: "abc123"},
		},
	}

	sender := &mockSender{}
	svc := mail.NewMailService(userRepo, sender)

	err := svc.SendWeatherUpdate(mail.Daily)
	if err != weather.ErrCityNotFound {
		t.Errorf("expected ErrCityNotFound, got %v", err)
	}
}

func TestSendWeatherUpdate_OtherHTTPError(t *testing.T) {
	// Start mock weather API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer server.Close()

	os.Setenv("WEATHER_APP_BASE_URL", server.URL)
	os.Setenv("BASE_URL", "http://localhost:8080")

	userRepo := &mockUserRepo{
		batch: []repository.UserEmailInfo{
			{Email: "test@example.com", City: "Kyiv", TokenValue: "abc123"},
		},
	}

	sender := &mockSender{}
	svc := mail.NewMailService(userRepo, sender)

	err := svc.SendWeatherUpdate(mail.Daily)
	if err == nil || err.Error() != "API error internal error\n" {
		t.Errorf("expected API error, got %v", err)
	}
}

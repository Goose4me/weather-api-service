package subscription_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"weather-app/internal/subscription"
)

type mockSubscriptionService struct {
	SubscribeFunc   func(email, city, frequency string) error
	ConfirmFunc     func(tokenValue string) error
	UnsubscribeFunc func(tokenValue string) error
}

func (m *mockSubscriptionService) Subscribe(email, city, frequency string) error {
	return m.SubscribeFunc(email, city, frequency)
}

func (m *mockSubscriptionService) Confirm(tokenValue string) error {
	return m.ConfirmFunc(tokenValue)
}

func (m *mockSubscriptionService) Unsubscribe(tokenValue string) error {
	return m.UnsubscribeFunc(tokenValue)
}

func TestSubscribeHandler_Success(t *testing.T) {
	form := url.Values{}
	form.Set("email", "test@example.com")
	form.Set("city", "Kyiv")
	form.Set("frequency", "daily")

	svc := &mockSubscriptionService{
		SubscribeFunc: func(email, city, freq string) error {
			return nil
		},
	}

	req := httptest.NewRequest("POST", "/api/subscribe", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler := subscription.NewHandler(svc)
	handler.SubscribeHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestSubscribeHandler_InvalidEmail(t *testing.T) {
	form := url.Values{}
	form.Set("email", "not-an-email")
	form.Set("city", "Kyiv")
	form.Set("frequency", "daily")

	svc := &mockSubscriptionService{}

	req := httptest.NewRequest("POST", "/api/subscribe", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler := subscription.NewHandler(svc)
	handler.SubscribeHandler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestSubscribeHandler_UserExists(t *testing.T) {
	form := url.Values{}
	form.Set("email", "test@example.com")
	form.Set("city", "Kyiv")
	form.Set("frequency", "daily")

	svc := &mockSubscriptionService{
		SubscribeFunc: func(email, city, freq string) error {
			return subscription.ErrUserAlreadyExists
		},
	}

	req := httptest.NewRequest("POST", "/api/subscribe", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler := subscription.NewHandler(svc)
	handler.SubscribeHandler(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestSubscribeHandler_UnsupportedMethod(t *testing.T) {
	svc := &mockSubscriptionService{}

	req := httptest.NewRequest("GET", "/api/subscribe", nil)
	w := httptest.NewRecorder()

	handler := subscription.NewHandler(svc)
	handler.SubscribeHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Unsupported method") {
		t.Errorf("expected method error message, got: %s", w.Body.String())
	}
}

func TestConfirmHandler_Success(t *testing.T) {
	svc := &mockSubscriptionService{
		ConfirmFunc: func(token string) error {
			return nil
		},
	}

	req := httptest.NewRequest("GET", "/api/confirm/token123", nil)
	w := httptest.NewRecorder()

	handler := subscription.NewHandler(svc)
	handler.ConfirmHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestConfirmHandler_TokenNotFound(t *testing.T) {
	svc := &mockSubscriptionService{
		ConfirmFunc: func(token string) error {
			return subscription.ErrTokenNotFound
		},
	}

	req := httptest.NewRequest("GET", "/api/confirm/token123", nil)
	w := httptest.NewRecorder()

	handler := subscription.NewHandler(svc)
	handler.ConfirmHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestConfirmHandler_UnsupportedMethod(t *testing.T) {
	svc := &mockSubscriptionService{}

	req := httptest.NewRequest("POST", "/api/confirm/token123", nil)
	w := httptest.NewRecorder()

	handler := subscription.NewHandler(svc)
	handler.ConfirmHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Unsupported method") {
		t.Errorf("expected method error message, got: %s", w.Body.String())
	}
}

func TestUnsubscribeHandler_Success(t *testing.T) {
	svc := &mockSubscriptionService{
		UnsubscribeFunc: func(token string) error {
			return nil
		},
	}

	req := httptest.NewRequest("GET", "/api/unsubscribe/token123", nil)
	w := httptest.NewRecorder()

	handler := subscription.NewHandler(svc)
	handler.UnsubscribeHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestUnsubscribeHandler_TokenWrongType(t *testing.T) {
	svc := &mockSubscriptionService{
		UnsubscribeFunc: func(token string) error {
			return subscription.ErrTokenWrongType
		},
	}

	req := httptest.NewRequest("GET", "/api/unsubscribe/token123", nil)
	w := httptest.NewRecorder()

	handler := subscription.NewHandler(svc)
	handler.UnsubscribeHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUnsubscribeHandler_UnsupportedMethod(t *testing.T) {
	svc := &mockSubscriptionService{}

	req := httptest.NewRequest("POST", "/api/unsubscribe/token123", nil)
	w := httptest.NewRecorder()

	handler := subscription.NewHandler(svc)
	handler.UnsubscribeHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Unsupported method") {
		t.Errorf("expected method error message, got: %s", w.Body.String())
	}
}

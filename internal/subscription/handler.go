package subscription

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

const (
	genericErrorMsg = "Something went wrong"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

var (
	ErrInvalidEmail     = errors.New("email parameter is invalid")
	ErrInvalidCity      = errors.New("city parameter is invalid")
	ErrInvalidFrequency = errors.New("frequency parameter is invalid")
)

var validFrequencies = map[string]struct{}{
	"hourly": {},
	"daily":  {},
}

type SubscriptionServiceInterface interface {
	Subscribe(email, city, frequency string) error
	Confirm(tokenValue string) error
	Unsubscribe(tokenValue string) error
}

type SubscriptionHandler struct {
	service SubscriptionServiceInterface
}

func NewHandler(svc SubscriptionServiceInterface) *SubscriptionHandler {
	return &SubscriptionHandler{service: svc}
}

type FormData struct {
	Email     string
	City      string
	Frequency string
}

func isValidFrequency(freq string) bool {
	_, ok := validFrequencies[freq]
	return ok
}

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func parseFormData(req *http.Request) (*FormData, error) {
	data := FormData{}

	if err := req.ParseForm(); err != nil {
		return nil, err
	}

	data.Email = req.FormValue("email")
	if data.Email == "" {
		return nil, ErrInvalidEmail
	}

	if !isValidEmail(data.Email) {
		return nil, ErrInvalidEmail
	}

	data.City = req.FormValue("city")
	if data.City == "" {
		return nil, ErrInvalidCity
	}

	data.Frequency = req.FormValue("frequency")
	if data.Frequency == "" {
		return nil, ErrInvalidFrequency
	}

	if !isValidFrequency(data.Frequency) {
		return nil, ErrInvalidFrequency
	}

	return &data, nil
}

func (h *SubscriptionHandler) SubscribeHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		errorMessage := fmt.Sprintf("Unsupported method %s", req.Method)
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	// Parse form data
	data, err := parseFormData(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	err = h.service.Subscribe(data.Email, data.City, data.Frequency)

	if err != nil {
		switch {
		case errors.Is(err, ErrUserAlreadyExists):
			http.Error(w, ErrUserAlreadyExists.Error(), http.StatusConflict)

		case errors.Is(err, ErrConfirmationMailError):
			log.Println(err.Error()) // We don't want to fail on confirmation mail error

		default:
			http.Error(w, genericErrorMsg, http.StatusInternalServerError)

		}

		log.Println(err.Error())

		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *SubscriptionHandler) ConfirmHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		errorMessage := fmt.Sprintf("Unsupported method %s", req.Method)
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	tokenValue := strings.TrimPrefix(req.URL.Path, "/api/confirm/")

	log.Printf("Token is: %s\n", tokenValue)

	err := h.service.Confirm(tokenValue)

	if err != nil {
		switch {
		case errors.Is(err, ErrTokenNotFound):
			http.Error(w, ErrTokenNotFound.Error(), http.StatusNotFound)

		case errors.Is(err, ErrTokenEmpty):
			http.Error(w, ErrTokenEmpty.Error(), http.StatusBadRequest)

		case errors.Is(err, ErrTokenWrongType):
			http.Error(w, ErrTokenWrongType.Error(), http.StatusBadRequest)

		default:
			http.Error(w, genericErrorMsg, http.StatusInternalServerError)
		}

		log.Println(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *SubscriptionHandler) UnsubscribeHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		errorMessage := fmt.Sprintf("Unsupported method %s", req.Method)
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	tokenValue := strings.TrimPrefix(req.URL.Path, "/api/unsubscribe/")

	log.Printf("Token is: %s\n", tokenValue)

	err := h.service.Unsubscribe(tokenValue)
	if err != nil {
		switch {
		case errors.Is(err, ErrTokenNotFound):
			http.Error(w, ErrTokenNotFound.Error(), http.StatusNotFound)

		case errors.Is(err, ErrTokenEmpty):
			http.Error(w, ErrTokenEmpty.Error(), http.StatusBadRequest)

		case errors.Is(err, ErrTokenWrongType):
			http.Error(w, ErrTokenWrongType.Error(), http.StatusBadRequest)

		default:
			http.Error(w, genericErrorMsg, http.StatusInternalServerError)
		}

		log.Println(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

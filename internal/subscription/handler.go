package subscription

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

const (
	genericErrorMsg = "Something went wrong"
)

type SubscriptionHandler struct {
	service *SubscriptionService
}

func NewHandler(svc *SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: svc}
}

var validFrequencies = map[string]struct{}{
	"hourly": {},
	"daily":  {},
}

func isValidFrequency(freq string) bool {
	_, ok := validFrequencies[freq]
	return ok
}

// TODO: write email validation
func isValidEmail(email string) bool {
	return email != ""
}

func (h *SubscriptionHandler) SubscribeHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		errorMessage := fmt.Sprintf("Unsupported method %s", req.Method)
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	// Parse form data
	if err := req.ParseForm(); err != nil {
		http.Error(w, genericErrorMsg, http.StatusInternalServerError)
		return
	}

	email := req.FormValue("email")
	if email == "" {
		http.Error(w, "\"email\" parameter is empty", http.StatusBadRequest)

		return
	}

	if !isValidEmail(email) {
		err := fmt.Sprintf("invalid \"email\" parameter \"%s\"", email)
		http.Error(w, err, http.StatusBadRequest)

		return
	}

	city := req.FormValue("city")
	if city == "" {
		http.Error(w, "\"city\" parameter is empty", http.StatusBadRequest)

		return
	}

	frequency := req.FormValue("frequency")
	if frequency == "" {
		http.Error(w, "\"frequency\" parameter is empty", http.StatusBadRequest)

		return
	}

	if !isValidFrequency(frequency) {
		err := fmt.Sprintf("invalid \"frequency\" parameter \"%s\"", frequency)
		http.Error(w, err, http.StatusBadRequest)

		return
	}

	err := h.service.Subscribe(email, city, frequency)

	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			http.Error(w, ErrUserAlreadyExists.Error(), http.StatusConflict)
		} else {
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

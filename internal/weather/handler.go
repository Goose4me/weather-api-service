package weather

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

const (
	GenericErrorMsg = "Something went wrong"
)

type WeatherHandler struct {
	service WeatherServiceInterface
}

func NewHandler(svc WeatherServiceInterface) *WeatherHandler {
	return &WeatherHandler{service: svc}
}

func (wh *WeatherHandler) Handler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Unsupported method", http.StatusBadRequest)
		return
	}

	query := req.URL.Query()

	city := query.Get("city")
	if city == "" {
		http.Error(w, "City parameter is empty", http.StatusBadRequest)
		return
	}

	weatherData, err := wh.service.GetWeather(city)
	if err != nil {
		if errors.Is(err, ErrCityNotFound) {
			http.Error(w, ErrCityNotFound.Error(), http.StatusNotFound)

		} else {
			http.Error(w, GenericErrorMsg, http.StatusInternalServerError)
		}

		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(weatherData); err != nil {
		log.Printf("Encoding error %s", err.Error())

		http.Error(w, "Encoding error", http.StatusInternalServerError)
	}
}

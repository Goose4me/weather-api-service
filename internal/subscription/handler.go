package subscription

import "net/http"

const (
	genericErrorMsg = "Something went wrong"
)

func SubscriptionHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "Unsupported method", http.StatusBadRequest)
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

}

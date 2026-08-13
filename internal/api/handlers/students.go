package handlers

import (
	"fmt"
	"net/http"
)

func StudentHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		fmt.Fprintf(w, "Hello Student GET Route")

	case http.MethodPost:
		fmt.Fprintf(w, "Hello Student POST Route")

	case http.MethodPut:
		fmt.Fprintf(w, "Hello Student PUT Route")

	case http.MethodPatch:
		fmt.Fprintf(w, "Hello Student PATCH Route")

	case http.MethodDelete:
		fmt.Fprintf(w, "Hello Student DELETE Route")

	default:
		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
	}
}

package handlers

import (
	"fmt"
	"net/http"
)

func ExecsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		fmt.Fprintf(w, "Hello Execs Route GET")

	case http.MethodPost:
		fmt.Fprintf(w, "Hello Execs Route POST")

	case http.MethodPut:
		fmt.Fprintf(w, "Hello Execs Route PUT")

	case http.MethodPatch:
		fmt.Fprintf(w, "Hello Execs Route PATCH")

	case http.MethodDelete:
		fmt.Fprintf(w, "Hello Execs Route DELETE")

	default:
		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
	}
}

package main

import (
	"fmt"
	"log"
	"net/http"
)

var PORT = "3000"

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello Root Router")
	})
	http.HandleFunc("/teachers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			fmt.Fprintf(w, "Hello Teacher GET Route")
		case http.MethodPost:
			fmt.Fprintf(w, "Hello Teacher POST route")
		case http.MethodPut:
			fmt.Fprintf(w, "Hello Teacher PUT route")
		case http.MethodPatch:
			fmt.Fprintf(w, "Hello Teacher PATCH route")
		case http.MethodDelete:
			fmt.Fprintf(w, "Hello Teacher DELETE route")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/students", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			fmt.Fprintf(w, "Hello Student GET Route")
		case http.MethodPost:
			fmt.Fprintf(w, "Hello Student POST route")
		case http.MethodPut:
			fmt.Fprintf(w, "Hello Student PUT route")
		case http.MethodPatch:
			fmt.Fprintf(w, "Hello Student PATCH route")
		case http.MethodDelete:
			fmt.Fprintf(w, "Hello Student DELETE route")
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/execs", func(w http.ResponseWriter, r *http.Request) {
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
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("Server is running on port", PORT)
	err := http.ListenAndServe(":"+PORT, nil)

	if err != nil {
		log.Fatal("Error starting the server", err)
		return
	}

}

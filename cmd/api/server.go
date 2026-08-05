package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

var PORT = "3000"

type User struct {
	Name string `json:"name"`
	Age  string `json:"age"`
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello Root Route")
}

func teacherHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fmt.Fprintf(w, "Hello Teacher GET Route")
		fmt.Println("URL", r.URL.Path)
		path := strings.TrimPrefix(r.URL.Path, "/teachers/")
		userID := strings.Trim(path, "/")

		if userID == "" {
			fmt.Fprintln(w, "Getting all teachers")
			return
		}

		fmt.Fprintf(w, "Getting teacher with ID: %s", userID)

	case http.MethodPost:

		fmt.Fprintf(w, "Hello Teacher POST route")
		//Form data
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		fmt.Println("Form Data:", r.Form)

		response := make(map[string]interface{})
		for key, value := range r.Form {
			response[key] = value[0]
		}
		fmt.Println("Response:", response)

		//Raw body data

		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Fatal("Error reading from the body", err)
			return
		}
		fmt.Println("Body:", string(body))

		//if you expect json data then unmarshal it into a struct
		var user User
		err = json.Unmarshal(body, &user)
		if err != nil {
			http.Error(w, "Error unmarshaling JSON", http.StatusBadRequest)
			return
		}
		fmt.Println("User:", user)
	case http.MethodPut:
		fmt.Fprintf(w, "Hello Teacher PUT route")
	case http.MethodPatch:
		fmt.Fprintf(w, "Hello Teacher PATCH route")
	case http.MethodDelete:
		fmt.Fprintf(w, "Hello Teacher DELETE route")
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func execsHandler(w http.ResponseWriter, r *http.Request) {
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
}

func studentHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fmt.Println("Query Params", r.URL.Query())
		paramsQuery := r.URL.Query()
		sortBy := paramsQuery.Get("sortBy")
		key := paramsQuery.Get("key")
		sortOrder := paramsQuery.Get("sortOrder")
		if sortOrder == "" {
			sortOrder = "DESC"
		}
		fmt.Println("Sort Order:", sortOrder)
		fmt.Println("Sort By:", sortBy)
		fmt.Println("Key:", key)
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
}

func main() {
	http.HandleFunc("/", rootHandler)

	http.HandleFunc("/teachers/", teacherHandler)

	http.HandleFunc("/students", studentHandler)

	http.HandleFunc("/execs", execsHandler)

	fmt.Println("Server is running on port", PORT)
	err := http.ListenAndServe(":"+PORT, nil)

	if err != nil {
		log.Fatal("Error starting the server", err)
		return
	}

}

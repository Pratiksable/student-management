package main

import (
	"fmt"
	"log"
	"net/http"
)

var PORT = "3000"

func main() {

	fmt.Println("Server is running on port", PORT)
	err := http.ListenAndServe(":"+PORT, nil)

	if err != nil {
		log.Fatal("Error starting the server", err)
		return
	}

}

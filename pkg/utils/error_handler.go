package utils

import (
	"fmt"
	"log"
	"os"
)

func ErrorHandler(err error, message string) error {
	errorLogger := log.New(os.Stderr, "Error: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger.Println(err, message)
	return fmt.Errorf("%s: %w", message, err)
}

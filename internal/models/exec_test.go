package models

import (
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
)

func TestExecJSONOmitsSecretsAndFormatsNullValues(t *testing.T) {
	exec := Exec{
		ID: 100, FirstName: "Pratik", LastName: "Sable",
		Email: "pratik@example.com", Username: "pratik",
		Password: "secret", PasswordResetCode: sql.NullString{String: "reset-secret", Valid: true},
		UserCreatedAt: sql.NullString{String: "2026-10-09 13:00:00", Valid: true},
		Role:          "admin",
	}

	data, err := json.Marshal(exec)
	if err != nil {
		t.Fatal(err)
	}
	jsonText := string(data)
	for _, secret := range []string{"secret", "password_reset_code", "\"password\""} {
		if strings.Contains(jsonText, secret) {
			t.Fatalf("response contains sensitive value %q: %s", secret, jsonText)
		}
	}
	if !strings.Contains(jsonText, `"password_changed_at":null`) ||
		!strings.Contains(jsonText, `"user_created_at":"2026-10-09 13:00:00"`) {
		t.Fatalf("nullable values were encoded incorrectly: %s", jsonText)
	}
}

func TestExecJSONStillAcceptsPasswordInput(t *testing.T) {
	var exec Exec
	if err := json.Unmarshal([]byte(`{"username":"pratik","password":"ChangeMe123!"}`), &exec); err != nil {
		t.Fatal(err)
	}
	if exec.Password != "ChangeMe123!" {
		t.Fatal("password input was not decoded")
	}
}

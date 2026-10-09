package sqlconnect

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"

	"github.com/Pratiksable/student-management/internal/models"
)

func TestExecQueries(t *testing.T) {
	exec := models.Exec{
		ID: 7, FirstName: "Priya", LastName: "Sharma", Email: "priya@example.com",
		Username: "priya", Password: "secret", PasswordChangedAt: sql.NullString{},
		UserCreatedAt: sql.NullString{}, PasswordResetCode: sql.NullString{},
		PasswordCodeExpiresAt: sql.NullString{}, Inactive: false, Role: "admin",
	}
	columns := "id, first_name, last_name, email, username, password, password_changed_at, user_created_at, password_reset_code, password_code_expires_at, inactive, role"
	editableColumns := "first_name, last_name, email, username, password, password_changed_at, user_created_at, password_reset_code, password_code_expires_at, inactive, role"
	values := []interface{}{
		"Priya", "Sharma", "priya@example.com", "priya", "secret",
		sql.NullString{}, sql.NullString{}, sql.NullString{}, sql.NullString{}, false, "admin",
	}

	for _, tc := range []struct {
		operation string
		query     string
		args      []interface{}
	}{
		{"insert", "INSERT INTO execs (" + editableColumns + ") VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", values},
		{"select", "SELECT " + columns + " FROM execs WHERE id = ?", []interface{}{7}},
		{"select_all", "SELECT " + columns + " FROM execs WHERE 1=1", nil},
		{"update", "UPDATE execs SET first_name = ?, last_name = ?, email = ?, username = ?, password = ?, password_changed_at = ?, user_created_at = ?, password_reset_code = ?, password_code_expires_at = ?, inactive = ?, role = ? WHERE id = ?", append(append([]interface{}{}, values...), 7)},
		{"delete", "DELETE FROM execs WHERE id = ?", []interface{}{7}},
	} {
		t.Run(tc.operation, func(t *testing.T) {
			query, args, err := generateExecsQuery(tc.operation, &exec)
			if err != nil || query != tc.query || !reflect.DeepEqual(args, tc.args) {
				t.Fatalf("got (%q, %v, %v), want (%q, %v)", query, args, err, tc.query, tc.args)
			}
			if strings.Count(query, "?") != len(args) {
				t.Fatal("placeholder count does not match arguments")
			}
		})
	}
}

func TestExecPatchAndInvalidInput(t *testing.T) {
	query, args, err := generateExecsQuery("patch", models.Exec{ID: 7}, map[string]interface{}{
		"first_name": "", "inactive": true, "id": float64(99),
	})
	if err != nil {
		t.Fatal(err)
	}
	if query != "UPDATE execs SET first_name = ?, inactive = ? WHERE id = ?" ||
		!reflect.DeepEqual(args, []interface{}{"", true, 7}) {
		t.Fatalf("unexpected patch: %s %v", query, args)
	}

	for _, patch := range []map[string]interface{}{
		nil, {}, {"id": float64(7)}, {"role": nil}, {"inactive": "true"}, {"not_a_field": "bad"}, {"password": "plaintext"},
	} {
		if _, _, err := generateExecsQuery("patch", models.Exec{ID: 7}, patch); err == nil {
			t.Fatalf("accepted invalid patch: %v", patch)
		}
	}
	for _, model := range []interface{}{nil, (*models.Exec)(nil), "exec"} {
		if _, _, err := generateExecsQuery("insert", model); err == nil {
			t.Fatalf("accepted invalid model: %v", model)
		}
	}
	if _, _, err := generateExecsQuery("unknown", models.Exec{}); err == nil {
		t.Fatal("accepted unsupported operation")
	}
}

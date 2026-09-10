package sqlconnect

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Pratiksable/student-management/internal/models"
)

func TestTeacherQueries(t *testing.T) {
	teacher := models.Teacher{ID: 7, FirstName: "Priya", LastName: "Sharma", Email: "priya@example.com", Class: "10A", Subject: "Math"}
	for _, tc := range []struct {
		operation string
		query     string
		args      []interface{}
	}{
		{"insert", "INSERT INTO teachers (first_name, last_name, email, class, subject) VALUES (?, ?, ?, ?, ?)", []interface{}{"Priya", "Sharma", "priya@example.com", "10A", "Math"}},
		{"select", "SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", []interface{}{7}},
		{"select_all", "SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE 1=1", nil},
		{"update", "UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?", []interface{}{"Priya", "Sharma", "priya@example.com", "10A", "Math", 7}},
		{"delete", "DELETE FROM teachers WHERE id = ?", []interface{}{7}},
	} {
		t.Run(tc.operation, func(t *testing.T) {
			query, args, err := generateTeacherQuery(tc.operation, &teacher)
			if err != nil || query != tc.query || !reflect.DeepEqual(args, tc.args) {
				t.Fatalf("got (%q, %v, %v), want (%q, %v)", query, args, err, tc.query, tc.args)
			}
			if strings.Count(query, "?") != len(args) {
				t.Fatal("placeholder count does not match arguments")
			}
		})
	}
}

func TestTeacherPatchPreservesOmittedFieldsAndID(t *testing.T) {
	query, args, err := generateTeacherQuery("patch", models.Teacher{ID: 7}, map[string]interface{}{
		"subject": "Science", "first_name": "", "id": float64(99),
	})
	if err != nil {
		t.Fatal(err)
	}
	if query != "UPDATE teachers SET first_name = ?, subject = ? WHERE id = ?" ||
		!reflect.DeepEqual(args, []interface{}{"", "Science", 7}) {
		t.Fatalf("unexpected patch: %s %v", query, args)
	}
}

func TestTeacherQueryRejectsInvalidInput(t *testing.T) {
	for _, model := range []interface{}{nil, (*models.Teacher)(nil), "teacher"} {
		if _, _, err := generateTeacherQuery("insert", model); err == nil {
			t.Fatalf("accepted invalid model: %v", model)
		}
	}
	for _, patch := range []map[string]interface{}{
		nil, {}, {"id": float64(7)}, {"subject": nil}, {"subject": 123}, {"subject = ?; DELETE FROM teachers": "bad"},
	} {
		if _, _, err := generateTeacherQuery("patch", models.Teacher{ID: 7}, patch); err == nil {
			t.Fatalf("accepted invalid patch: %v", patch)
		}
	}
	if _, _, err := generateTeacherQuery("patch", models.Teacher{}); err == nil {
		t.Fatal("accepted patch without field map")
	}
	if _, _, err := generateTeacherQuery("unknown", models.Teacher{}); err == nil {
		t.Fatal("accepted unsupported operation")
	}
}

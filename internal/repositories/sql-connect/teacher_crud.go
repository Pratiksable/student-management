package sqlconnect

import (
	"database/sql"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/Pratiksable/student-management/internal/models"
	"github.com/Pratiksable/student-management/pkg/utils"
)

func isValidSortOrder(order string) bool {
	return order == "asc" || order == "desc"
}

func isValidSortField(field string) bool {
	validField := map[string]bool{
		"first_name": true,
		"last_name":  true,
		"email":      true,
		"class":      true,
		"subject":    true,
	}

	return validField[field]
}

func generateTeacherQuery(
	operation string,
	model interface{},
	patches ...map[string]interface{},
) (string, []interface{}, error) {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("model must be a struct or non-nil struct pointer")
	}

	if operation == "patch" && len(patches) != 1 {
		return "", nil, fmt.Errorf("patch requires one field map")
	}

	t := v.Type()

	var columns, placeholders, assignments, selectColumns []string
	var values []interface{}
	var id interface{}
	hasID := false

	for i := 0; i < t.NumField(); i++ {
		column := strings.Split(t.Field(i).Tag.Get("db"), ",")[0]

		if column == "" || column == "-" || !v.Field(i).CanInterface() {
			continue
		}

		selectColumns = append(selectColumns, column)

		if column == "id" {
			id = v.Field(i).Interface()
			hasID = true
			continue
		}

		fieldValue := v.Field(i).Interface()
		if operation == "patch" {
			name := strings.Split(t.Field(i).Tag.Get("json"), ",")[0]
			value, present := patches[0][name]
			if !present {
				continue
			}
			if value == nil || reflect.TypeOf(value) != v.Field(i).Type() {
				return "", nil, fmt.Errorf("invalid value for %s", name)
			}
			fieldValue = value
		}

		columns = append(columns, column)
		placeholders = append(placeholders, "?")
		assignments = append(assignments, column+" = ?")
		values = append(values, fieldValue)
	}

	switch operation {
	case "insert":
		if len(columns) == 0 {
			return "", nil, fmt.Errorf("no columns to insert")
		}

		query := fmt.Sprintf(
			"INSERT INTO teachers (%s) VALUES (%s)",
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "),
		)
		return query, values, nil

	case "select_all":
		if len(selectColumns) == 0 {
			return "", nil, fmt.Errorf("no columns to select")
		}
		return "SELECT " + strings.Join(selectColumns, ", ") + " FROM teachers WHERE 1=1", nil, nil

	case "select":
		if !hasID {
			return "", nil, fmt.Errorf("model must have an id db tag")
		}

		query := fmt.Sprintf(
			"SELECT %s FROM teachers WHERE id = ?",
			strings.Join(selectColumns, ", "),
		)
		return query, []interface{}{id}, nil

	case "update", "patch":
		if operation == "patch" {
			for name := range patches[0] {
				if name == "id" {
					continue
				}
				known := false
				for i := 0; i < t.NumField(); i++ {
					field := t.Field(i)
					column := strings.Split(field.Tag.Get("db"), ",")[0]
					if strings.Split(field.Tag.Get("json"), ",")[0] == name && column != "" && column != "-" && column != "id" {
						known = true
						break
					}
				}
				if !known {
					return "", nil, fmt.Errorf("unknown teacher field: %s", name)
				}
			}
		}
		if !hasID || len(assignments) == 0 {
			return "", nil, fmt.Errorf("update requires an id and editable columns")
		}

		query := fmt.Sprintf(
			"UPDATE teachers SET %s WHERE id = ?",
			strings.Join(assignments, ", "),
		)
		return query, append(values, id), nil

	case "delete":
		if !hasID {
			return "", nil, fmt.Errorf("model must have an id db tag")
		}

		return "DELETE FROM teachers WHERE id = ?", []interface{}{id}, nil

	default:
		return "", nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

func addSorting(r *http.Request, query string) string {
	sortParams := r.URL.Query()["sortby"]
	if len(sortParams) > 0 {
		query += " ORDER BY"
		for i, param := range sortParams {
			parts := strings.Split(param, ":")
			if len(parts) != 2 {
				continue
			}

			field, order := parts[0], parts[1]
			if !isValidSortField(field) || !isValidSortOrder(order) {
				continue
			}
			if i > 0 {
				query += ","
			}

			query += " " + field + " " + order

		}
	}
	return query
}

func AddFilter(r *http.Request, query string, args []interface{}) (string, []interface{}) {
	params := map[string]string{
		"first_name": "first_name",
		"last_name":  "last_name",
		"email":      "email",
		"class":      "class",
		"subject":    "subject",
	}

	for param, dbField := range params {
		value := r.URL.Query().Get(param)
		if value != "" {
			query += " AND " + dbField + " =?"
			args = append(args, value)
		}
	}
	return query, args
}

func TeachersDBHandler(teachers []models.Teacher, r *http.Request) ([]models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	query, args, err := generateTeacherQuery("select_all", models.Teacher{})
	if err != nil {
		return nil, err
	}

	query, args = AddFilter(r, query, args)

	query = addSorting(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Database Query Error")
	}
	defer rows.Close()

	// teacherList := make([]models.Teacher, 0)
	for rows.Next() {
		var teacher models.Teacher
		err := rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error scanning database result")
		}
		teachers = append(teachers, teacher)
	}
	return teachers, rows.Err()
}

// Both *sql.DB and *sql.Tx support these operations, so bulk requests can
// use the same queries inside their transaction.
type teacherDB interface {
	Exec(string, ...interface{}) (sql.Result, error)
	QueryRow(string, ...interface{}) *sql.Row
}

func readTeacher(db teacherDB, id int) (models.Teacher, error) {
	query, args, err := generateTeacherQuery("select", models.Teacher{ID: id})
	if err != nil {
		return models.Teacher{}, err
	}
	var teacher models.Teacher
	err = db.QueryRow(query, args...).Scan(
		&teacher.ID, &teacher.FirstName, &teacher.LastName,
		&teacher.Email, &teacher.Class, &teacher.Subject,
	)
	return teacher, err
}

func TeacherDBHandler(id int) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Teacher{}, err
	}
	defer db.Close()
	return readTeacher(db, id)
}

func AddTeachersDBHandler(newTeachers []models.Teacher) ([]models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	addedTeachers := make([]models.Teacher, 0, len(newTeachers))
	for _, teacher := range newTeachers {
		query, args, err := generateTeacherQuery("insert", teacher)
		if err != nil {
			return nil, err
		}
		result, err := db.Exec(query, args...)
		if err != nil {
			return nil, err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}
		teacher.ID = int(id)
		addedTeachers = append(addedTeachers, teacher)
	}
	return addedTeachers, nil
}

func UpdateTeacherPutDBHandler(id int, teacher models.Teacher) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Teacher{}, err
	}
	defer db.Close()
	if _, err := readTeacher(db, id); err != nil {
		return models.Teacher{}, err
	}
	teacher.ID = id
	query, args, err := generateTeacherQuery("update", teacher)
	if err != nil {
		return models.Teacher{}, err
	}
	if _, err := db.Exec(query, args...); err != nil {
		return models.Teacher{}, err
	}
	return teacher, nil
}

func patchTeacher(db teacherDB, id int, update map[string]interface{}) (models.Teacher, error) {
	query, args, err := generateTeacherQuery("patch", models.Teacher{ID: id}, update)
	if err != nil {
		return models.Teacher{}, err
	}
	if _, err := readTeacher(db, id); err != nil {
		return models.Teacher{}, err
	}
	if _, err := db.Exec(query, args...); err != nil {
		return models.Teacher{}, err
	}
	return readTeacher(db, id)
}

func UpdateOneTeacherPatchDBHandler(id int, update map[string]interface{}) (models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Teacher{}, err
	}
	defer db.Close()
	return patchTeacher(db, id, update)
}

func UpdateTeachersPatchDBHandler(updates []map[string]interface{}) ([]models.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	teachers := make([]models.Teacher, 0, len(updates))
	for _, update := range updates {
		idValue, ok := update["id"].(float64)
		if !ok || idValue <= 0 || float64(int(idValue)) != idValue {
			return nil, fmt.Errorf("invalid teacher id in update")
		}
		teacher, err := patchTeacher(tx, int(idValue), update)
		if err != nil {
			return nil, err
		}
		teachers = append(teachers, teacher)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return teachers, nil
}

func deleteTeacher(db teacherDB, id int) error {
	query, args, err := generateTeacherQuery("delete", models.Teacher{ID: id})
	if err != nil {
		return err
	}
	result, err := db.Exec(query, args...)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func DeleteOneTeacherDBHandler(id int) error {
	db, err := ConnectDB()
	if err != nil {
		return err
	}
	defer db.Close()
	return deleteTeacher(db, id)
}

func DeleteTeachersDBHandler(ids []int) ([]int, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("no teacher ids supplied")
	}
	db, err := ConnectDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	for _, id := range ids {
		if err := deleteTeacher(tx, id); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return ids, nil
}

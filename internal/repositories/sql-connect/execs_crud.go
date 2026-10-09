package sqlconnect

import (
	"database/sql"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/Pratiksable/student-management/internal/auth"
	"github.com/Pratiksable/student-management/internal/models"
	"github.com/Pratiksable/student-management/pkg/utils"
)

func generateExecsQuery(operation string, model interface{}, patches ...map[string]interface{}) (string, []interface{}, error) {
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
		field := t.Field(i)
		column := strings.Split(field.Tag.Get("db"), ",")[0]
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
			name := strings.Split(field.Tag.Get("json"), ",")[0]
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
		query := fmt.Sprintf("INSERT INTO execs (%s) VALUES (%s)", strings.Join(columns, ", "), strings.Join(placeholders, ", "))
		return query, values, nil

	case "select_all":
		if len(selectColumns) == 0 {
			return "", nil, fmt.Errorf("no columns to select")
		}
		return "SELECT " + strings.Join(selectColumns, ", ") + " FROM execs WHERE 1=1", nil, nil

	case "select":
		if !hasID {
			return "", nil, fmt.Errorf("model must have an id db tag")
		}
		query := fmt.Sprintf("SELECT %s FROM execs WHERE id = ?", strings.Join(selectColumns, ", "))
		return query, []interface{}{id}, nil

	case "update", "patch":
		if operation == "patch" {
			for name := range patches[0] {
				if name == "id" {
					continue
				}
				if name == "password" {
					return "", nil, fmt.Errorf("password must be changed through the update-password endpoint")
				}
				known := false
				for i := 0; i < t.NumField(); i++ {
					field := t.Field(i)
					column := strings.Split(field.Tag.Get("db"), ",")[0]
					jsonName := strings.Split(field.Tag.Get("json"), ",")[0]
					if jsonName == name && column != "" && column != "-" && column != "id" {
						known = true
						break
					}
				}
				if !known {
					return "", nil, fmt.Errorf("unknown exec field: %s", name)
				}
			}
		}
		if !hasID || len(assignments) == 0 {
			return "", nil, fmt.Errorf("update requires an id and editable columns")
		}
		query := fmt.Sprintf("UPDATE execs SET %s WHERE id = ?", strings.Join(assignments, ", "))
		return query, append(values, id), nil

	case "delete":
		if !hasID {
			return "", nil, fmt.Errorf("model must have an id db tag")
		}
		return "DELETE FROM execs WHERE id = ?", []interface{}{id}, nil

	default:
		return "", nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

func ExecsDBHandler(execs []models.Exec, _ *http.Request) ([]models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query, args, err := generateExecsQuery("select_all", models.Exec{})
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var exec models.Exec
		if err := rows.Scan(
			&exec.ID, &exec.FirstName, &exec.LastName, &exec.Email,
			&exec.Username, &exec.Password, &exec.PasswordChangedAt,
			&exec.UserCreatedAt, &exec.PasswordResetCode,
			&exec.PasswordCodeExpiresAt, &exec.Inactive, &exec.Role,
		); err != nil {
			return nil, err
		}
		execs = append(execs, exec)
	}
	return execs, rows.Err()
}

func ExecDBHandler(exec []models.Exec, r *http.Request) ([]models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	query, args, err := generateExecsQuery("select_all", models.Exec{})
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error generating query for execs")
	}

	query, args = AddFilter(r, query, args)
	query = addSorting(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error executing query for execs")
	}
	defer rows.Close()

	for rows.Next() {
		var execs models.Exec
		err := scanExec(rows, &execs)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error scanning database result")
		}
		exec = append(exec, execs)
	}
	return exec, rows.Err()
}

type execsDB interface {
	Exec(string, ...interface{}) (sql.Result, error)
	QueryRow(string, ...interface{}) *sql.Row
}

type execScanner interface {
	Scan(...interface{}) error
}

func scanExec(row execScanner, exec *models.Exec) error {
	return row.Scan(
		&exec.ID, &exec.FirstName, &exec.LastName, &exec.Email,
		&exec.Username, &exec.Password, &exec.PasswordChangedAt,
		&exec.UserCreatedAt, &exec.PasswordResetCode,
		&exec.PasswordCodeExpiresAt, &exec.Inactive, &exec.Role,
	)
}

func readExec(db execsDB, id int) (models.Exec, error) {
	query, args, err := generateExecsQuery("select", models.Exec{ID: id})
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Error generating query for execs")
	}
	var execs models.Exec
	err = scanExec(db.QueryRow(query, args...), &execs)
	return execs, utils.ErrorHandler(err, "Error scanning database result")
}

func GetOneExecDBHandler(id int) (models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Error connecting to the DB")
	}
	defer db.Close()
	return readExec(db, id)
}

func AddExecsDBHandler(newExec []models.Exec) ([]models.Exec, error) {
	for _, exec := range newExec {
		if exec.Password == "" {
			return nil, fmt.Errorf("password is required")
		}
	}

	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to the DB")
	}
	defer db.Close()

	addedExecs := make([]models.Exec, 0, len(newExec))
	for _, exec := range newExec {

		encodedHash, err := auth.HashPassword(exec.Password)
		if err != nil {
			return nil, err
		}
		exec.Password = encodedHash

		query, args, err := generateExecsQuery("insert", exec)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error generating query for execs")
		}
		result, err := db.Exec(query, args...)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error executing query for execs")
		}
		id, err := result.LastInsertId()
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error getting last insert ID for execs")
		}
		exec.ID = int(id)
		addedExecs = append(addedExecs, exec)
	}
	return addedExecs, nil
}

func patchExecs(db execsDB, id int, update map[string]interface{}) (models.Exec, error) {
	query, args, err := generateExecsQuery("patch", models.Exec{ID: id}, update)
	if err != nil {
		return models.Exec{}, err
	}
	if _, err := readExec(db, id); err != nil {
		return models.Exec{}, err
	}
	if _, err := db.Exec(query, args...); err != nil {
		return models.Exec{}, err
	}
	return readExec(db, id)
}

func UpdateOneExecPatchDBHandler(id int, update map[string]interface{}) (models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Exec{}, err
	}
	defer db.Close()
	return patchExecs(db, id, update)
}

func UpdateExecsPatchDBHandler(updates []map[string]interface{}) ([]models.Exec, error) {
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

	execs := make([]models.Exec, 0, len(updates))
	for _, update := range updates {
		idValue, ok := update["id"].(float64)
		if !ok || idValue <= 0 || float64(int(idValue)) != idValue {
			return nil, fmt.Errorf("Invalid execs id in update")
		}
		exec, err := patchExecs(tx, int(idValue), update)
		if err != nil {
			return nil, err
		}
		execs = append(execs, exec)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return execs, nil
}

func deleteExec(db execsDB, id int) error {
	query, args, err := generateExecsQuery("delete", models.Exec{ID: id})
	if err != nil {
		return err
	}
	result, err := db.Exec(query, args...)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil

}

func DeleteOneExecsDBHandler(id int) error {
	db, err := ConnectDB()
	if err != nil {
		return err
	}
	defer db.Close()
	return deleteExec(db, id)
}

func GetExecByUsername(username string) (models.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Exec{}, err
	}
	defer db.Close()

	var exec models.Exec

	err = db.QueryRow(`
        SELECT id, first_name, last_name, email, username,
               password, inactive, role
        FROM execs
        WHERE username = ?
    `, username).Scan(
		&exec.ID,
		&exec.FirstName,
		&exec.LastName,
		&exec.Email,
		&exec.Username,
		&exec.Password,
		&exec.Inactive,
		&exec.Role,
	)

	return exec, err
}

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

func generateStudentsQuery(operation string, model interface{}, patches ...map[string]interface{}) (string, []interface{}, error) {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("models must be a struct or non-nil struct pointer")
	}
	if operation == "patch" && len(patches) != 1 {
		return "", nil, fmt.Errorf("patch requires one field map")
	}

	t := v.Type()
	var columns, placeholders, assignments, selectedColumns []string
	var values []interface{}
	var id interface{}
	hasID := false

	for i := 0; i < t.NumField(); i++ {
		column := strings.Split(t.Field(i).Tag.Get("db"), ",")[0]
		if column == "" || column == "-" || !v.Field(i).CanInterface() {
			continue
		}
		selectedColumns = append(selectedColumns, column)

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
				return "", nil, fmt.Errorf("Invalid Value for %s", name)
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
			return "", nil, fmt.Errorf("No Columns to insert")
		}
		query := fmt.Sprintf("INSERT INTO students (%s) VALUES (%s)", strings.Join(columns, ", "), strings.Join(placeholders, ", "))
		return query, values, nil
	case "select_all":
		if len(selectedColumns) == 0 {
			return "", nil, fmt.Errorf("no columns to select")
		}
		return "SELECT " + strings.Join(selectedColumns, ", ") + " FROM students WHERE 1=1", nil, nil
	case "select":
		if !hasID {
			return "", nil, fmt.Errorf("model must have an id db tag")
		}
		query := fmt.Sprintf("SELECT %s FROM students WHERE id = ?", strings.Join(selectedColumns, ", "))
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
					return "", nil, fmt.Errorf("Unknown student field: %s", name)
				}
			}
		}
		if !hasID || len(assignments) == 0 {
			return "", nil, fmt.Errorf("update require an id and editable columns")
		}
		query := fmt.Sprintf("UPDATE students SET %s WHERE id = ?", strings.Join(assignments, ", "))
		return query, append(values, id), nil
	case "delete":
		if !hasID {
			return "", nil, fmt.Errorf("model must have an id db tag")
		}
		return "DELETE FROM students WHERE id=?", []interface{}{id}, nil
	default:
		return "", nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

func StudentsDBHandler(students []models.Student, r *http.Request) ([]models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error Connecting to the Database")
	}
	defer db.Close()

	query, args, err := generateStudentsQuery("select_all", models.Student{})
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error in the generate query from select all")
	}
	query, args = AddFilter(r, query, args)
	query = addSorting(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Database query Error")
	}
	defer rows.Close()

	for rows.Next() {
		var student models.Student
		err := rows.Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error scanning database result")
		}
		students = append(students, student)
	}

	return students, rows.Err()

}

type studentDB interface {
	Exec(string, ...interface{}) (sql.Result, error)
	QueryRow(string, ...interface{}) *sql.Row
}

func readStudent(db studentDB, id int) (models.Student, error) {
	query, args, err := generateStudentsQuery("select", models.Student{ID: id})
	if err != nil {
		return models.Student{}, err
	}
	var student models.Student
	err = db.QueryRow(query, args...).Scan(
		&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class,
	)
	return student, err

}

func StudentDBHandler(id int) (models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Student{}, err
	}
	defer db.Close()
	return readStudent(db, id)
}

func AddStudentDBHandler(newStudent []models.Student) ([]models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	addedStudents := make([]models.Student, 0, len(newStudent))
	for _, student := range newStudent {
		query, args, err := generateStudentsQuery("insert", student)
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
		student.ID = int(id)
		addedStudents = append(addedStudents, student)

	}
	return addedStudents, nil
}

func patchStudent(db studentDB, id int, update map[string]interface{}) (models.Student, error) {
	query, args, err := generateStudentsQuery("patch", models.Student{ID: id}, update)
	if err != nil {
		return models.Student{}, err
	}
	if _, err := readStudent(db, id); err != nil {
		return models.Student{}, err
	}
	if _, err := db.Exec(query, args...); err != nil {
		return models.Student{}, err
	}
	return readStudent(db, id)
}

func UpdateOneStudentPatchDBHandler(id int, update map[string]interface{}) (models.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.Student{}, err
	}
	defer db.Close()
	return patchStudent(db, id, update)
}

func UpdateStudentsPatchDBHandler(updates []map[string]interface{}) ([]models.Student, error) {
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
	students := make([]models.Student, 0, len(updates))
	for _, update := range updates {
		idValue, ok := update["id"].(float64)
		if !ok || idValue <= 0 || float64(int(idValue)) != idValue {
			return nil, fmt.Errorf("Invalid student id in update")
		}
		student, err := patchStudent(tx, int(idValue), update)
		if err != nil {
			return nil, err
		}
		students = append(students, student)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return students, nil
}

func deleteStudent(db studentDB, id int) error {
	query, args, err := generateStudentsQuery("delete", models.Student{ID: id})
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

func DeleteOneStudentHandler(id int) error {
	db, err := ConnectDB()
	if err != nil {
		return err
	}
	defer db.Close()
	return deleteStudent(db, id)
}

func DeleteStudentsDBHandler(ids []int) ([]int, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("No student ids supplied")
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
		if err := deleteStudent(tx, id); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return ids, nil
}

func StudentRiskDBHandler(studentID int) (models.StudentRisk, error) {
	db, err := ConnectDB()
	if err != nil {
		return models.StudentRisk{}, err
	}
	defer db.Close()

	var attendancePercentage float64
	var attendanceCount int

	var averageMarks float64
	var marksCount int

	var missingAssignments int

	attendanceQuery := `
		SELECT
			COUNT(*) AS attendance_count,
			COALESCE(
				SUM(
					CASE
						WHEN status = 'present' THEN 1
						ELSE 0
					END
				) * 100.0 / NULLIF(COUNT(*), 0),
				0
			) AS attendance_percentage
		FROM attendance
		WHERE student_id = ?
	`

	err = db.QueryRow(
		attendanceQuery,
		studentID,
	).Scan(
		&attendanceCount,
		&attendancePercentage,
	)

	if err != nil {
		return models.StudentRisk{}, err
	}

	marksQuery := `
		SELECT
			COUNT(*) AS marks_count,
			COALESCE(
				AVG((score / max_score) * 100),
				0
			) AS average_marks
		FROM marks
		WHERE student_id = ?
	`

	err = db.QueryRow(
		marksQuery,
		studentID,
	).Scan(
		&marksCount,
		&averageMarks,
	)

	if err != nil {
		return models.StudentRisk{}, err
	}

	assignmentQuery := `
		SELECT COUNT(*)
		FROM assignments
		WHERE student_id = ?
		AND status = 'missing'
	`

	err = db.QueryRow(
		assignmentQuery,
		studentID,
	).Scan(&missingAssignments)

	if err != nil {
		return models.StudentRisk{}, err
	}

	performanceQuery := `
		SELECT
			(score / max_score) * 100.0 AS percentage
		FROM marks
		WHERE student_id = ?
		ORDER BY exam_date DESC
		LIMIT 6
	`

	rows, err := db.Query(
		performanceQuery,
		studentID,
	)

	if err != nil {
		return models.StudentRisk{}, err
	}
	defer rows.Close()

	var percentages []float64

	for rows.Next() {

		var percentage float64

		err := rows.Scan(&percentage)
		if err != nil {
			return models.StudentRisk{}, err
		}

		percentages = append(
			percentages,
			percentage,
		)
	}

	if err := rows.Err(); err != nil {
		return models.StudentRisk{}, err
	}
	performanceTrend := "insufficient_data"

	if len(percentages) >= 6 {

		recentAverage := (percentages[0] +
			percentages[1] +
			percentages[2]) / 3

		previousAverage := (percentages[3] +
			percentages[4] +
			percentages[5]) / 3

		difference := recentAverage - previousAverage

		if difference <= -10 {

			performanceTrend = "declining"

		} else if difference >= 10 {

			performanceTrend = "improving"

		} else {

			performanceTrend = "stable"
		}
	}

	riskScore := 0
	reasons := []string{}

	if attendanceCount > 0 {

		if attendancePercentage < 75 {

			riskScore += 30

			reasons = append(
				reasons,
				"Attendance below 75%",
			)
		}

		if attendancePercentage < 60 {

			riskScore += 15

			reasons = append(
				reasons,
				"Attendance critically low",
			)
		}
	}

	if marksCount > 0 {

		if averageMarks < 50 {

			riskScore += 25

			reasons = append(
				reasons,
				"Average marks below 50%",
			)
		}

		if averageMarks < 35 {

			riskScore += 15

			reasons = append(
				reasons,
				"Academic performance critically low",
			)
		}
	}

	if performanceTrend == "declining" {

		riskScore += 15

		reasons = append(
			reasons,
			"Recent academic performance is declining",
		)
	}

	switch {

	case missingAssignments >= 3:

		riskScore += 20

		reasons = append(
			reasons,
			"3 or more assignments missing",
		)

	case missingAssignments == 2:

		riskScore += 10

		reasons = append(
			reasons,
			"2 assignments missing",
		)

	case missingAssignments == 1:

		riskScore += 5

		reasons = append(
			reasons,
			"1 assignment missing",
		)
	}

	if riskScore > 100 {
		riskScore = 100
	}

	riskLevel := "LOW"

	if riskScore > 60 {

		riskLevel = "HIGH"

	} else if riskScore > 30 {

		riskLevel = "MEDIUM"
	}

	if attendanceCount == 0 && marksCount == 0 {

		riskLevel = "INSUFFICIENT_DATA"

		reasons = append(
			reasons,
			"Insufficient student data to calculate risk",
		)
	}

	return models.StudentRisk{
		StudentID:            studentID,
		RiskScore:            riskScore,
		RiskLevel:            riskLevel,
		AttendancePercentage: attendancePercentage,
		AverageMarks:         averageMarks,
		MissingAssignments:   missingAssignments,
		PerformanceTrend:     performanceTrend,
		Reasons:              reasons,
	}, nil
}

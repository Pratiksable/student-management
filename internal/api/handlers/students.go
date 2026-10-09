package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/Pratiksable/student-management/internal/models"
	sqlconnect "github.com/Pratiksable/student-management/internal/repositories/sql-connect"
)

func StudentHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		fmt.Fprintf(w, "Hello Student GET Route")

	case http.MethodPost:
		fmt.Fprintf(w, "Hello Student POST Route")

	case http.MethodPut:
		fmt.Fprintf(w, "Hello Student PUT Route")

	case http.MethodPatch:
		fmt.Fprintf(w, "Hello Student PATCH Route")

	case http.MethodDelete:
		fmt.Fprintf(w, "Hello Student DELETE Route")

	default:
		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
	}
}

func GetMultipleStudentHandler(w http.ResponseWriter, r *http.Request) {
	var students []models.Student
	students, err := sqlconnect.StudentsDBHandler(students, r)
	if err != nil {
		return
	}

	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Student `json:"data"`
	}{
		Status: "success",
		Count:  len(students),
		Data:   students,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func GetOneStudentHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid student id ", http.StatusBadRequest)
		return
	}
	student, err := sqlconnect.StudentDBHandler(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(student)
}

func AddStudentHandler(w http.ResponseWriter, r *http.Request) {
	var newStudent []models.Student
	var rawStudents []map[string]json.RawMessage

	err := json.NewDecoder(r.Body).Decode(&rawStudents)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	val := reflect.TypeOf(models.Student{})
	allowedFields := make(map[string]struct{})
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		key := strings.Split(field.Tag.Get("json"), ",")[0]
		allowedFields[key] = struct{}{}
	}
	for _, student := range rawStudents {
		for key := range student {
			if _, ok := allowedFields[key]; !ok {
				http.Error(w, "Unacceptable Fields found in the request. Only use allowed fields", http.StatusBadRequest)
				return
			}
		}
	}
	data, err := json.Marshal(rawStudents)
	if err != nil {
		http.Error(w, "Invalid Request body", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(data, &newStudent); err != nil {

		http.Error(w, "Invalid Request body", http.StatusBadRequest)
		return
	}
	for _, student := range newStudent {
		val := reflect.ValueOf(student)
		for i := 0; i < val.NumField(); i++ {
			if val.Field(i).Kind() == reflect.String && val.Field(i).String() == "" {
				http.Error(w, "All fields are required", http.StatusBadRequest)
				return
			}
		}
	}
	addedStudents, err := sqlconnect.AddStudentDBHandler(newStudent)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Stauts string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Student `json:"data"`
	}{
		Stauts: "success",
		Count:  len(addedStudents),
		Data:   addedStudents,
	}
	json.NewEncoder(w).Encode(response)

}

func UpdateOneStudentPatchHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		return
	}

	var update map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		log.Println(err)
		return
	}
	existingStudent, err := sqlconnect.UpdateOneStudentPatchDBHandler(id, update)
	if err != nil {
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingStudent)
}

func UpdateStudentsPatchHandler(w http.ResponseWriter, r *http.Request) {
	var updates []map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&updates)

	if err != nil {
		http.Error(w, "Invalid Response Body", http.StatusBadRequest)
		return
	}

	updateFromDB, err := sqlconnect.UpdateStudentsPatchDBHandler(updates)
	if err != nil {
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Status":  "success",
		"Message": "Students Updated Successfully",
		"data":    updateFromDB,
	})
}

func DeleteStudentHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Student ID", http.StatusBadRequest)
		return
	}
	err = sqlconnect.DeleteOneStudentHandler(id)
	if err != nil {
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := struct {
		Status string `json:"status"`
		ID     int    `json:"id"`
	}{
		Status: "Teacher Successfully deleted",
		ID:     id,
	}

	json.NewEncoder(w).Encode(response)
}

func DeleteStudentsHandler(w http.ResponseWriter, r *http.Request) {
	var ids []int

	err := json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		http.Error(w, "Unable to decode the json", http.StatusInternalServerError)
		return
	}

	deleteIds, err := sqlconnect.DeleteStudentsDBHandler(ids)
	if err != nil {
		log.Println(err)
		return
	}
	response := struct {
		Status     string `json:"status"`
		Count      int    `json:"count"`
		DeletedIDs []int  `json:"deleted_ids"`
	}{
		Status:     "Teachers succesfully deleted",
		Count:      len(deleteIds),
		DeletedIDs: deleteIds,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func GetStudentRiskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid student ID", http.StatusInternalServerError)
		return
	}
	risk, err := sqlconnect.StudentRiskDBHandler(id)
	if err != nil {
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(risk)
}

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

func TeacherHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		GetTeachersHandler(w, r)

	case http.MethodPost:
		AddTeacherHandler(w, r)

	case http.MethodPut:
		UpdateTeacherHandler(w, r)

	case http.MethodPatch:
		UpdateOneTeacherPatchHandler(w, r)

	case http.MethodDelete:
		DeleteTeacherHandler(w, r)

	default:
		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
	}
}

func GetTeachersHandler(w http.ResponseWriter, r *http.Request) {

	var teachers []models.Teacher
	teachers, err := sqlconnect.TeachersDBHandler(teachers, r)
	if err != nil {
		return
	}

	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Teacher `json:"data"`
	}{
		Status: "success",
		Count:  len(teachers),
		Data:   teachers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	return
}

func GetOneTeacherHandler(w http.ResponseWriter, r *http.Request) {

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	teacher, err := sqlconnect.TeacherDBHandler(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teacher)
}

func AddTeacherHandler(w http.ResponseWriter, r *http.Request) {

	var newTeachers []models.Teacher
	var rawTeachers []map[string]json.RawMessage

	err := json.NewDecoder(r.Body).Decode(&rawTeachers)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	val := reflect.TypeOf(models.Teacher{})
	allowedFields := make(map[string]struct{})
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		key := strings.Split(field.Tag.Get("json"), ",")[0]
		allowedFields[key] = struct{}{}
	}

	for _, teacher := range rawTeachers {
		for key := range teacher {
			if _, ok := allowedFields[key]; !ok {
				http.Error(w, "Unacceptable fields found in request. Only use allowed fields", http.StatusBadRequest)
				return
			}
		}
	}
	data, err := json.Marshal(rawTeachers)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(data, &newTeachers); err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}
	for _, teacher := range newTeachers {
		val := reflect.ValueOf(teacher)
		for i := 0; i < val.NumField(); i++ {
			if val.Field(i).Kind() == reflect.String && val.Field(i).String() == "" {
				http.Error(w, "All fields are required", http.StatusBadRequest)
				return
			}
		}
	}

	addedTeacher, err := sqlconnect.AddTeachersDBHandler(newTeachers)
	if err != nil {
		fmt.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Teacher `json:"data"`
	}{
		Status: "Success",
		Count:  len(addedTeacher),
		Data:   addedTeacher,
	}
	json.NewEncoder(w).Encode(response)
}

// /Put

func UpdateTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		return
	}
	var updatedTeacher models.Teacher
	err = json.NewDecoder(r.Body).Decode(&updatedTeacher)
	if err != nil {
		log.Println(err)
		return
	}
	updatedTeacherFromDB, err := sqlconnect.UpdateTeacherPutDBHandler(id, updatedTeacher)
	if err != nil {
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTeacherFromDB)

}

func UpdateOneTeacherPatchHandler(w http.ResponseWriter, r *http.Request) {
	idstr := r.PathValue("id")
	id, err := strconv.Atoi(idstr)
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
	existingTeacher, err := sqlconnect.UpdateOneTeacherPatchDBHandler(id, update)
	if err != nil {
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingTeacher)
}

func UpdateTeachersPatchHandler(w http.ResponseWriter, r *http.Request) {

	var updates []map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid Request payload", http.StatusBadRequest)
		return
	}

	updatesFromDB, err := sqlconnect.UpdateTeachersPatchDBHandler(updates)
	if err != nil {
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Teachers updated successfully",
		"data":    updatesFromDB,
	})

}

func DeleteTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	err = sqlconnect.DeleteOneTeacherDBHandler(id)
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
		Status: "Teacher successfully deleted",
		ID:     id,
	}

	json.NewEncoder(w).Encode(response)
}

func DeleteTeachersHandler(w http.ResponseWriter, r *http.Request) {

	var ids []int

	err := json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		http.Error(w, "Unable to decode the json", http.StatusInternalServerError)
		return
	}

	deletedIds, err := sqlconnect.DeleteTeachersDBHandler(ids)
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
		Count:      len(deletedIds),
		DeletedIDs: deletedIds,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func GetStudentsByTeacherID(w http.ResponseWriter, r *http.Request) {
	teacheridstr := r.PathValue("id")
	teacherId, err := strconv.Atoi(teacheridstr)
	if err != nil {
		http.Error(w, "Invalid teacher ID", http.StatusBadRequest)
		return
	}
	var students []models.Student

	students, shouldReturn := sqlconnect.GetStudentsByTeacherIDDBHandler(w, teacherId, students)
	if shouldReturn {
		return
	}

	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Student `json:"data"`
	}{
		Status: "success",
		Count:  len((students)),
		Data:   students,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(response)
}

func GetStudentsCountByTeacherID(w http.ResponseWriter, r *http.Request) {
	teacherStrId := r.PathValue("id")
	teacherID, err := strconv.Atoi(teacherStrId)
	if err != nil {
		return
	}

	count, shouldReturn := sqlconnect.GetStudentsCountByTeacherIDDBHandler(teacherID, w)
	if shouldReturn {
		return
	}

	response := struct {
		Status       string `json:"status"`
		TeacherID    int    `json:"teacher_id"`
		StudentCount int    `json:"student_count"`
	}{
		Status:       "success",
		TeacherID:    teacherID,
		StudentCount: count,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

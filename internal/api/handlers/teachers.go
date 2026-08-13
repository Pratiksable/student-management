package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/Pratiksable/student-management/internal/models"
)

var (
	teacher = make(map[int]models.Teacher)
	mutex   = &sync.Mutex{}
	nextID  = 1
)

func init() {
	teacher[nextID] = models.Teacher{
		ID:        nextID,
		FirstName: "John",
		LastName:  "Doe",
		Class:     "9A",
		Subject:   "Math",
	}
	nextID++
	teacher[nextID] = models.Teacher{
		ID:        nextID,
		FirstName: "Alex",
		LastName:  "Bane",
		Class:     "10A",
		Subject:   "Algebra",
	}
	nextID++
	teacher[nextID] = models.Teacher{
		ID:        nextID,
		FirstName: "Atul",
		LastName:  "Kumar",
		Class:     "1A",
		Subject:   "Hindi",
	}
	nextID++
}

func TeacherHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		getTeachersHandler(w, r)

	case http.MethodPost:
		addTeacherHandler(w, r)

	case http.MethodPut:
		fmt.Fprintf(w, "Hello Teacher PUT Route")

	case http.MethodPatch:
		fmt.Fprintf(w, "Hello Teacher PATCH Route")

	case http.MethodDelete:
		fmt.Fprintf(w, "Hello Teacher DELETE Route")

	default:
		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
	}
}

func getTeachersHandler(w http.ResponseWriter, r *http.Request) {

	path := strings.TrimPrefix(r.URL.Path, "/teachers")
	idStr := strings.Trim(path, "/")

	// GET /teachers
	if idStr == "" {

		firstName := r.URL.Query().Get("first_name")
		lastName := r.URL.Query().Get("last_name")

		teacherList := make([]models.Teacher, 0, len(teacher))

		for _, t := range teacher {
			if (firstName == "" || t.FirstName == firstName) &&
				(lastName == "" || t.LastName == lastName) {

				teacherList = append(teacherList, t)
			}
		}

		response := struct {
			Status string           `json:"status"`
			Count  int              `json:"count"`
			Data   []models.Teacher `json:"data"`
		}{
			Status: "success",
			Count:  len(teacherList),
			Data:   teacherList,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

		return
	}

	// GET /teachers/1
	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	t, exists := teacher[id]

	if !exists {
		http.Error(w, "Teacher not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}
func addTeacherHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	var newTeachers []models.Teacher

	err := json.NewDecoder(r.Body).Decode(&newTeachers)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}
	addedToTeacher := make([]models.Teacher, len(newTeachers))

	for i, newTeachers := range newTeachers {
		newTeachers.ID = nextID
		teacher[nextID] = newTeachers
		addedToTeacher[i] = newTeachers
		nextID++
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Teacher `json:"data"`
	}{
		Status: "Success",
		Count:  len(addedToTeacher),
		Data:   addedToTeacher,
	}
	json.NewEncoder(w).Encode(response)
}

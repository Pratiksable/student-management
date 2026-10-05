package router

import (
	"net/http"

	"github.com/Pratiksable/student-management/internal/api/handlers"
)

func Router() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handlers.RootHandler)
	mux.HandleFunc("GET /teachers", handlers.GetTeachersHandler)
	mux.HandleFunc("POST /teachers", handlers.AddTeacherHandler)
	mux.HandleFunc("PATCH /teachers", handlers.UpdateTeachersPatchHandler)
	mux.HandleFunc("DELETE /teachers", handlers.DeleteTeachersHandler)

	mux.HandleFunc("GET /teachers/{id}", handlers.GetOneTeacherHandler)
	mux.HandleFunc("PATCH /teachers/{id}", handlers.UpdateOneTeacherPatchHandler)
	mux.HandleFunc("DELETE /teachers/{id}", handlers.DeleteTeacherHandler)

	mux.HandleFunc("GET /teachers/{id}/students", handlers.GetStudentsByTeacherID)
	mux.HandleFunc("GET /teachers/{id}/count", handlers.GetStudentsCountByTeacherID)
	mux.HandleFunc("GET /students/{id}/risk", handlers.GetStudentRiskHandler)

	mux.HandleFunc("GET /students", handlers.GetMultipleStudentHandler)
	mux.HandleFunc("POST /students", handlers.AddStudentHandler)
	mux.HandleFunc("PATCH /students", handlers.UpdateStudentsPatchHandler)
	mux.HandleFunc("DELETE /students", handlers.DeleteStudentsHandler)

	mux.HandleFunc("GET /students/{id}", handlers.GetOneStudentHandler)
	mux.HandleFunc("PATCH /students/{id}", handlers.UpdateOneStudentPatchHandler)
	mux.HandleFunc("DELETE /students/{id}", handlers.DeleteStudentHandler)
	// mux.HandleFunc("/execs", handlers.ExecsHandler)

	return mux
}

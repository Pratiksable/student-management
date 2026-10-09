package router

import (
	"net/http"

	"github.com/Pratiksable/student-management/internal/api/handlers"
	"github.com/Pratiksable/student-management/internal/api/middleware"
	"github.com/Pratiksable/student-management/internal/auth"
)

func studentsRouter() *http.ServeMux {

	mux := http.NewServeMux()
	requireAuth := middleware.RequireAuthentication(auth.DefaultSessionStore)

	mux.Handle("GET /students", requireAuth(http.HandlerFunc(handlers.GetMultipleStudentHandler)))
	mux.Handle("POST /students", requireAuth(http.HandlerFunc(handlers.AddStudentHandler)))
	mux.Handle("PATCH /students", requireAuth(http.HandlerFunc(handlers.UpdateStudentsPatchHandler)))
	mux.Handle("DELETE /students", requireAuth(http.HandlerFunc(handlers.DeleteStudentsHandler)))

	mux.Handle("GET /students/{id}", requireAuth(http.HandlerFunc(handlers.GetOneStudentHandler)))
	mux.Handle("PATCH /students/{id}", requireAuth(http.HandlerFunc(handlers.UpdateOneStudentPatchHandler)))
	mux.Handle("DELETE /students/{id}", requireAuth(http.HandlerFunc(handlers.DeleteStudentHandler)))
	return mux
}

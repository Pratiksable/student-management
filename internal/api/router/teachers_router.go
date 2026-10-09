package router

import (
	"net/http"

	"github.com/Pratiksable/student-management/internal/api/handlers"
	"github.com/Pratiksable/student-management/internal/api/middleware"
	"github.com/Pratiksable/student-management/internal/auth"
)

func TeachersRouter() *http.ServeMux {
	mux := http.NewServeMux()
	requireAuth := middleware.RequireAuthentication(auth.DefaultSessionStore)

	mux.Handle("GET /teachers", requireAuth(http.HandlerFunc(handlers.GetTeachersHandler)))
	mux.Handle("POST /teachers", requireAuth(http.HandlerFunc(handlers.AddTeacherHandler)))
	mux.Handle("PATCH /teachers", requireAuth(http.HandlerFunc(handlers.UpdateTeachersPatchHandler)))
	mux.Handle("DELETE /teachers", requireAuth(http.HandlerFunc(handlers.DeleteTeachersHandler)))

	mux.Handle("GET /teachers/{id}", requireAuth(http.HandlerFunc(handlers.GetOneTeacherHandler)))
	mux.Handle("PATCH /teachers/{id}", requireAuth(http.HandlerFunc(handlers.UpdateOneTeacherPatchHandler)))
	mux.Handle("DELETE /teachers/{id}", requireAuth(http.HandlerFunc(handlers.DeleteTeacherHandler)))

	mux.Handle("GET /teachers/{id}/students", requireAuth(http.HandlerFunc(handlers.GetStudentsByTeacherID)))
	mux.Handle("GET /teachers/{id}/count", requireAuth(http.HandlerFunc(handlers.GetStudentsCountByTeacherID)))
	mux.Handle("GET /students/{id}/risk", requireAuth(http.HandlerFunc(handlers.GetStudentRiskHandler)))
	return mux
}

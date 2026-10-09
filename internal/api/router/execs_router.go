package router

import (
	"net/http"

	"github.com/Pratiksable/student-management/internal/api/handlers"
	"github.com/Pratiksable/student-management/internal/api/middleware"
	"github.com/Pratiksable/student-management/internal/auth"
)

func ExecsRouter() *http.ServeMux {
	mux := http.NewServeMux()
	requireAuth := middleware.RequireAuthentication(auth.DefaultSessionStore)
	mux.Handle("GET /execs", requireAuth(http.HandlerFunc(handlers.GetExecHandler)))
	mux.Handle("POST /execs", requireAuth(http.HandlerFunc(handlers.AddManyExecHandler)))
	mux.Handle("PATCH /execs", requireAuth(http.HandlerFunc(handlers.UpdateExecsPatchHandler)))

	mux.Handle("GET /execs/{id}", requireAuth(http.HandlerFunc(handlers.GetOneExecHandler)))
	mux.Handle("PATCH /execs/{id}", requireAuth(http.HandlerFunc(handlers.UpdateOneExecPatchHandler)))
	mux.Handle("DELETE /execs/{id}", requireAuth(http.HandlerFunc(handlers.DeleteExecsHandler)))
	mux.Handle("POST /execs/{id}/updatepassword", requireAuth(http.HandlerFunc(handlers.ExecsHandler)))

	mux.HandleFunc("POST /execs/login", handlers.LoginHandler)
	mux.Handle("POST /execs/logout", requireAuth(http.HandlerFunc(handlers.LogoutHandler)))
	mux.HandleFunc("POST /execs/forgotpassword", handlers.ExecsHandler)
	mux.HandleFunc("POST /execs/resetpassword/reset/{resetcode}", handlers.ExecsHandler)

	return mux
}

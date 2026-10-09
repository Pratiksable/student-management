package router

import (
	"net/http"
)

func MainRouter() *http.ServeMux {

	teachersRouter := TeachersRouter()
	studentsRouter := studentsRouter()
	execsRouter := ExecsRouter()

	studentsRouter.Handle("/", execsRouter)
	teachersRouter.Handle("/", studentsRouter)

	return teachersRouter
}

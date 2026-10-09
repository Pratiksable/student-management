package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/Pratiksable/student-management/internal/auth"
	"github.com/Pratiksable/student-management/internal/models"
	sqlconnect "github.com/Pratiksable/student-management/internal/repositories/sql-connect"
)

func writeExecError(w http.ResponseWriter, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Exec not found", http.StatusNotFound)
		return
	}
	http.Error(w, "Unable to process exec request", http.StatusInternalServerError)
}

func ExecsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		fmt.Fprintf(w, "Hello Execs Route GET")

	case http.MethodPost:
		fmt.Fprintf(w, "Hello Execs Route POST")

	case http.MethodPut:
		fmt.Fprintf(w, "Hello Execs Route PUT")

	case http.MethodPatch:
		fmt.Fprintf(w, "Hello Execs Route PATCH")

	case http.MethodDelete:
		fmt.Fprintf(w, "Hello Execs Route DELETE")

	default:
		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
	}
}

func GetExecHandler(w http.ResponseWriter, r *http.Request) {
	execs := make([]models.Exec, 0)

	execs, err := sqlconnect.ExecDBHandler(execs, r)
	if err != nil {
		writeExecError(w, err)
		return
	}

	response := struct {
		Status string        `json:"status"`
		Count  int           `json:"count"`
		Data   []models.Exec `json:"data"`
	}{
		Status: "success",
		Count:  len(execs),
		Data:   execs,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(response)

}

func GetOneExecHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid exec ID", http.StatusBadRequest)
		return
	}
	exec, err := sqlconnect.GetOneExecDBHandler(id)
	if err != nil {
		writeExecError(w, err)
		return
	}

	response := struct {
		Status string      `json:"status"`
		Count  int         `json:"count"`
		Data   models.Exec `json:"data"`
	}{
		Status: "success",
		Count:  1,
		Data:   exec,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}

func AddManyExecHandler(w http.ResponseWriter, r *http.Request) {
	var newExec []models.Exec
	var rawExec []map[string]json.RawMessage
	err := json.NewDecoder(r.Body).Decode(&rawExec)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if len(rawExec) == 0 {
		http.Error(w, "At least one exec is required", http.StatusBadRequest)
		return
	}

	val := reflect.TypeOf(models.Exec{})
	allowedFields := make(map[string]struct{})
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		key := strings.Split(field.Tag.Get("json"), ",")[0]
		allowedFields[key] = struct{}{}
	}

	for _, execs := range rawExec {
		for key := range execs {
			if _, ok := allowedFields[key]; !ok {
				http.Error(w, fmt.Sprintf("Invalid field: %s", key), http.StatusBadRequest)
				return
			}
		}
	}
	data, err := json.Marshal(rawExec)
	if err != nil {
		http.Error(w, "Invalid Request body", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(data, &newExec); err != nil {
		http.Error(w, "Invalid Request body", http.StatusBadRequest)
		return
	}

	for _, exec := range newExec {
		val := reflect.ValueOf(exec)
		for i := 0; i < val.NumField(); i++ {
			if val.Field(i).Kind() == reflect.String && val.Field(i).String() == "" {
				http.Error(w, "All fields are required", http.StatusBadRequest)
				return
			}
		}
	}

	addedExecs, err := sqlconnect.AddExecsDBHandler(newExec)
	if err != nil {
		writeExecError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string        `json:"status"`
		Count  int           `json:"count"`
		Data   []models.Exec `json:"data"`
	}{
		Status: "success",
		Count:  len(addedExecs),
		Data:   addedExecs,
	}
	json.NewEncoder(w).Encode(response)

}

func UpdateOneExecPatchHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid exec ID", http.StatusBadRequest)
		return
	}
	var update map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	existingExecs, err := sqlconnect.UpdateOneExecPatchDBHandler(id, update)
	if err != nil {
		if strings.Contains(err.Error(), "invalid value") || strings.Contains(err.Error(), "unknown exec field") || strings.Contains(err.Error(), "editable columns") || strings.Contains(err.Error(), "password must be changed") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeExecError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(existingExecs)

}

func UpdateExecsPatchHandler(w http.ResponseWriter, r *http.Request) {
	var updates []map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid Request payload", http.StatusBadRequest)
		return
	}
	if len(updates) == 0 {
		http.Error(w, "At least one exec update is required", http.StatusBadRequest)
		return
	}

	updatesDB, err := sqlconnect.UpdateExecsPatchDBHandler(updates)
	if err != nil {
		if strings.Contains(err.Error(), "id") || strings.Contains(err.Error(), "field") || strings.Contains(err.Error(), "value") || strings.Contains(err.Error(), "editable columns") || strings.Contains(err.Error(), "password must be changed") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeExecError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Execs updated successfully",
		"data":    updatesDB,
	})

}

func DeleteExecsHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid exec ID", http.StatusBadRequest)
		return
	}
	err = sqlconnect.DeleteOneExecsDBHandler(id)
	if err != nil {
		writeExecError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := struct {
		Status string `json:"status"`
		ID     int    `json:"id"`
	}{
		Status: "Exec successfully deleted",
		ID:     id,
	}

	json.NewEncoder(w).Encode(response)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	loginHandler(w, r, sqlconnect.GetExecByUsername, auth.DefaultSessionStore, auth.DefaultLoginLimiter)
}

type execByUsernameLookup func(string) (models.Exec, error)

func loginHandler(w http.ResponseWriter, r *http.Request, lookup execByUsernameLookup, sessions *auth.SessionStore, limiter *auth.LoginLimiter) {
	var req models.Login
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		http.Error(w, "Request body must contain one JSON object", http.StatusBadRequest)
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}
	limiterKeys := []string{"username:" + req.Username, "ip:" + requestIP(r)}
	for _, key := range limiterKeys {
		if !limiter.Allow(key) {
			http.Error(w, "Too many login attempts", http.StatusTooManyRequests)
			return
		}
	}
	user, err := lookup(req.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			recordLoginFailure(limiter, limiterKeys)
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Unable to process login", http.StatusInternalServerError)
		return
	}

	valid, err := auth.VerifyPassword(req.Password, user.Password)
	if err != nil {
		recordLoginFailure(limiter, limiterKeys)
		http.Error(w, "Unable to process login", http.StatusInternalServerError)
		return
	}
	if !valid || user.Inactive {
		recordLoginFailure(limiter, limiterKeys)
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}
	for _, key := range limiterKeys {
		limiter.Reset(key)
	}

	token, expiresAt, err := sessions.Create(user.ID)
	if err != nil {
		http.Error(w, "Unable to create session", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   24 * 60 * 60,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   user,
	}); err != nil {
		return
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	logoutHandler(w, r, auth.DefaultSessionStore)
}

func logoutHandler(w http.ResponseWriter, r *http.Request, sessions *auth.SessionStore) {
	cookie, err := r.Cookie(auth.SessionCookieName)
	if err == nil {
		sessions.Delete(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

func requestIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func recordLoginFailure(limiter *auth.LoginLimiter, keys []string) {
	for _, key := range keys {
		limiter.RecordFailure(key)
	}
}

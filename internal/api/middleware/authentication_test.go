package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Pratiksable/student-management/internal/auth"
)

func TestRequireAuthentication(t *testing.T) {
	store := auth.NewSessionStore(time.Hour)
	token, _, err := store.Create(23)
	if err != nil {
		t.Fatal(err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		execID, ok := AuthenticatedExecID(r)
		if !ok || execID != 23 {
			t.Fatalf("unexpected authenticated exec: id=%d ok=%v", execID, ok)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	handler := RequireAuthentication(store)(next)

	request := httptest.NewRequest(http.MethodGet, "/execs", nil)
	request.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusNoContent)
	}
}

func TestRequireAuthenticationRejectsMissingOrInvalidSession(t *testing.T) {
	store := auth.NewSessionStore(time.Hour)
	handler := RequireAuthentication(store)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("protected handler was called")
	}))

	for _, token := range []string{"", "invalid"} {
		request := httptest.NewRequest(http.MethodGet, "/execs", nil)
		if token != "" {
			request.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("token %q got status %d", token, response.Code)
		}
	}
}

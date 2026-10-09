package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Pratiksable/student-management/internal/auth"
	"github.com/Pratiksable/student-management/internal/models"
)

func TestLoginHandler(t *testing.T) {
	passwordHash, err := auth.HashPassword("ChangeMe123!")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		body       string
		lookup     execByUsernameLookup
		statusCode int
		contains   string
	}{
		{
			name: "valid credentials",
			body: `{"username":"pratik","password":"ChangeMe123!"}`,
			lookup: func(username string) (models.Exec, error) {
				return models.Exec{ID: 1, Username: username, Password: passwordHash, Role: "admin"}, nil
			},
			statusCode: http.StatusOK,
			contains:   `"status":"success"`,
		},
		{
			name: "wrong password",
			body: `{"username":"pratik","password":"wrong"}`,
			lookup: func(username string) (models.Exec, error) {
				return models.Exec{Username: username, Password: passwordHash}, nil
			},
			statusCode: http.StatusUnauthorized,
			contains:   "Invalid username or password",
		},
		{
			name: "unknown username",
			body: `{"username":"missing","password":"anything"}`,
			lookup: func(string) (models.Exec, error) {
				return models.Exec{}, sql.ErrNoRows
			},
			statusCode: http.StatusUnauthorized,
			contains:   "Invalid username or password",
		},
		{
			name: "inactive account",
			body: `{"username":"pratik","password":"ChangeMe123!"}`,
			lookup: func(username string) (models.Exec, error) {
				return models.Exec{Username: username, Password: passwordHash, Inactive: true}, nil
			},
			statusCode: http.StatusUnauthorized,
			contains:   "Invalid username or password",
		},
		{
			name:       "unknown JSON field",
			body:       `{"username":"pratik","password":"x","role":"admin"}`,
			lookup:     unexpectedLoginLookup(t),
			statusCode: http.StatusBadRequest,
			contains:   "Invalid Request Body",
		},
		{
			name:       "trailing JSON value",
			body:       `{"username":"pratik","password":"x"} {}`,
			lookup:     unexpectedLoginLookup(t),
			statusCode: http.StatusBadRequest,
			contains:   "one JSON object",
		},
		{
			name: "database failure",
			body: `{"username":"pratik","password":"x"}`,
			lookup: func(string) (models.Exec, error) {
				return models.Exec{}, errors.New("database unavailable")
			},
			statusCode: http.StatusInternalServerError,
			contains:   "Unable to process login",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/execs/login", strings.NewReader(tc.body))
			response := httptest.NewRecorder()
			sessions := auth.NewSessionStore(time.Hour)
			limiter := auth.NewLoginLimiter(5, time.Minute)

			loginHandler(response, request, tc.lookup, sessions, limiter)

			if response.Code != tc.statusCode {
				t.Fatalf("got status %d, want %d; body=%q", response.Code, tc.statusCode, response.Body.String())
			}
			if !strings.Contains(response.Body.String(), tc.contains) {
				t.Fatalf("response %q does not contain %q", response.Body.String(), tc.contains)
			}
			if strings.Contains(response.Body.String(), passwordHash) {
				t.Fatal("response exposed the password hash")
			}
			if tc.statusCode == http.StatusOK {
				cookies := response.Result().Cookies()
				if len(cookies) != 1 || cookies[0].Name != auth.SessionCookieName || !cookies[0].HttpOnly || !cookies[0].Secure {
					t.Fatalf("login did not return a secure session cookie: %#v", cookies)
				}
				if _, err := sessions.Get(cookies[0].Value); err != nil {
					t.Fatalf("cookie does not reference a valid server session: %v", err)
				}
			}
		})
	}
}

func unexpectedLoginLookup(t *testing.T) execByUsernameLookup {
	t.Helper()
	return func(string) (models.Exec, error) {
		t.Fatal("lookup should not be called")
		return models.Exec{}, nil
	}
}

func TestLoginRateLimit(t *testing.T) {
	passwordHash, err := auth.HashPassword("ChangeMe123!")
	if err != nil {
		t.Fatal(err)
	}
	lookup := func(username string) (models.Exec, error) {
		return models.Exec{ID: 1, Username: username, Password: passwordHash}, nil
	}
	sessions := auth.NewSessionStore(time.Hour)
	limiter := auth.NewLoginLimiter(2, time.Minute)

	for attempt := 1; attempt <= 3; attempt++ {
		request := httptest.NewRequest(http.MethodPost, "/execs/login", strings.NewReader(`{"username":"pratik","password":"wrong"}`))
		request.RemoteAddr = "192.0.2.1:1234"
		response := httptest.NewRecorder()
		loginHandler(response, request, lookup, sessions, limiter)

		want := http.StatusUnauthorized
		if attempt == 3 {
			want = http.StatusTooManyRequests
		}
		if response.Code != want {
			t.Fatalf("attempt %d got status %d, want %d", attempt, response.Code, want)
		}
	}
}

func TestLogoutDeletesSessionAndCookie(t *testing.T) {
	store := auth.NewSessionStore(time.Hour)
	token, _, err := store.Create(5)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/execs/logout", nil)
	request.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: token})
	response := httptest.NewRecorder()

	logoutHandler(response, request, store)

	if response.Code != http.StatusNoContent {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusNoContent)
	}
	if _, err := store.Get(token); err == nil {
		t.Fatal("logout did not invalidate the server session")
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != auth.SessionCookieName || cookies[0].MaxAge >= 0 {
		t.Fatalf("logout did not clear the session cookie: %#v", cookies)
	}
}

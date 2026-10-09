package middleware

import (
	"context"
	"net/http"

	"github.com/Pratiksable/student-management/internal/auth"
)

type authenticatedExecIDKey struct{}

func RequireAuthentication(store *auth.SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(auth.SessionCookieName)
			if err != nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}
			session, err := store.Get(cookie.Value)
			if err != nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), authenticatedExecIDKey{}, session.ExecID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AuthenticatedExecID(r *http.Request) (int, bool) {
	execID, ok := r.Context().Value(authenticatedExecIDKey{}).(int)
	return execID, ok
}

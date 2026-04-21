package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/todooo/todo-api/internal/auth"
	"github.com/todooo/todo-api/internal/httpx"
)

type ctxKey string

const userIDKey ctxKey = "userID"

func UserID(r *http.Request) (string, bool) {
	v, ok := r.Context().Value(userIDKey).(string)
	return v, ok && v != ""
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	allowAll := len(allowedOrigins) == 0
	set := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		set[strings.TrimSpace(o)] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				_, ok := set[origin]
				if allowAll || ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, "+auth.CSRFHeader)
					w.Header().Set("Access-Control-Max-Age", "600")
				}
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v %s %s", rec, r.Method, r.URL.Path)
				httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func Auth(mgr *auth.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(auth.SessionCookie)
			if err != nil || cookie.Value == "" {
				httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			claims, err := mgr.ParseToken(cookie.Value)
			if err != nil {
				httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			ctx := WithUserID(r.Context(), claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
			cookie, err := r.Cookie(auth.CSRFCookie)
			header := r.Header.Get(auth.CSRFHeader)
			if err != nil || cookie.Value == "" || header == "" || cookie.Value != header {
				httpx.WriteError(w, http.StatusForbidden, "invalid csrf token")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

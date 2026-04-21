package router

import (
	"net/http"
	"strings"

	"github.com/todooo/todo-api/internal/auth"
	"github.com/todooo/todo-api/internal/handlers"
	"github.com/todooo/todo-api/internal/httpx"
	"github.com/todooo/todo-api/internal/middleware"
)

type Deps struct {
	AuthMgr     *auth.Manager
	AuthH       *handlers.AuthHandler
	CategoriesH *handlers.CategoryHandler
	TasksH      *handlers.TaskHandler
}

func New(deps Deps, allowedOrigins []string) http.Handler {
	mux := http.NewServeMux()

	// Health (unauthenticated, mounted both at /health and /api/health)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/health", healthHandler)

	// Public auth endpoints
	mux.Handle("/api/auth/register", methodOnly(http.MethodPost, http.HandlerFunc(deps.AuthH.Register)))
	mux.Handle("/api/auth/login", methodOnly(http.MethodPost, http.HandlerFunc(deps.AuthH.Login)))

	// Authenticated endpoints — wrap with Auth + CSRF middleware
	authMW := middleware.Auth(deps.AuthMgr)

	mux.Handle("/api/auth/logout", authMW(middleware.CSRF(methodOnly(http.MethodPost, http.HandlerFunc(deps.AuthH.Logout)))))
	mux.Handle("/api/auth/me", authMW(methodOnly(http.MethodGet, http.HandlerFunc(deps.AuthH.Me))))

	mux.Handle("/api/categories", authMW(middleware.CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			deps.CategoriesH.List(w, r)
		case http.MethodPost:
			deps.CategoriesH.Create(w, r)
		default:
			httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	}))))

	mux.Handle("/api/categories/", authMW(middleware.CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/categories/")
		if id == "" || strings.Contains(id, "/") {
			httpx.WriteError(w, http.StatusNotFound, "not found")
			return
		}
		if r.Method != http.MethodDelete {
			httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		deps.CategoriesH.Delete(w, r, id)
	}))))

	mux.Handle("/api/tasks", authMW(middleware.CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			deps.TasksH.List(w, r)
		case http.MethodPost:
			deps.TasksH.Create(w, r)
		default:
			httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	}))))

	mux.Handle("/api/tasks/", authMW(middleware.CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
		if id == "" || strings.Contains(id, "/") {
			httpx.WriteError(w, http.StatusNotFound, "not found")
			return
		}
		switch r.Method {
		case http.MethodGet:
			deps.TasksH.Get(w, r, id)
		case http.MethodPut:
			deps.TasksH.Update(w, r, id)
		case http.MethodDelete:
			deps.TasksH.Delete(w, r, id)
		default:
			httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	}))))

	// Compose middleware: Recover -> CORS -> mux
	var h http.Handler = mux
	h = middleware.CORS(allowedOrigins)(h)
	h = middleware.Recover(h)
	return h
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func methodOnly(method string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.ServeHTTP(w, r)
	})
}

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/todooo/todo-api/internal/auth"
	"github.com/todooo/todo-api/internal/handlers"
	"github.com/todooo/todo-api/internal/router"
	"github.com/todooo/todo-api/internal/store"
)

func main() {
	port := getEnv("PORT", "9090")
	jwtSecret := getEnv("JWT_SECRET", "dev-insecure-secret-change-me-please-0123456789abcdef")
	cookieSecure := strings.EqualFold(getEnv("COOKIE_SECURE", "false"), "true")
	cookieDomain := getEnv("COOKIE_DOMAIN", "")

	originsEnv := getEnv("ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:8080,http://localhost:3000")
	var allowedOrigins []string
	for _, o := range strings.Split(originsEnv, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			allowedOrigins = append(allowedOrigins, o)
		}
	}

	st := store.NewMemory()
	mgr := auth.NewManager([]byte(jwtSecret))

	authH := &handlers.AuthHandler{
		Store:        st,
		Auth:         mgr,
		CookieSecure: cookieSecure,
		CookieDomain: cookieDomain,
	}
	catH := &handlers.CategoryHandler{Store: st}
	taskH := &handlers.TaskHandler{Store: st}

	handler := router.New(router.Deps{
		AuthMgr:     mgr,
		AuthH:       authH,
		CategoriesH: catH,
		TasksH:      taskH,
	}, allowedOrigins)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       20 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("todo-api listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

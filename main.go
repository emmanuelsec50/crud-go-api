package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"chi_practise/internal/user"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var repo = user.NewRepository()
var service = user.NewService(repo)
var handler = user.NewHandler(service)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) writeHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

var secretKey = "secret-key-123"

func requireAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey != secretKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		path := r.URL.Path
		method := r.Method
		fmt.Printf("[REQUEST] %s %s", method, path)
		start := time.Now()
		next.ServeHTTP(rw, r)
		elasped := time.Since(start)
		newPath := r.URL.Path
		newMethod := r.Method
		status := rw.statusCode
		fmt.Printf("[RESPONSE] %s %s %d %s", newMethod, newPath, status, elasped)
	})
}

func main() {

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(requestLogger)
	r.Get("/", helloWorld)
	r.Group(func(r chi.Router) {
		r.Use(requireAPI)
		r.Use(middleware.Timeout(5 * time.Second))
		// r.Route("/users", userRoutes)
		r.Route("/users", func(r chi.Router) {
			userRoutes(r, handler)
		})

	})

	fmt.Println("Server running in port 3000")
	err := http.ListenAndServe(":3000", r)
	if err != nil {
		fmt.Println("Server crashed")
	}
}

func userRoutes(r chi.Router, h *user.UserHandler) {
	r.Get("/", h.AllUsers)
	r.Get("/search", h.GetUserByName)
	r.Get("/{id}", h.GetUserByID)
	r.With(adminOnly).Delete("/{id}", h.DeleteUserByID)
	r.Post("/", h.CreateUser)

}

type ctxKey string

const userRole ctxKey = "Role"

type Authuser struct {
	ID   int
	Role string
}

func adminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := r.Header.Get("X-Role")
		if role != "admin" {
			http.Error(w, "Unauthorized access", http.StatusForbidden)
			return
		}
		obj := Authuser{
			ID:   1,
			Role: role,
		}
		ctx := context.WithValue(r.Context(), userRole, obj)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world"))
}

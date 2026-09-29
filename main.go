package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type User struct {
	Name string `json:"name"`
	Age  string `json:"age"`
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) writeHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

var (
	userStore = map[int]User{
		1: {Name: "Emmanuel", Age: "20"},
	}
	nextID = 2
)
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
		fmt.Printf("[RESPONSE] %s %s %d", newMethod, newPath, status)
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
		r.Route("/users", userRoutes)

	})

	fmt.Println("Server running in port 3000")
	err := http.ListenAndServe(":3000", r)
	if err != nil {
		fmt.Println("Server crashed")
	}
}

func userRoutes(r chi.Router) {
	r.Get("/", allUsers)
	r.Get("/search", getUserByName)
	r.Get("/{id}", getUserByID)
	r.With(adminOnly).Delete("/{id}", deleteUserByID)
	r.Post("/", createUser)

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

func createUser(w http.ResponseWriter, r *http.Request) {
	ch := make(chan bool)
	go func() {
		select {
		case <-time.After(2 * time.Second):
			ch <- true
		case <-r.Context().Done():
			return

		}
	}()
	select {
	case <-ch:
		fmt.Println("Finished processing")
	case <-r.Context().Done():
		fmt.Println("Request timed out")
		return
	}
	id := middleware.GetReqID(r.Context())
	fmt.Printf("Request ID is: %s\n", id)
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Error in decoding", http.StatusBadRequest)
		return
	}
	userStore[nextID] = user
	nextID++
	w.WriteHeader(http.StatusCreated)

}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world"))
}

func getUserByName(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	for index, item := range userStore {
		if name == item.Name {
			var body map[int]User
			body = map[int]User{index: {item.Name, item.Age}}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(body)
			return
		}

	}
	http.Error(w, "Invalid username", http.StatusNotFound)

}

func allUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userStore)
}
func getUserByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	userID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	obj, exists := userStore[userID]
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	body := User{
		Name: obj.Name,
		Age:  obj.Age,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(body)
}

func deleteUserByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, ok := r.Context().Value(userRole).(Authuser)
	if !ok {
		fmt.Println("Role not found")
		return
	}
	fmt.Printf("Authenticated user: %d\n", role.ID)
	fmt.Printf("Role: %s\n", role.Role)

	userID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	_, exists := userStore[userID]
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	delete(userStore, userID)

	w.WriteHeader(http.StatusNoContent)
}

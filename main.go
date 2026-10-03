package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"chi_practise/internal/user"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// type User struct {
// 	Name string `json:"name"`
// 	Age  string `json:"age"`
// }

type userService struct {
	userStore UserRepository
	nextID    int
}

type userRepository struct {
	mapStore map[int]user.User
}

type backupRepository struct {
	user map[int]user.User
}
type userHandler struct {
	service *userService
}
type UserRepository interface {
	Create(user user.User, id int)
	Get(id int) (user.User, bool)
	GetUserByName(name string) map[int]user.User
	DeleteUserByID(id int) bool
	GetAllUsers() []map[int]user.User
}

func (r *backupRepository) Create(user user.User, id int) {
	r.user[id] = user
}
func (r *backupRepository) Get(id int) (user.User, bool) {
	user, exists := r.user[id]
	return user, exists
}
func (r *backupRepository) GetUserByName(name string) map[int]user.User {
	for index, item := range r.user {
		if name == item.Name {
			body := map[int]user.User{index: {Name: item.Name, Age: item.Age}}
			return body
		}
	}
	return map[int]user.User{}
}
func (r *backupRepository) DeleteUserByID(id int) bool {
	_, exists := r.Get(id)
	if exists {
		delete(r.user, id)
	}
	return exists

}
func (r *backupRepository) GetAllUsers() []map[int]user.User {
	var allUsers []map[int]user.User
	for i, item := range r.user {
		allUsers = append(allUsers, map[int]user.User{i: {Name: item.Name, Age: item.Age}})
	}

	return allUsers
}

func (r *userRepository) Create(user user.User, id int) {
	r.mapStore[id] = user
}

func (r *userRepository) Get(id int) (user.User, bool) {
	user, exists := r.mapStore[id]
	return user, exists
}
func (r *userRepository) GetUserByName(name string) map[int]user.User {
	for index, item := range r.mapStore {
		if name == item.Name {
			body := map[int]user.User{index: {Name: item.Name, Age: item.Age}}
			return body
		}
	}
	return map[int]user.User{}
}

func (r *userRepository) DeleteUserByID(id int) bool {
	_, exists := r.Get(id)
	if exists {
		delete(r.mapStore, id)
	}
	return exists

}

func (r *userRepository) GetAllUsers() []map[int]user.User {
	var allUsers []map[int]user.User
	for i, item := range r.mapStore {
		allUsers = append(allUsers, map[int]user.User{i: {Name: item.Name, Age: item.Age}})
	}

	return allUsers
}

func (u *userService) Create(user user.User) int {
	id := u.nextID
	u.userStore.Create(user, id)
	u.nextID++
	return id
}

func (u *userService) Get(id int) (user.User, bool) {
	obj, exists := u.userStore.Get(id)
	if !exists {
		return user.User{}, exists // line 24
	}
	return obj, exists
}
func (u *userService) getUserByName(name string) map[int]user.User {
	user := u.userStore.GetUserByName(name)
	return user
}

func (u *userService) deleteUserByID(id int) bool {
	exists := u.userStore.DeleteUserByID(id)
	return exists
}

func (u *userService) getAllUsers() []map[int]user.User {
	return u.userStore.GetAllUsers()
}

func NewService(repo UserRepository) *userService {
	return &userService{
		userStore: repo,
		nextID:    1,
	}
}
func NewHandler(service *userService) *userHandler {
	return &userHandler{
		service: service,
	}
}

var repo = &userRepository{
	mapStore: make(map[int]user.User),
}
var service = NewService(repo)
var handler = NewHandler(service)

// var store = userService{
// 	userStore: &userRepository{map[int]User{1: {Name: "Emmanuel", Age: "65"}}},
// 	nextID:    2,
// }
// var backupStore = userService{
// 	userStore: &backupRepository{make(map[int]User)},
// 	nextID:    0,
// }

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) writeHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// var (
//
//	userStore = map[int]User{
//		1: {Name: "Emmanuel", Age: "20"},
//	}
//	nextID = 2
//
// )
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

func userRoutes(r chi.Router, h *userHandler) {
	r.Get("/", h.allUsers)
	r.Get("/search", h.getUserByName)
	r.Get("/{id}", h.getUserByID)
	r.With(adminOnly).Delete("/{id}", h.deleteUserByID)
	r.Post("/", h.createUser)

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

func (h *userHandler) createUser(w http.ResponseWriter, r *http.Request) {
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
	var obj user.User
	err := json.NewDecoder(r.Body).Decode(&obj)
	if err != nil {
		http.Error(w, "Error in decoding", http.StatusBadRequest)
		return
	}

	newID := h.service.Create(obj)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[int]user.User{newID: {Name: obj.Name, Age: obj.Age}}) // line 310

}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world"))
}

func (h *userHandler) getUserByName(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	body := h.service.getUserByName(name)
	if body != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(body)
		return
	}
	http.Error(w, "Invalid username", http.StatusNotFound)

}

func (h *userHandler) getUserByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	userID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	obj, exists := h.service.Get(userID)
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	body := user.User{
		Name: obj.Name,
		Age:  obj.Age,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(body)
}

func (h *userHandler) allUsers(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.service.getAllUsers())
}

func (h *userHandler) deleteUserByID(w http.ResponseWriter, r *http.Request) {
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

	exists := h.service.deleteUserByID(userID)
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

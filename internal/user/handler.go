package user

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Authuser struct {
	ID   int
	Role string
}
type UserHandler struct {
	service *UserService
}

func NewHandler(service *UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

type ctxKey string

const userRole ctxKey = "Role"

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
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
	var obj User
	err := json.NewDecoder(r.Body).Decode(&obj)
	if err != nil {
		http.Error(w, "Error in decoding", http.StatusBadRequest)
		return
	}

	newID := h.service.Create(obj)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[int]User{newID: {Name: obj.Name, Age: obj.Age}}) // line 310

}

func (h *UserHandler) GetUserByName(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	body := h.service.GetUserByName(name)
	if body != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(body)
		return
	}
	http.Error(w, "Invalid username", http.StatusNotFound)

}
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	userID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	obj, err := h.service.Get(userID)
	if err != nil {
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
func (h *UserHandler) AllUsers(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.service.GetAllUsers())
}
func (h *UserHandler) DeleteUserByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, ok := r.Context().Value(userRole).(Authuser)
	if !ok {
		return
	}

	fmt.Printf("Authenticated user: %d\n", role.ID)
	fmt.Printf("Role: %s\n", role.Role)

	userID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	exists := h.service.DeleteUserByID(userID)
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

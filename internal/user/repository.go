package user

type userRepository struct {
	mapStore map[int]User
}
type backupRepository struct {
	user map[int]User
}
type UserRepository interface {
	Create(user User, id int)
	Get(id int) (User, bool)
	GetUserByName(name string) map[int]User
	DeleteUserByID(id int) bool
	GetAllUsers() []map[int]User
}

func (r *backupRepository) Create(user User, id int) {
	r.user[id] = user
}
func (r *backupRepository) Get(id int) (User, bool) {
	user, exists := r.user[id]
	return user, exists
}
func (r *backupRepository) GetUserByName(name string) map[int]User {
	for index, item := range r.user {
		if name == item.Name {
			body := map[int]User{index: {Name: item.Name, Age: item.Age}}
			return body
		}
	}
	return map[int]User{}
}
func (r *backupRepository) DeleteUserByID(id int) bool {
	_, exists := r.Get(id)
	if exists {
		delete(r.user, id)
	}
	return exists

}
func (r *backupRepository) GetAllUsers() []map[int]User {
	var allUsers []map[int]User
	for i, item := range r.user {
		allUsers = append(allUsers, map[int]User{i: {Name: item.Name, Age: item.Age}})
	}

	return allUsers
}

func (r *userRepository) Create(user User, id int) {
	r.mapStore[id] = user
}

func (r *userRepository) Get(id int) (User, bool) {
	user, exists := r.mapStore[id]
	return user, exists
}
func (r *userRepository) GetUserByName(name string) map[int]User {
	for index, item := range r.mapStore {
		if name == item.Name {
			body := map[int]User{index: {Name: item.Name, Age: item.Age}}
			return body
		}
	}
	return map[int]User{}
}

func (r *userRepository) DeleteUserByID(id int) bool {
	_, exists := r.Get(id)
	if exists {
		delete(r.mapStore, id)
	}
	return exists

}

func (r *userRepository) GetAllUsers() []map[int]User {
	var allUsers []map[int]User
	for i, item := range r.mapStore {
		allUsers = append(allUsers, map[int]User{i: {Name: item.Name, Age: item.Age}})
	}

	return allUsers
}
func NewRepository() UserRepository {
	return &userRepository{
		mapStore: make(map[int]User),
	}
}

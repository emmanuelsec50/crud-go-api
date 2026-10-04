package user

type UserService struct {
	repo   UserRepository
	nextID int
}

func (u *UserService) Create(user User) int {
	id := u.nextID
	u.repo.Create(user, id)
	u.nextID++
	return id
}

func (u *UserService) Get(id int) (User, bool) {
	obj, exists := u.repo.Get(id)
	if !exists {
		return User{}, exists // line 24
	}
	return obj, exists
}
func (u *UserService) GetUserByName(name string) map[int]User {
	user := u.repo.GetUserByName(name)
	return user
}

func (u *UserService) DeleteUserByID(id int) bool {
	exists := u.repo.DeleteUserByID(id)
	return exists
}

func (u *UserService) GetAllUsers() []map[int]User {
	return u.repo.GetAllUsers()
}

func NewService(repo UserRepository) *UserService {
	return &UserService{
		repo:   repo,
		nextID: 1,
	}
}

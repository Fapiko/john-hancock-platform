package users

type User struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
	Password  string
}

func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
	}
}

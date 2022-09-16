package daos

import "github.com/fapiko/john-hancock-platform/app/contracts"

func NewUserFromProps(props map[string]interface{}) *User {
	return &User{
		ID:        props["uuid"].(string),
		FirstName: props["firstName"].(string),
		LastName:  props["lastName"].(string),
		Email:     props["email"].(string),
		Password:  props["password"].(string),
	}
}

type User struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
	Password  string
}

func (u *User) ToResponse() *contracts.UserResponse {
	return &contracts.UserResponse{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
	}
}

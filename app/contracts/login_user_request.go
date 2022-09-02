package contracts

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

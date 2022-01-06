package users

type LoginUserResponse struct {
	Session *Session      `json:"session"`
	User    *UserResponse `json:"user"`
}

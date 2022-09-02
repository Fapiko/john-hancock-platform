package contracts

type LoginUserResponse struct {
	Session *SessionResponse `json:"session"`
	User    *UserResponse    `json:"user"`
}

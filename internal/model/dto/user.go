package dto

type RegisterUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

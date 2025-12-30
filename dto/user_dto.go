package dto

type UserResponse struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Address    string `json:"address"`
	IsVerified bool   `json:"is_verified"`
	IsAdmin    bool   `json:"is_admin"`
}

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Address  string `json:"address"`
	Password string `json:"password" binding:"required,min=6"`
}

type UpdateUserRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email" binding:"omitempty,email"`
	Address string `json:"address"`
}

type UserFilterRequest struct {
	PaginationRequest
	Email    string `form:"email"`
	IsAdmin  *bool  `form:"is_admin"`
	Name     string `form:"name"`
}

package dto

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=1"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required,min=1"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type UpdateUserRequest struct {
	Name  *string `json:"name" binding:"omitempty,min=1"`
	Email *string `json:"email" binding:"omitempty,email"`
}

type LotterySearchRequest struct {
	Pattern string `form:"pattern" binding:"required,len=6"`
	Limit   int    `form:"limit" binding:"omitempty,min=1,max=100"`
}

package domain

type RegisterRequest struct {
	UserName string `json:"username"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type CreateUserResponse struct {
	UserId int64 `json:"user_id"`
}

type RegisterResponse struct {
	Status string             `json:"status" example:"success"`
	Data   CreateUserResponse `json:"data"`
}

type LoginRequest struct {
	Password string `json:"password" binding:"required,min=8"`
	Email    string `json:"email" binding:"required,email"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
type LoginResponse struct {
	Status string        `json:"status" example:"success"`
	Data   TokenResponse `json:"data"`
}

type TokensRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokensResponse struct {
	Status string        `json:"status" example:"success"`
	Data   TokenResponse `json:"data"`
}

// --- PROFILE DTO ---
type ProfileResponse struct {
	UserID   int64  `json:"user_id"`
	UserName string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// --- ADMIN DTO ---
type AssignRoleReq struct {
	TargetUserID int64  `json:"target_user_id" binding:"required"`
	NewRole      string `json:"new_role" binding:"required"`
}

type PendingUser struct {
	UserID    int64  `json:"user_id"`
	UserName  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type ListPendingUsersResponse struct {
	Users      []PendingUser `json:"users"`
	TotalCount int32         `json:"total_count"`
}
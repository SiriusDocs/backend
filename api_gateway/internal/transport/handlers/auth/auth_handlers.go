package auth

import (
	"context"

	"github.com/SiriusDocs/backend/api_gateway/internal/domain"
	"github.com/SiriusDocs/backend/api_gateway/internal/lib/response"
	"github.com/SiriusDocs/protos/gen/go/auth"
	"github.com/gin-gonic/gin"
)

// @Summary      User registration
// @Description  Creates a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body domain.RegisterRequest true "user data"
// @Success      200  {object} domain.RegisterResponse "user ID"
// @Failure      400  {object}  response.ErrorResponseMes "validation error"
// @Failure      500  {object}  response.ErrorResponseMes "Internal server error"
// @Router       /auth/sign-up [post]
func (h *Handler) signUp(c *gin.Context) {
	var input domain.RegisterRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.client.Timeout)
	defer cancel()

	resp, err := h.service.Register(ctx, &auth.RegisterRequest{
		Email:    input.Email,
		Password: input.Password,
		Username: input.UserName,
	})
	
	if err != nil {
		response.ParseGRPCError(c, h.log, err, "register user")
		return
	}
	response.Success(c, gin.H{"user_id": resp.UserId})
}

// @Summary      User authentication
// @Description  Authenticates a user and returns access and refresh tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body domain.LoginRequest true "User credentials"
// @Success      200  {object}  domain.LoginResponse "Successful login with tokens"
// @Failure      400  {object}  response.ErrorResponseMes "Validation error"
// @Failure      401  {object}  response.ErrorResponseMes "Incorrect password or email"
// @Failure      500  {object}  response.ErrorResponseMes "Internal server error"
// @Router       /auth/sign-in [post]
func (h *Handler) signIn(c *gin.Context) {
	var input domain.LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.client.Timeout)
	defer cancel()

	resp, err := h.service.Login(ctx, &auth.LoginRequest{
		Password: input.Password,
		Email: input.Email,
	})
	
	if err != nil {
		response.ParseGRPCError(c, h.log, err, "login user")
		return
	}
	response.Success(c, gin.H{"access_token": resp.AccessToken, "refresh_token": resp.RefreshToken})
}

package auth

import (
	"context"
	"net/http"
	"strconv"

	"github.com/SiriusDocs/backend/api_gateway/internal/domain"
	"github.com/SiriusDocs/backend/api_gateway/internal/lib/response"
	"github.com/SiriusDocs/backend/api_gateway/internal/transport/middleware"
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
		Email:    input.Email,
	})

	if err != nil {
		response.ParseGRPCError(c, h.log, err, "login user")
		return
	}
	response.Success(c, gin.H{"access_token": resp.AccessToken, "refresh_token": resp.RefreshToken})
}

// @Summary      Request new tokens
// @Description  After the access token expires (15 minutes), you need to send a request to update the tokens. The request body must have a refresh token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body domain.TokensRequest true "refresh token"
// @Success      200  {object}  domain.TokensResponse "Successful login with tokens"
// @Failure      400  {object}  response.ErrorResponseMes "Validation error"
// @Failure      401  {object}  response.ErrorResponseMes "Incorrect password or email"
// @Failure      500  {object}  response.ErrorResponseMes "Internal server error"
// @Router       /auth/refresh [post]
func (h *Handler) refreshToken(c *gin.Context) {
	var input domain.TokensRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.client.Timeout)
	defer cancel()

	resp, err := h.service.GetNewTokens(ctx, &auth.TokensRequest{
		RefreshToken: input.RefreshToken,
	})

	if err != nil {
		response.ParseGRPCError(c, h.log, err, "get new tokens")
		return
	}
	response.Success(c, gin.H{"access_token": resp.AccessToken, "refresh_token": resp.RefreshToken})
}

// @Summary      Get user profile
// @Description  Returns user profile data for the current authenticated user
// @Security ApiKeyAuth
// @Tags         profile
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  domain.ProfileResponse
// @Router       /profile/me [get]
func (h *Handler) getProfile(c *gin.Context) {
	// Берем ID из токена, а не из URL! Защита от просмотра чужих профилей.
	userID, err := middleware.GetUserId(c)
	if err != nil {
		response.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.client.Timeout)
	defer cancel()

	resp, err := h.service.GetProfile(ctx, &auth.GetProfileRequest{
		UserId: userID,
	})

	if err != nil {
		response.ParseGRPCError(c, h.log, err, "get user profile")
		return
	}

	response.Success(c, domain.ProfileResponse{
		UserID:   resp.UserId,
		UserName: resp.Username,
		Email:    resp.Email,
		Role:     resp.Role,
	})
}

// @Summary      List pending users
// @Description  Admin only. Returns a paginated list of users waiting for a role assignment
// @Security ApiKeyAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        limit query int false "Limit" default(10)
// @Param        offset query int false "Offset" default(0)
// @Success      200  {object}  domain.ListPendingUsersResponse
// @Failure      400  {object}  response.ErrorResponseMes
// @Failure      403  {object}  response.ErrorResponseMes "Access denied"
// @Failure      500  {object}  response.ErrorResponseMes
// @Router       /admin/pending-users [get]
func (h *Handler) listPendingUsers(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "invalid limit parameter")
		return
	}

	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "invalid offset parameter")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.client.Timeout)
	defer cancel()

	resp, err := h.service.ListPendingUsers(ctx, &auth.ListPendingUsersRequest{
		Limit:  int32(limit),
		Offset: int32(offset),
	})

	if err != nil {
		response.ParseGRPCError(c, h.log, err, "list pending users")
		return
	}

	// Мапим proto массив в доменный массив
	users := make([]domain.PendingUser, 0, len(resp.Users))
	for _, u := range resp.Users {
		users = append(users, domain.PendingUser{
			UserID:    u.UserId,
			UserName:  u.Username,
			Email:     u.Email,
			CreatedAt: u.CreatedAt,
		})
	}

	response.Success(c, domain.ListPendingUsersResponse{
		Users:      users,
		TotalCount: resp.TotalCount,
	})
}

// @Summary      Assign role to user
// @Description  Admin only. Assigns a specific role to a pending user
// @Security ApiKeyAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        input body domain.AssignRoleReq true "Assign role data"
// @Success      200  {object}  map[string]bool "success: true"
// @Failure      400  {object}  response.ErrorResponseMes
// @Failure      403  {object}  response.ErrorResponseMes "Access denied"
// @Failure      404  {object}  response.ErrorResponseMes "User not found"
// @Failure      500  {object}  response.ErrorResponseMes
// @Router       /admin/assign-role [post]
func (h *Handler) assignRole(c *gin.Context) {
	var input domain.AssignRoleReq
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.client.Timeout)
	defer cancel()

	resp, err := h.service.AssignRole(ctx, &auth.AssignRoleRequest{
		TargetUserId: input.TargetUserID,
		NewRole:      input.NewRole,
	})

	if err != nil {
		response.ParseGRPCError(c, h.log, err, "assign role")
		return
	}

	response.Success(c, gin.H{"success": resp.Success})
}

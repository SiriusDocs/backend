package auth

import (
	"context"
	"net/http"
	"strconv"

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

	resp, err := h.service.GetNewTokens(ctx,&auth.TokensRequest{
		RefreshToken: input.RefreshToken,
	})

	if err != nil {
		response.ParseGRPCError(c, h.log, err, "get new tokens")
		return
	}
	response.Success(c, gin.H{"access_token": resp.AccessToken, "refresh_token": resp.RefreshToken})
}

// @Summary      Get user profile
// @Description  Returns profile info including avatar presigned download URL
// @Tags         auth
// @Produce      json
// @Param        user_id path int true "User ID"
// @Success      200  {object}  response.Response{data=domain.GetProfileResponse}
// @Failure      400  {object}  response.ErrorResponseMes
// @Failure      404  {object}  response.ErrorResponseMes
// @Failure      500  {object}  response.ErrorResponseMes
// @Router       /auth/profile/{user_id} [get]
func (h *Handler) getProfile(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		response.ErrorResponse(c, http.StatusBadRequest, "invalid user_id")
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

	response.Success(c, domain.GetProfileResponse{
		UserId:    resp.UserId,
		Username:  resp.Username,
		Email:     resp.Email,
		Role:      resp.UserRole,
		AvatarUrl: resp.AvatarUrl,
	})
}

// @Summary      Get user avatar URL
// @Description  Returns presigned GET URL for user avatar
// @Tags         auth
// @Produce      json
// @Param        user_id path int true "User ID"
// @Success      200  {object}  response.Response{data=domain.GetAvatarResponse}
// @Failure      400  {object}  response.ErrorResponseMes
// @Failure      404  {object}  response.ErrorResponseMes
// @Failure      500  {object}  response.ErrorResponseMes
// @Router       /auth/avatar/{user_id} [get]
func (h *Handler) getAvatar(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		response.ErrorResponse(c, http.StatusBadRequest, "invalid user_id")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.client.Timeout)
	defer cancel()

	resp, err := h.service.GetAvatar(ctx, &auth.GetAvatarRequest{
		UserId: userID,
	})
	if err != nil {
		response.ParseGRPCError(c, h.log, err, "get user avatar")
		return
	}

	response.Success(c, domain.GetAvatarResponse{
		AvatarUrl: resp.AvatarUrl,
	})
}

// @Summary      Request avatar upload URL
// @Description  Generates presigned S3 PUT URL for uploading user avatar directly
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body domain.UploadAvatarUrlRequest true "Upload avatar request"
// @Success      200  {object}  response.Response{data=domain.UploadAvatarUrlResponse}
// @Failure      400  {object}  response.ErrorResponseMes
// @Failure      500  {object}  response.ErrorResponseMes
// @Router       /auth/avatar/upload-url [post]
func (h *Handler) generateAvatarUploadUrl(c *gin.Context) {
	var input domain.UploadAvatarUrlRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.client.Timeout)
	defer cancel()

	resp, err := h.service.GenerateAvatarUploadUrl(ctx, &auth.GenerateAvatarUploadUrlRequest{
		UserId:      input.UserId,
		ContentType: input.ContentType,
	})
	if err != nil {
		response.ParseGRPCError(c, h.log, err, "generate avatar upload url")
		return
	}

	response.Success(c, domain.UploadAvatarUrlResponse{
		UploadUrl: resp.UploadUrl,
		AvatarKey: resp.AvatarKey,
	})
}
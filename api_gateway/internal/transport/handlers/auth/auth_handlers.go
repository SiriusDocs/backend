package auth

import (
	"context"
	"net/http"

	_ "github.com/SiriusDocs/backend/api_gateway/internal/domain"
	"github.com/SiriusDocs/backend/api_gateway/internal/lib/response"
	"github.com/SiriusDocs/protos/gen/go/auth"
	"github.com/gin-gonic/gin"
)

// signUp регистрирует нового пользователя
// @Summary      User registration
// @Description  Creates a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body domain.RegisterRequest true "user data"
// @Success      200  {integer} map[string]string "user UUID"
// @Failure      400  {object}  map[string]string "validation error"
// @Router       /auth/sign-up [post]
func (h *Handler) signUp(c *gin.Context) {
	var input signUpInput

	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.client.Timeout)
	defer cancel()

	resp, err := h.service.Register(ctx, &auth.RegisterRequest{
		Email:    input.Email,
		Password: input.Password,
		Username: input.Username,
	})
	
	if err != nil {
		response.ParseGRPCError(c, h.log, err, "register user")
		return
	}
	response.Success(c, gin.H{"user_id": resp.UserId})
}
type signUpInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Username string `json:"username" binding:"required"`
}

func (h *Handler) signIn(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"id": 67,
	})
}

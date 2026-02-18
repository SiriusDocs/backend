package auth

import (
	"net/http"

	_ "github.com/SiriusDocs/backend/api_gateway/internal/domain"
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
	c.JSON(http.StatusOK, map[string]interface{}{
		"id": 67,
	})
}

func (h *Handler) signIn(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"id": 67,
	})
}

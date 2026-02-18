package auth

import (
	"net/http"

	_ "github.com/SiriusDocs/backend/api_gateway/internal/domain"
	"github.com/gin-gonic/gin"
)

// signUp регистрирует нового пользователя
// @Summary      Регистрация
// @Description  Создает новый аккаунт пользователя через auth-service
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body domain.RegisterRequest true "Данные пользователя"
// @Success      200  {integer} map[string]string "UUID пользователя"
// @Failure      400  {object}  map[string]string "Ошибка валидации"
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


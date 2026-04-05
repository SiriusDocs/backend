package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/SiriusDocs/backend/api_gateway/internal/lib/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	authorizationHeader = "Authorization"
	CtxUserID           = "userId"
	CtxUserRole         = "userRole"
)

type tokenClaims struct {
	jwt.RegisteredClaims
	UserId int64  `json:"user_id"`
	Role   string `json:"role"`
}

// UserIdentity проверяет JWT и кладет ID и Role в контекст Gin
func UserIdentity(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader(authorizationHeader)
		if header == "" {
			response.ErrorResponse(c, http.StatusUnauthorized, "empty auth header")
			return
		}

		headerParts := strings.Split(header, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			response.ErrorResponse(c, http.StatusUnauthorized, "invalid auth header")
			return
		}

		if len(headerParts[1]) == 0 {
			response.ErrorResponse(c, http.StatusUnauthorized, "token is empty")
			return
		}

		claims, err := parseToken(headerParts[1], secretKey)
		if err != nil {
			response.ErrorResponse(c, http.StatusUnauthorized, "invalid token")
			return
		}

		// Записываем данные в контекст
		c.Set(CtxUserID, claims.UserId)
		c.Set(CtxUserRole, claims.Role)
		c.Next()
	}
}

// RequireRole проверяет, есть ли у пользователя нужная роль
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get(CtxUserRole)
		if !exists {
			response.ErrorResponse(c, http.StatusUnauthorized, "role not found in context")
			return
		}

		role, ok := roleVal.(string)
		if !ok || role != requiredRole {
			response.ErrorResponse(c, http.StatusForbidden, "access denied: insufficient permissions")
			return
		}

		c.Next()
	}
}

func parseToken(accessToken, signingKey string) (*tokenClaims, error) {
	token, err := jwt.ParseWithClaims(accessToken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(signingKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return nil, errors.New("token claims are not of type *tokenClaims")
	}

	return claims, nil
}

// GetUserId — удобная функция для хэндлеров, чтобы доставать ID
func GetUserId(c *gin.Context) (int64, error) {
	idVal, exists := c.Get(CtxUserID)
	if !exists {
		return 0, errors.New("user id not found")
	}
	
	id, ok := idVal.(int64)
	if !ok {
		return 0, errors.New("user id is of invalid type")
	}
	
	return id, nil
}
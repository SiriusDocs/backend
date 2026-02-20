package response

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponseMes struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"invalid request body"`
}

const (
	StatusSuccess = "success"
	StatusError   = "error"
)

// Success отправляет успешный ответ
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Status: StatusSuccess,
		Data:   data,
	})
}

// Error отправляет кастомную ошибку с заданным кодом
func ErrorResponse(c *gin.Context, code int, message string) {
	c.AbortWithStatusJSON(code, Response{
		Status:  StatusError,
		Message: message,
	})
}

// ValidationError разбирает ошибки тегов binding:"required,email..."
func ValidationError(c *gin.Context, err error) {
	var errMsgs []string
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is invalid (%s)", e.Field(), e.Tag()))
		}
	} else {
		errMsgs = append(errMsgs, "invalid request body")
	}

	c.AbortWithStatusJSON(http.StatusBadRequest, Response{
		Status:  StatusError,
		Message: strings.Join(errMsgs, ", "),
	})
}

// ParseGRPCError — наш "переводчик" с языка gRPC на HTTP
func ParseGRPCError(c *gin.Context, log *slog.Logger, err error, action string) {
	st, ok := status.FromError(err)
	if !ok {
		log.Error("non-grpc error", slog.String("action", action), slog.Any("err", err))
		ErrorResponse(c, http.StatusInternalServerError, "internal server error")
		return
	}

	log.Warn("grpc error", 
		slog.String("action", action), 
		slog.Int("code", int(st.Code())), 
		slog.String("msg", st.Message()),
	)

	switch st.Code() {
	case codes.InvalidArgument:
		ErrorResponse(c, http.StatusBadRequest, st.Message())
	case codes.Unauthenticated:
		ErrorResponse(c, http.StatusUnauthorized, "unauthorized access")
	case codes.PermissionDenied:
		ErrorResponse(c, http.StatusForbidden, "access denied")
	case codes.AlreadyExists:
		ErrorResponse(c, http.StatusConflict, st.Message())
	case codes.NotFound:
		ErrorResponse(c, http.StatusNotFound, st.Message())
	case codes.DeadlineExceeded:
		ErrorResponse(c, http.StatusGatewayTimeout, "service timeout")
	default:
		ErrorResponse(c, http.StatusInternalServerError, "internal service error")
	}
}
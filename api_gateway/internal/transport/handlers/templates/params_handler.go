package templates

import (
	"context"

	"github.com/SiriusDocs/backend/api_gateway/internal/domain"
	"github.com/SiriusDocs/backend/api_gateway/internal/lib/response"
	"github.com/gin-gonic/gin"
)

// @Summary      Создание параметров шаблона (схемы таблицы)
// @Description  Принимает ID задачи и список переменных (имя: тип) для создания структуры данных шаблона.
// @Description	// Строки
// @Description	"string": "TEXT",
// @Description	"text":   "TEXT",
// @Description
// @Description	// Числа
// @Description	"int":     "INTEGER",
// @Description	"integer": "INTEGER",
// @Description	"number":  "INTEGER",
// @Description
// @Description	// Дробные
// @Description	"float":  "DOUBLE PRECISION",
// @Description	"double": "DOUBLE PRECISION",
// @Description
// @Description	// Булевы
// @Description	"bool":    "BOOLEAN",
// @Description	"boolean": "BOOLEAN",
// @Description
// @Description	// Даты
// @Description	"date":      "DATE",
// @Description	"datetime":  "TIMESTAMP",
// @Description	"timestamp": "TIMESTAMP",
// @Tags         templates
// @Accept       json
// @Produce      json
// @Param        input body domain.CreateParamsRequest true "Конфигурация параметров"
// @Success      200  {object}  response.Response{data=domain.CreateParamsResponse} "Успешное создание"
// @Failure      400  {object}  response.ErrorResponseMes "Ошибка валидации входных данных"
// @Failure      500  {object}  response.ErrorResponseMes "Внутренняя ошибка сервера"
// @Router       /temp/params/create [post]
func (h *Handler) createParams(c *gin.Context) {
	var input domain.CreateParamsRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.client.Timeout)
	defer cancel()

	tmplID, err := h.service.CreateParams(ctx, input.TaskID, input.Params)
	if err != nil {
		response.ParseGRPCError(c, h.log, err, "template.CreateParams")
		return
	}

	response.Success(c, domain.CreateParamsResponse{
        TemplateID: tmplID,
    })
}
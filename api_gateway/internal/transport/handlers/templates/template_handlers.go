package templates

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/SiriusDocs/backend/api_gateway/internal/lib/response"
	"github.com/SiriusDocs/protos/gen/go/templates"
	"github.com/gin-gonic/gin"
)

type UploadDTO struct {
	TaskID string `json:"task_id"`
}

type StatusDTO struct {
	Status string `json:"status"`
}

// @Summary      Загрузка файла для обработки
// @Description  Принимает файл, отправляет его в gRPC сервис и возвращает ID задачи.
// @Tags         templates
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "Файл для парсинга"
// @Success      202  {object}  domain.UploadFileResponse
// @Failure      400  {object}  response.ErrorResponseMes
// @Failure      500  {object}  response.ErrorResponseMes
// @Router       /temp/files/upload [post]
func (h *Handler) uploadFile(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "no file provided")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "failed to open file")
		return
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.client.Timeout)
	defer cancel()

	taskID, err := h.service.UploadFile(
		ctx,
		fileHeader.Filename,
		file,
	)
	if err != nil {
		response.ParseGRPCError(c, h.log, err, "template.UploadAndProcessFile")
		return
	}

	c.JSON(http.StatusAccepted, response.Response{
		Status: response.StatusSuccess,
		Data: UploadDTO{
			TaskID: taskID,
		},
	})
}

// @Summary      Проверка статуса задачи
// @Description  Возвращает статус обработки. Если готово — возвращает результат парсинга.
// @Tags         templates
// @Produce      json
// @Param        task_id  path      string  true  "ID задачи"
// @Success      200  {object}  domain.TaskResultResponse  "Результат парсинга"
// @Success      202  {object}  domain.TaskStatusResponse  "Задача в процессе"
// @Failure      400  {object}  response.ErrorResponseMes
// @Failure      404  {object}  response.ErrorResponseMes
// @Failure      500  {object}  response.ErrorResponseMes
// @Router       /temp/files/status/{task_id} [get]
func (h *Handler) checkStatus(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "task_id is required")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.client.Timeout)
	defer cancel()

	statusResp, err := h.service.CheckStatus(ctx, &templates.StatusRequest{
		TaskId: taskID,
	})

	if err != nil {
		response.ParseGRPCError(c, h.log, err, "template.CheckTaskStatus")
		return
	}

	switch statusResp.Status {
	case "done":
        if !json.Valid(statusResp.ResultData) {
            response.ErrorResponse(c, http.StatusInternalServerError, "invalid json result from service")
            return
        }

        c.JSON(http.StatusOK, response.Response{
            Status: response.StatusSuccess,
            Data:   json.RawMessage(statusResp.ResultData),
        })
        return


	case "error":
		response.ErrorResponse(c, http.StatusInternalServerError, "processing failed")
		return

	default:
		// pending / processing
		c.JSON(http.StatusAccepted, response.Response{
			Status: response.StatusSuccess,
			Data: StatusDTO{
				Status: statusResp.Status,
			},
		})
		return
	}
}

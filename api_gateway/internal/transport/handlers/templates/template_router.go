package templates

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/SiriusDocs/backend/api_gateway/internal/config"
	"github.com/SiriusDocs/backend/api_gateway/internal/domain"
	"github.com/SiriusDocs/backend/api_gateway/internal/lib/response"
	"github.com/SiriusDocs/backend/api_gateway/internal/services"
	"github.com/SiriusDocs/protos/gen/go/templates"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	log     *slog.Logger
	service services.TempService
	client  *config.GRPCClient
}

func New(log *slog.Logger, service services.TempService, client *config.GRPCClient) *Handler {
	return &Handler{
		log:     log,
		service: service,
		client:  client,
	}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	tempGroup := rg.Group("/temp")
	{
		fileGroup := tempGroup.Group("/files")
		{
			fileGroup.POST("/upload", h.uploadFile)
			fileGroup.GET("/status/:task_id", h.checkStatus)
		}
		paramsGroup := tempGroup.Group("/params")
		{
			paramsGroup.POST("/create", h.createParams)
		}
		tempGroup.GET("/list", h.listTemplates)
	}
}

// @Summary      Получение списка всех шаблонов
// @Description  Возвращает список всех шаблонов, пагинация контролируется через query-параметры `per_page` и `page`
// @Tags         templates
// @Produce      json
// @Router       /temp/list [get]
func (h *Handler) listTemplates(c *gin.Context) {
	per_page_query := c.Query("per_page")
	if per_page_query == "" {
		per_page_query = "10"
	}

	per_page, err := strconv.ParseInt(per_page_query, 10, 64)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "per_page should be int")
	}

	page_query := c.Query("page")
	if page_query == "" {
		page_query = "0"
	}

	page, err := strconv.ParseInt(page_query, 10, 64)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "page should be int")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.client.Timeout)
	defer cancel()

	resp, err := h.service.ListTemplates(ctx, &templates.ListTemplatesRequest{
		PerPage: per_page,
		Page:    page,
	})

	if err != nil {
		response.ParseGRPCError(c, h.log, err, "template.ListTemplates")
		return
	}

	templateList := make([]domain.Template, len(resp.Templates))
	for _, t := range resp.Templates {
		templateList = append(templateList, domain.Template{
			ID:          t.Id,
			Name:        t.Name,
			Description: t.Description,
			Vars:        t.Vars,
			CreatedAt:   t.CreatedAt.AsTime(),
			UpdatedAt:   t.UpdatedAt.AsTime(),
		})
	}

	c.JSON(http.StatusAccepted, response.Response{
		Status: response.StatusSuccess,
		Data: domain.ListTemplatesResponse{
			Templates: templateList,
		},
	})
}


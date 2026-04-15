package files

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/SiriusDocs/backend/api_gateway/internal/domain"
	"github.com/SiriusDocs/backend/api_gateway/internal/lib/response"
	"github.com/SiriusDocs/protos/gen/go/file"
	"github.com/gin-gonic/gin"
)

// resolveBucket маппит context на bucket с валидацией
func (h *Handler) resolveBucket(context string) (string, error) {
	mapping := map[string]string{
		"avatar":    "auth",
		"document":  "documents",
		"template":  "templates",
	}

	bucket, ok := mapping[context]
	if !ok {
		return "", fmt.Errorf("unknown context: %s", context)
	}
	return bucket, nil
}

// ==================== Upload ====================

// @Summary      Загрузка файла
// @Description  Загружает файл в S3 хранилище. Bucket определяется по параметру context
// @Tags         files
// @Accept       multipart/form-data
// @Produce      json
// @Param        file     formData  file    true   "Файл для загрузки"
// @Param        context  query     string  true   "Контекст файла (avatar, passport, document, template)"
// @Success      202  {object}  response.Response{data=domain.UploadFileResponse}
// @Failure      400  {object}  response.ErrorResponseMes
// @Failure      500  {object}  response.ErrorResponseMes
// @Router       /files/upload [post]
func (h *Handler) uploadFile(c *gin.Context) {
	// Получаем context из query параметра
	contextParam := c.Query("context")
	if contextParam == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "context parameter is required")
		return
	}

	// Валидация и маппинг context -> bucket
	bucket, err := h.resolveBucket(contextParam)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

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

	// Получаем user_id и roles из контекста (middleware auth)
	userID := c.GetString("user_id")
	roles := c.GetStringSlice("roles")

	// Формируем метаданные
	meta := &files.UploadMeta{
		Filename:    fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		Size:        fileHeader.Size,
		Bucket:      bucket, // <- Определили через валидацию
		UserId:      userID,
		Roles:       roles,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.client.Timeout)
	defer cancel()

	// Вызываем сервис
	resp, err := h.service.UploadFile(ctx, meta, file)
	if err != nil {
		response.ParseGRPCError(c, h.log, err, "file.UploadFile")
		return
	}

	c.JSON(http.StatusAccepted, response.Response{
		Status: response.StatusSuccess,
		Data: domain.UploadFilesResponse{
			FileID:   resp.FileId,
			UploadID: resp.UploadId,
		},
	})
}

// ==================== Download ====================

// @Summary      Скачивание файла
// @Description  Получает файл из хранилища (стрим). Файл отображается в браузере, не скачивается автоматически
// @Tags         files
// @Produce      application/octet-stream
// @Param        file_id  path  string  true  "ID файла"
// @Success      200  {file}  binary
// @Failure      400  {object}  response.ErrorResponseMes
// @Failure      403  {object}  response.ErrorResponseMes
// @Failure      404  {object}  response.ErrorResponseMes
// @Failure      500  {object}  response.ErrorResponseMes
// @Router       /files/download/{file_id} [get]
func (h *Handler) downloadFile(c *gin.Context) {
	fileID := c.Param("file_id")
	if fileID == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "file_id is required")
		return
	}

	// Получаем user_id и roles из контекста
	userID := c.GetString("user_id")
	roles := c.GetStringSlice("roles")

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.client.Timeout)
	defer cancel()

	// Сначала получаем метаданные для Content-Type и filename
	meta, err := h.service.GetFileMeta(ctx, fileID)
	if err != nil {
		response.ParseGRPCError(c, h.log, err, "file.GetFileMeta")
		return
	}

	// Получаем stream reader
	reader, err := h.service.DownloadFile(ctx, fileID, userID, roles)
	if err != nil {
		response.ParseGRPCError(c, h.log, err, "file.DownloadFile")
		return
	}
	defer reader.Close()

	// Устанавливаем заголовки
	c.Header("Content-Type", meta.ContentType)
	// inline - браузер попытается отобразить, а не скачать
	// для принудительного скачивания замени на "attachment"
	c.Header("Content-Disposition", "inline; filename=\""+meta.Filename+"\"")

	// Стримим файл напрямую в response
	_, err = io.Copy(c.Writer, reader)
	if err != nil {
		h.log.Error("failed to stream file", "error", err)
		return
	}
}

// ==================== Upload Status ====================

// @Summary      Проверка статуса загрузки
// @Description  Возвращает прогресс загрузки файла в процентах
// @Tags         files
// @Produce      json
// @Param        upload_id  path  string  true  "ID загрузки"
// @Success      200  {object}  response.Response{data=domain.UploadStatusResponse}
// @Failure      400  {object}  response.ErrorResponseMes
// @Failure      404  {object}  response.ErrorResponseMes
// @Failure      500  {object}  response.ErrorResponseMes
// @Router       /files/status/{upload_id} [get]
func (h *Handler) getUploadStatus(c *gin.Context) {
	uploadID := c.Param("upload_id")
	if uploadID == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "upload_id is required")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.client.Timeout)
	defer cancel()

	progress, completed, err := h.service.GetUploadStatus(ctx, uploadID)
	if err != nil {
		response.ParseGRPCError(c, h.log, err, "file.GetUploadStatus")
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Status: response.StatusSuccess,
		Data: domain.UploadStatusResponse{
			ProgressPercent: progress,
			Completed:       completed,
		},
	})
}

// ==================== File Meta ====================

// @Summary      Получение метаданных файла
// @Description  Возвращает информацию о файле (имя, размер, тип и т.д.)
// @Tags         files
// @Produce      json
// @Param        file_id  path  string  true  "ID файла"
// @Success      200  {object}  response.Response{data=domain.FileMetaResponse}
// @Failure      400  {object}  response.ErrorResponseMes
// @Failure      404  {object}  response.ErrorResponseMes
// @Failure      500  {object}  response.ErrorResponseMes
// @Router       /files/meta/{file_id} [get]
func (h *Handler) getFileMeta(c *gin.Context) {
	fileID := c.Param("file_id")
	if fileID == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "file_id is required")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.client.Timeout)
	defer cancel()

	meta, err := h.service.GetFileMeta(ctx, fileID)
	if err != nil {
		response.ParseGRPCError(c, h.log, err, "file.GetFileMeta")
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Status: response.StatusSuccess,
		Data: domain.FileMetaResponse{
			FileID:      meta.FileId,
			Filename:    meta.Filename,
			ContentType: meta.ContentType,
			Size:        meta.Size,
			Bucket:      meta.Bucket,
		},
	})
}

// ==================== Admin: Create Bucket ====================

// @Summary      Создание bucket (admin)
// @Description  Создаёт новый bucket в S3 хранилище
// @Tags         files-admin
// @Accept       json
// @Produce      json
// @Param        request  body  domain.CreateBucketRequest  true  "Название bucket"
// @Success      201  {object}  response.Response
// @Failure      400  {object}  response.ErrorResponseMes
// @Failure      403  {object}  response.ErrorResponseMes
// @Failure      500  {object}  response.ErrorResponseMes
// @Router       /files/admin/buckets [post]
func (h *Handler) createBucket(c *gin.Context) {
	var req domain.CreateBucketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.BucketName == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "bucket_name is required")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.client.Timeout)
	defer cancel()

	err := h.service.CreateBucket(ctx, req.BucketName)
	if err != nil {
		response.ParseGRPCError(c, h.log, err, "file.CreateBucket")
		return
	}

	c.JSON(http.StatusCreated, response.Response{
		Status: response.StatusSuccess,
		Data:   nil,
	})
}

// ==================== Admin: Delete File ====================

// @Summary      Удаление файла (admin)
// @Description  Удаляет файл из хранилища
// @Tags         files-admin
// @Produce      json
// @Param        file_id  path  string  true  "ID файла"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.ErrorResponseMes
// @Failure      403  {object}  response.ErrorResponseMes
// @Failure      404  {object}  response.ErrorResponseMes
// @Failure      500  {object}  response.ErrorResponseMes
// @Router       /files/admin/{file_id} [delete]
func (h *Handler) deleteFile(c *gin.Context) {
	fileID := c.Param("file_id")
	if fileID == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "file_id is required")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.client.Timeout)
	defer cancel()

	err := h.service.DeleteFile(ctx, fileID)
	if err != nil {
		response.ParseGRPCError(c, h.log, err, "file.DeleteFile")
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Status: response.StatusSuccess,
		Data:   nil,
	})
}

// ==================== Admin: List Files ====================

// @Summary      Список файлов в bucket (admin)
// @Description  Возвращает список всех файлов в указанном bucket
// @Tags         files-admin
// @Produce      json
// @Param        bucket  path  string  true  "Название bucket"
// @Success      200  {object}  response.Response{data=domain.ListFilesResponse}
// @Failure      400  {object}  response.ErrorResponseMes
// @Failure      403  {object}  response.ErrorResponseMes
// @Failure      500  {object}  response.ErrorResponseMes
// @Router       /files/admin/list/{bucket} [get]
func (h *Handler) listFiles(c *gin.Context) {
	bucket := c.Param("bucket")
	if bucket == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "bucket is required")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.client.Timeout)
	defer cancel()

	files, err := h.service.ListFiles(ctx, bucket)
	if err != nil {
		response.ParseGRPCError(c, h.log, err, "file.ListFiles")
		return
	}

	// Конвертируем proto в domain
	fileList := make([]domain.FileMetaResponse, 0, len(files))
	for _, f := range files {
		fileList = append(fileList, domain.FileMetaResponse{
			FileID:      f.FileId,
			Filename:    f.Filename,
			ContentType: f.ContentType,
			Size:        f.Size,
			Bucket:      f.Bucket,
		})
	}

	c.JSON(http.StatusOK, response.Response{
		Status: response.StatusSuccess,
		Data: domain.ListFilesResponse{
			Files: fileList,
		},
	})
}
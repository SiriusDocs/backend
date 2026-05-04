package templates

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/SiriusDocs/backend/template_service/internal/domain"
	"github.com/SiriusDocs/backend/template_service/internal/storage"
	"github.com/SiriusDocs/backend/template_service/internal/storage/postgres"
)

type ParamsService struct {
	log       *slog.Logger
	store     storage.TemplateOperations
	taskStore storage.TaskOperations
}

func NewParamsService(log *slog.Logger, store storage.TemplateOperations, taskStore storage.TaskOperations) *ParamsService {
	return &ParamsService{
		log:       log,
		store:     store,
		taskStore: taskStore,
	}
}

func (s *ParamsService) CreateTemplate(ctx context.Context, taskID string, params map[string]string) (string, error) {
	const op = "services.templates.params_service.CreateTemplate"

	log := s.log.With(
		slog.String("op", op),
		slog.String("task_id", taskID),
	)

	// 1. Получаем задачу
	task, err := s.taskStore.GetTask(ctx, taskID)
	if err != nil {
		return "", err
	}

	// Проверяем, что парсинг вообще завершился
	if task.FileStatus != domain.TaskStatusDone {
		return "", domain.Validation(op,
			fmt.Sprintf("task is not ready for parameter assignment (current status: %s)", task.FileStatus))
	}

	// 2. Достаем список полей, которые нашел парсер
	var parserFields []string
	if err := json.Unmarshal(task.ResultData, &parserFields); err != nil {
		log.Error("failed to unmarshal parser results", slog.String("error", err.Error()))
		return "", domain.Internal(op, "failed to read task data", err)
	}

	// 3. СРАВНЕНИЕ ПОЛЕЙ
	// Мы должны убедиться, что keys(params) == parserFields
	if err := s.validateFieldsMatch(parserFields, params); err != nil {
		log.Warn("fields validation failed", slog.String("error", err.Error()))
		return "", err
	}

	// 4. Валидация: типы данных и имена колонок (бизнес-валидация В СЕРВИСЕ)
	if err := s.validateParamValues(params); err != nil {
		log.Warn("params validation failed", slog.String("error", err.Error()))
		return "", err // Уже Validation
	}

	// 5. Создаем таблицу шаблона
	if err := s.store.CreateTemplateTable(ctx, taskID, params); err != nil {
		log.Error("failed to create dynamic table", slog.String("error", err.Error()))
		return "", err
	}

	// 6. Добавляем шаблон в таблицу templates
	if err := s.store.AddTemplate(ctx, taskID, task.FileName, parserFields); err != nil {
		log.Error("failed to insert template", slog.String("error", err.Error()))
		return "", err
	}

	// 7. Удаляем задачу (Cleanup)
	// Если удаление упадет, это не критично для бизнеса (таблица-то создалась),
	// поэтому мы просто логируем ошибку, но не фейлим запрос пользователю.
	if err := s.taskStore.DeleteTask(ctx, taskID); err != nil {
		log.Error("failed to cleanup task after template creation", slog.String("error", err.Error()))
		// Мы НЕ возвращаем error здесь, так как основная цель достигнута
	} else {
		log.Info("task deleted successfully", slog.String("task_id", taskID))
	}

	log.Info("dynamic template created successfully")

	return taskID, nil
}

func (s *ParamsService) ListTemplates(ctx context.Context, perPage int, page int) ([]domain.Template, error) {
	const op = "services.templates.params_service.ListTemplates"
	return s.store.ListTemplates(ctx, perPage, page*perPage)
}

// validateFieldsMatch проверяет, что набор полей от парсера совпадает с набором полей от пользователя
func (s *ParamsService) validateFieldsMatch(parserFields []string, userParams map[string]string) error {
	const op = "services.templates.params_service.validateFieldsMatch"

	// 1. Проверка количества
	if len(parserFields) != len(userParams) {
		return domain.Validation(op,
			fmt.Sprintf("field count mismatch: expected %d fields, got %d",
				len(parserFields), len(userParams)))
	}

	// 2. Проверка наличия каждого поля
	// Создаем мапу для быстрого поиска, или просто пробегаем циклом
	for _, field := range parserFields {
		if _, ok := userParams[field]; !ok {
			return domain.Validation(op,
				fmt.Sprintf("missing type for field '%s' found in document", field))
		}
	}

	// Обратная проверка не нужна, так как длины совпадают.
	// Если длины равны и все поля из парсера есть в юзер-мапе, значит лишних полей в юзер-мапе нет.

	return nil
}

// validateParamValues проверяет имена колонок и типы данных ДО передачи в storage
func (s *ParamsService) validateParamValues(params map[string]string) error {
	const op = "services.templates.params_service.validateParamValues"

	for colName, userType := range params {
		// Проверка имени колонки
		if !postgres.ValidColName.MatchString(colName) {
			return domain.Validation(op,
				fmt.Sprintf("invalid field name '%s': only letters, digits, and underscores are allowed", colName))
		}

		// Проверка типа данных
		if _, ok := postgres.AllowedTypes[strings.ToLower(userType)]; !ok {
			return domain.Validation(op,
				fmt.Sprintf("unsupported data type '%s' for field '%s'", userType, colName))
		}
	}

	return nil
}

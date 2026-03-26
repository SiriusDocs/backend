package templates

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/SiriusDocs/backend/template_service/internal/config"
	"github.com/SiriusDocs/backend/template_service/internal/domain"
	"github.com/SiriusDocs/backend/template_service/internal/storage"
	"github.com/google/uuid"
)

var (
	magicRTF = []byte("{\\rtf")               // Сигнатура RTF
	magicZIP = []byte{0x50, 0x4B, 0x03, 0x04} // Сигнатура ZIP (docx, odt)
)

var VarRegexRTF = regexp.MustCompile(`{\\field\\fldlock{\\\*\\fldinst  DOCVARIABLE (.*?) }{\\fldrslt (.*?)}}`)

type TemplateService struct {
	log   *slog.Logger
	store storage.TaskOperations
	cfg   config.TasksConfig
}

func NewTasksService(log *slog.Logger, store storage.TaskOperations, cfg config.TasksConfig) *TemplateService {
	return &TemplateService{
		log:   log,
		store: store,
		cfg:   cfg,
	}
}

// UploadAndStartTask создает задачу, валидирует файл и запускает фоновую обработку
func (s *TemplateService) UploadAndStartTask(ctx context.Context, filename string, fileData []byte) (string, error) {
	const op = "TemplateService.UploadAndStartTask"

	// Обогащаем логгер контекстом операции и именем файла
	log := s.log.With(
		slog.String("op", op),
		slog.String("filename", filename),
		slog.Int("file_size", len(fileData)),
	)

	// 1. Валидация файла
	if err := s.validateData(filename, fileData); err != nil {
		log.Warn("file validation failed", slog.String("error", err.Error()))
		return "", err
	}

	// 2. Генерация ID
	taskID := uuid.New().String()

	// 3. Создание записи в БД
	if err := s.store.CreateTask(ctx, taskID, filename); err != nil {
		log.Error("failed to create task in db", slog.String("error", err.Error()))
		return "", err
	}

	log.Info("task created successfully", slog.String("task_id", taskID))

	// 4. Запуск парсера в фоне
	go s.parseRTF(taskID, fileData)

	return taskID, nil
}

// CheckTaskStatus проверяет состояние задачи
func (s *TemplateService) CheckTaskStatus(ctx context.Context, taskID string) (string, []string, error) {
	const op = "TemplateService.CheckTaskStatus"

	log := s.log.With(
		slog.String("op", op),
		slog.String("task_id", taskID),
	)

	// 1. Получение задачи из БД
	task, err := s.store.GetTask(ctx, taskID)
	if err != nil {
		log.Warn("failed to get task", slog.String("error", err.Error()))
		return "", nil, err // NotFound или Internal — уже типизировано
	}

	// СЦЕНАРИЙ 1: Задача завершилась с ошибкой (валидация, парсер упал)
	if task.FileStatus == domain.TaskStatusError {
		errMsg := "unknown error"
		if task.ErrorMessage != nil {
			errMsg = *task.ErrorMessage
		}
		// Возвращаем статус "error" И саму ошибку
		return domain.TaskStatusError, nil, domain.Validation(op, errMsg)
	}
	// СЦЕНАРИЙ 2: Задача еще не готова
	if task.FileStatus != domain.TaskStatusDone {
		return task.FileStatus, nil, nil
	}

	// 3. Распаковка результата
	var names []string
	if len(task.ResultData) > 0 {
		if err := json.Unmarshal(task.ResultData, &names); err != nil {
			log.Error("data corruption: failed to unmarshal result data", slog.String("error", err.Error()))
			return "", nil, domain.Internal(op, "failed to read task results", err)
		}
	}

	return task.FileStatus, names, nil
}

// Парсинг RTF файлов - возвращает список id'шников всех переменных в документе
func (s *TemplateService) parseRTF(taskID string, data []byte) {
	const op = "TemplateService.parseRTF"

	// Создаем логгер для этой горутины
	log := s.log.With(
		slog.String("op", op),
		slog.String("task_id", taskID),
	)

	// Background context, т.к. родительский контекст запроса уже завершен
	ctx := context.Background()

	log.Debug("start processing task")

	// 1. Ставим статус Processing
	if err := s.store.SetStatus(ctx, taskID, domain.TaskStatusProcessing); err != nil {
		log.Error("failed to set status processing", slog.String("error", err.Error()))
		return
	}

	// --- САМ ПАРСЕР ---
	// TODO: подумать, может стоит возвращать map'у
	//       где значения это описание переменных, а не список
	keys := make([]string, 0)
	vars := make(map[string]string)

	// TODO: полуболванка, стоит перерассмотреть как мы это делаем
	matches := VarRegexRTF.FindAllSubmatch(data, -1)
	for _, v := range matches {
		id := string(v[1])
		desc := string(v[2])

		vars[id] = desc
		keys = append(keys, id)
	}

	resultJSON, err := json.Marshal(keys)
	if err != nil {
		log.Error("failed to marshal parse results", slog.String("error", err.Error()))

		// Пытаемся записать ошибку в БД
		if err := s.store.SetError(ctx, taskID, "internal parser error"); err != nil {
			log.Error("failed to save error status to db", slog.String("error", err.Error()))
		}
		return
	}
	// ------------------------

	// 2. Сохраняем результат
	if err := s.store.SetResult(ctx, taskID, resultJSON); err != nil {
		log.Error("failed to set task result", slog.String("error", err.Error()))
		return
	}

	log.Info("task processing completed successfully")
}

// processMock — заглушка работы парсера (запускается в горутине)
func (s *TemplateService) processMock(taskID string, _ []byte) {
	const op = "TemplateService.processMock"

	// Создаем логгер для этой горутины
	log := s.log.With(
		slog.String("op", op),
		slog.String("task_id", taskID),
	)

	// Background context, т.к. родительский контекст запроса уже завершен
	ctx := context.Background()

	log.Debug("start processing task")

	// 1. Ставим статус Processing
	if err := s.store.SetStatus(ctx, taskID, domain.TaskStatusProcessing); err != nil {
		log.Error("failed to set status processing", slog.String("error", err.Error()))
		return
	}

	// --- ЗАГЛУШКА ПАРСЕРА ---
	// Имитация работы
	sleepTime := time.Duration(rand.Intn(3000)+2000) * time.Millisecond
	time.Sleep(sleepTime)

	// Имитация данных
	mockNames := []string{
		"contract_number",
		"client_name",
		"start_date",
		fmt.Sprintf("random_var_%d", rand.Intn(100)),
	}

	resultJSON, err := json.Marshal(mockNames)
	if err != nil {
		log.Error("failed to marshal mock results", slog.String("error", err.Error()))

		// Пытаемся записать ошибку в БД
		if err := s.store.SetError(ctx, taskID, "internal parser error"); err != nil {
			log.Error("failed to save error status to db", slog.String("error", err.Error()))
		}
		return
	}
	// ------------------------

	// 2. Сохраняем результат
	if err := s.store.SetResult(ctx, taskID, resultJSON); err != nil {
		log.Error("failed to set task result", slog.String("error", err.Error()))
		return
	}

	log.Info("task processing completed successfully")
}

// validateData проверяет размер, расширение и целостность файла
func (s *TemplateService) validateData(filename string, data []byte) error {
	const op = "TemplateService.validateData"

	// 1. Проверка размера
	if len(data) < s.cfg.MinFileSize {
		return domain.Validation(op, "file is too small or empty")
	}
	if len(data) > s.cfg.MaxFileSize {
		return domain.Validation(op, fmt.Sprintf("file size exceeds limit of %d MB", s.cfg.MaxFileSize/(1024*1024)))
	}

	// 2. Проверка расширения
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".rtf":
		return s.validateRTF(data)
	case ".docx", ".odt":
		return s.validateZipBased(data, ext)
	default:
		return domain.Validation(op, fmt.Sprintf("unsupported file format: %s (allowed: .rtf, .docx, .odt)", ext))
	}
}

// validateRTF проверяет сигнатуру и базовую структуру RTF
func (s *TemplateService) validateRTF(data []byte) error {
	const op = "TemplateService.validateRTF"

	if !bytes.HasPrefix(data, magicRTF) {
		return domain.Validation(op, "file content does not match RTF format")
	}

	trimmed := bytes.TrimRight(data, " \r\n\t\x00")
	if len(trimmed) == 0 || trimmed[len(trimmed)-1] != '}' {
		return domain.Validation(op, "corrupted RTF file")
	}

	return nil
}

// validateZipBased проверяет файлы, являющиеся ZIP-архивами (.docx, .odt)
func (s *TemplateService) validateZipBased(data []byte, ext string) error {
	const op = "TemplateService.validateZipBased"

	if !bytes.HasPrefix(data, magicZIP) {
		return domain.Validation(op, fmt.Sprintf("file content does not match %s format", ext))
	}

	reader := bytes.NewReader(data)
	if _, err := zip.NewReader(reader, int64(len(data))); err != nil {
		return domain.Validation(op, fmt.Sprintf("corrupted %s file", ext))
	}

	return nil
}

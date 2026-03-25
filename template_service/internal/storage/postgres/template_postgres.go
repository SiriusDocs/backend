package postgres

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/SiriusDocs/backend/template_service/internal/domain"
	"github.com/jmoiron/sqlx"
)

// Допустимые типы данных
var AllowedTypes = map[string]string{
	"string":    "TEXT",
	"text":      "TEXT",
	"int":       "INTEGER",
	"integer":   "INTEGER",
	"float":     "DOUBLE PRECISION",
	"double":    "DOUBLE PRECISION",
	"bool":      "BOOLEAN",
	"boolean":   "BOOLEAN",
	"date":      "DATE",
	"datetime":  "TIMESTAMP",
	"timestamp": "TIMESTAMP",
}

// Регулярка для имен колонок (только буквы, цифры и подчеркивания)
var ValidTableName = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
var ValidColName = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

type TemplateOperationsPostgres struct {
	db *sqlx.DB
}

func NewTemplateOperationsPostgres(db *sqlx.DB) *TemplateOperationsPostgres {
	return &TemplateOperationsPostgres{
		db: db,
	}
}

func (r *TemplateOperationsPostgres) CreateTemplateTable(ctx context.Context, templateID string, columns map[string]string) error {
	const op = "storage.postgres.template_postgres.CreateTemplateTable"

	sanitizedID := strings.ReplaceAll(templateID, "-", "_")
	// ВАЖНО: Проверяем, что ID безопасен
	if !ValidTableName.MatchString(sanitizedID) {
		// Это не ошибка пользователя, templateID генерируется сервером
		return domain.Internal(op, "invalid template ID format", nil)
	}

	tableName := fmt.Sprintf("template_%s", sanitizedID)

	// 2. Сборка колонок
	// Аллоцируем слайс сразу нужного размера (+1 для row_id)
	columnDefs := make([]string, 0, len(columns)+1)

	// Используем кавычки для защиты от зарезервированных слов
	columnDefs = append(columnDefs, "\"row_id\" UUID PRIMARY KEY DEFAULT gen_random_uuid()")

	keys := make([]string, 0, len(columns))
	for k := range columns {
		keys = append(keys, k)
	}
	sort.Strings(keys)

		for _, colName := range keys {
		userType := columns[colName]

		// Защитный слой (основная валидация — в сервисе)
		if !ValidColName.MatchString(colName) {
			return domain.Validation(op, fmt.Sprintf("invalid column name: '%s'", colName))
		}

		sqlType, ok := AllowedTypes[strings.ToLower(userType)]
		if !ok {
			return domain.Validation(op, fmt.Sprintf("unsupported data type '%s' for column '%s'", userType, colName))
		}

		columnDefs = append(columnDefs, fmt.Sprintf(`"%s" %s`, colName, sqlType))
	}

	// 3. Сборка запроса
	// Имя таблицы тоже берем в кавычки для надежности (хотя regex выше защищает структуру)
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS \"%s\" (%s);",
		tableName,
		strings.Join(columnDefs, ", "),
	)

	// 4. Выполнение
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return domain.Internal(op, "failed to create table", err)
	}

	return nil
}

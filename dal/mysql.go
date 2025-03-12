package dal

import (
	"ac/bootstrap/logger"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Entity interface {
	// model.UserChange | model.UserChangeRemind | model.Task | model.User | model.Revoke | model.Transfer | model.TaskV2 | model.SubtaskV2
}

type Repository[T Entity] interface {
	Insert(ctx context.Context, db *gorm.DB, newValue *T) error
	BatchInsert(ctx context.Context, db *gorm.DB, valuesToAdd []*T, batchSize int) error
	Update(ctx context.Context, db *gorm.DB, newValue *T, funcs ...func(db *gorm.DB) *gorm.DB) error
	UpdateWithMap(ctx context.Context, db *gorm.DB, newValue map[string]interface{}, funcs ...func(db *gorm.DB) *gorm.DB) error
	Query(ctx context.Context, db *gorm.DB, funcs ...func(db *gorm.DB) *gorm.DB) (*T, error)
	QueryList(ctx context.Context, db *gorm.DB, funcs ...func(db *gorm.DB) *gorm.DB) ([]T, error)
	Count(ctx context.Context, db *gorm.DB, funcs ...func(db *gorm.DB) *gorm.DB) (int64, error)
	Delete(ctx context.Context, db *gorm.DB, funcs ...func(db *gorm.DB) *gorm.DB) error
}

var ErrMySQL = fmt.Errorf("MySQL error occurred")

type Repo[T Entity] struct{}

func NewRepo[T Entity]() *Repo[T] {
	return &Repo[T]{}
}

func logWithError(ctx context.Context, operation string, err error) {
	if err == nil {
		return
	}
	logger.Errorf(ctx, "operation: %s, err: %s", operation, err.Error())
}

func (r *Repo[T]) Insert(ctx context.Context, db *gorm.DB, newValue *T) error {
	if newValue == nil {
		return fmt.Errorf("invalid argument: newValue is nil")
	}
	result := db.WithContext(ctx).Create(newValue)
	if result.Error != nil {
		logWithError(ctx, "insert", result.Error)
		return errors.Join(ErrMySQL, fmt.Errorf("failed to insert record, err: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no records inserted")
	}
	return nil
}

func (r *Repo[T]) BatchInsert(ctx context.Context, db *gorm.DB, valuesToAdd []*T, batchSize int) error {
	if len(valuesToAdd) == 0 {
		return fmt.Errorf("invalid argument: valuesToAdd is empty")
	}
	for i, v := range valuesToAdd {
		if v == nil {
			return fmt.Errorf("invalid argument: value at index %d is nil", i)
		}
	}
	if batchSize <= 0 {
		batchSize = 10
	}
	result := db.WithContext(ctx).CreateInBatches(valuesToAdd, batchSize)
	if result.Error != nil {
		logWithError(ctx, "batch insert", result.Error)
		return errors.Join(ErrMySQL, fmt.Errorf("failed to batch insert records, err: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no records inserted")
	}
	return nil
}

func (r *Repo[T]) Update(ctx context.Context, db *gorm.DB, newValue *T, funcs ...func(db *gorm.DB) *gorm.DB) error {
	if newValue == nil {
		return fmt.Errorf("invalid argument: newValue is nil")
	}
	result := db.WithContext(ctx).Model(new(T)).Scopes(funcs...).Updates(newValue)
	if result.Error != nil {
		logWithError(ctx, "update", result.Error)
		return errors.Join(ErrMySQL, fmt.Errorf("failed to update record, err: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no records inserted")
	}
	return nil
}

func (r *Repo[T]) UpdateWithMap(ctx context.Context, db *gorm.DB, newValue map[string]interface{}, funcs ...func(db *gorm.DB) *gorm.DB) error {
	if newValue == nil {
		return fmt.Errorf("invalid argument: newValue is nil")
	}
	result := db.WithContext(ctx).Model(new(T)).Scopes(funcs...).Updates(newValue)
	if result.Error != nil {
		logWithError(ctx, "update with map", result.Error)
		return errors.Join(ErrMySQL, fmt.Errorf("failed to update with map, err: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no records inserted")
	}
	return nil
}

func (r *Repo[T]) Delete(ctx context.Context, db *gorm.DB, funcs ...func(db *gorm.DB) *gorm.DB) error {
	result := db.WithContext(ctx).Model(new(T)).Scopes(funcs...).Delete(new(T))
	if result.Error != nil {
		logWithError(ctx, "delete", result.Error)
		return errors.Join(ErrMySQL, fmt.Errorf("failed to delete record, err: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no records inserted")
	}
	return nil
}

func (r *Repo[T]) Query(ctx context.Context, db *gorm.DB, funcs ...func(db *gorm.DB) *gorm.DB) (*T, error) {
	var record T
	result := db.WithContext(ctx).Scopes(funcs...).Limit(1).Find(&record)
	if result.Error != nil {
		logWithError(ctx, "query one", result.Error)
		return nil, errors.Join(ErrMySQL, fmt.Errorf("failed to query one record, err: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &record, nil
}

func (r *Repo[T]) QueryList(ctx context.Context, db *gorm.DB, funcs ...func(db *gorm.DB) *gorm.DB) ([]T, error) {
	var recordList []T
	result := db.WithContext(ctx).Scopes(funcs...).Find(&recordList)
	if result.Error != nil {
		logWithError(ctx, "query list", result.Error)
		return nil, errors.Join(ErrMySQL, fmt.Errorf("failed to query list of records, err: %w", result.Error))
	}
	return recordList, nil
}

func (r *Repo[T]) Count(ctx context.Context, db *gorm.DB, funcs ...func(db *gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	result := db.WithContext(ctx).Model(new(T)).Scopes(funcs...).Count(&count)
	if result.Error != nil {
		logWithError(ctx, "count", result.Error)
		return 0, errors.Join(ErrMySQL, fmt.Errorf("failed to count records, err: %w", result.Error))
	}
	return count, nil
}

const (
	defaultPageSize = 10
	maxPageSize     = 100
)

// Paginate applies pagination to the database query.
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = defaultPageSize
		} else if pageSize > maxPageSize {
			pageSize = maxPageSize
		}
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

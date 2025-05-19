package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// Common repository errors
var (
	ErrNotFound         = errors.New("record not found")
	ErrDuplicateRecord  = errors.New("duplicate record")
	ErrValidationFailed = errors.New("validation failed")
	ErrDatabase         = errors.New("database error")
	ErrInvalidInput     = errors.New("invalid input")
)

type Repository interface {
	WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error
}

type BaseRepository struct {
	db *gorm.DB
}

func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// Executes the given function within a transaction
func (r *BaseRepository) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("%w: %v", ErrDatabase, tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	return nil
}

// Standardizes error handling from database operations
func HandleDBError(err error, entity string) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("%w: %s not found", ErrNotFound, entity)
	}

	// Check for unique constraint violations
	if errors.Is(err, gorm.ErrDuplicatedKey) ||
		(err.Error() != "" && (contains(err.Error(), "duplicate") || contains(err.Error(), "unique constraint"))) {
		return fmt.Errorf("%w: %s already exists", ErrDuplicateRecord, entity)
	}

	// Check for foreign key violations
	if err.Error() != "" && contains(err.Error(), "foreign key") {
		return fmt.Errorf("%w: invalid related entity for %s", ErrValidationFailed, entity)
	}

	// Wrap other database errors
	return fmt.Errorf("%w when processing %s: %v", ErrDatabase, entity, err)
}

func IsErrNotFound(err error) bool {
	return errors.Is(err, ErrNotFound) || errors.Is(err, gorm.ErrRecordNotFound)
}

func IsErrDuplicate(err error) bool {
	return errors.Is(err, ErrDuplicateRecord)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr
}

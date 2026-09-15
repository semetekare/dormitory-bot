package service

import (
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

func (tm *TransactionManager) BeginTransaction() (*gorm.DB, error) {
	tx := tm.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return tx, nil
}

func (tm *TransactionManager) CommitTransaction(tx *gorm.DB) error {
	return tx.Commit().Error
}

func (tm *TransactionManager) RollbackTransaction(tx *gorm.DB) error {
	return tx.Rollback().Error
}

func (tm *TransactionManager) ExecuteInTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	tx, err := tm.BeginTransaction()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tm.RollbackTransaction(tx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tm.RollbackTransaction(tx)
		return err
	}

	return tm.CommitTransaction(tx)
}

func (tm *TransactionManager) ExecuteInTransactionWithID(ctx context.Context, transactionID uuid.UUID, fn func(*gorm.DB) error) error {
	return tm.ExecuteInTransaction(ctx, fn)
}

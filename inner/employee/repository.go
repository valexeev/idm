package employee

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// TxWrapper wraps sqlx.Tx to implement Transaction interface
type TxWrapper struct {
	tx *sqlx.Tx
}

func (w *TxWrapper) Rollback() error {
	return w.tx.Rollback()
}

func (w *TxWrapper) Commit() error {
	return w.tx.Commit()
}

func (w *TxWrapper) Get(dest interface{}, query string, args ...interface{}) error {
	return w.tx.Get(dest, query, args...)
}

func (w *TxWrapper) QueryRow(query string, args ...interface{}) Row {
	return &RowWrapper{row: w.tx.QueryRow(query, args...)}
}

func (w *TxWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	return &RowWrapper{row: w.tx.QueryRowContext(ctx, query, args...)}
}

// RowWrapper wraps sql.Row to implement Row interface
type RowWrapper struct {
	row *sql.Row
}

func (w *RowWrapper) Scan(dest ...interface{}) error {
	return w.row.Scan(dest...)
}

// Repository представляет репозиторий для работы с сотрудниками
type Repository struct {
	db *sqlx.DB
}

// NewRepository создает новый экземпляр Repository
func NewRepository(database *sqlx.DB) *Repository {
	return &Repository{db: database}
}

func (r *Repository) FindById(ctx context.Context, id int64) (res Entity, err error) {
	err = r.db.GetContext(ctx, &res, "select * from employee where id = $1", id)
	return res, err
}

func (r *Repository) Add(ctx context.Context, e *Entity) error {
	query := `INSERT INTO employee (name, created_at, updated_at) VALUES ($1, $2, $3) RETURNING id`
	return r.db.QueryRowContext(ctx, query, e.Name, e.CreatedAt, e.UpdatedAt).Scan(&e.Id)
}

func (r *Repository) FindAll(ctx context.Context) ([]Entity, error) {
	var res []Entity
	err := r.db.SelectContext(ctx, &res, "SELECT * FROM employee")
	return res, err
}

func (r *Repository) FindByIds(ctx context.Context, ids []int64) ([]Entity, error) {
	query := `SELECT * FROM employee WHERE id = ANY($1)`
	var res []Entity
	err := r.db.SelectContext(ctx, &res, query, pq.Array(ids))
	return res, err
}

func (r *Repository) DeleteById(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM employee WHERE id = $1", id)
	return err
}

func (r *Repository) DeleteByIds(ctx context.Context, ids []int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM employee WHERE id = ANY($1)", pq.Array(ids))
	return err
}

// BeginTransaction начинает новую транзакцию
func (r *Repository) BeginTransaction(ctx context.Context) (Transaction, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &TxWrapper{tx: tx}, nil
}

// FindByNameTx проверяет наличие в базе данных сотрудника с заданным именем в рамках транзакции
func (r *Repository) FindByNameTx(_ context.Context, tx Transaction, name string) (bool, error) {
	var exists bool
	err := tx.Get(
		&exists,
		"SELECT EXISTS(SELECT 1 FROM employee WHERE name = $1)",
		name,
	)
	return exists, err
}

// AddTx добавляет нового сотрудника в рамках транзакции
func (r *Repository) AddTx(ctx context.Context, tx Transaction, e *Entity) error {
	query := `INSERT INTO employee (name, created_at, updated_at) VALUES ($1, $2, $3) RETURNING id`
	return tx.QueryRowContext(ctx, query, e.Name, e.CreatedAt, e.UpdatedAt).Scan(&e.Id)
}

// FindPage возвращает сотрудников с учетом пагинации (limit, offset, textFilter)
func (r *Repository) FindPage(ctx context.Context, limit, offset int, textFilter string) ([]Entity, error) {
	var res []Entity
	var (
		query string
		args  []interface{}
	)
	baseQuery := "SELECT * FROM employee WHERE 1=1"
	if validTextFilter(textFilter) {
		baseQuery += " AND name ilike $1"
		args = append(args, "%"+textFilter+"%")
		baseQuery += " OFFSET $2 LIMIT $3"
		args = append(args, offset, limit)
	} else {
		baseQuery += " OFFSET $1 LIMIT $2"
		args = append(args, offset, limit)
	}
	query = baseQuery
	err := r.db.SelectContext(ctx, &res, query, args...)
	return res, err
}

// CountAll возвращает общее количество сотрудников с учетом фильтра
func (r *Repository) CountAll(ctx context.Context, textFilter string) (int64, error) {
	var total int64
	var (
		query string
		args  []interface{}
	)
	baseQuery := "SELECT COUNT(*) FROM employee WHERE 1=1"
	if validTextFilter(textFilter) {
		baseQuery += " AND name ilike $1"
		args = append(args, "%"+textFilter+"%")
	}
	query = baseQuery
	err := r.db.GetContext(ctx, &total, query, args...)
	return total, err
}

// validTextFilter проверяет, что фильтр содержит минимум 3 непробельных символа
func validTextFilter(s string) bool {
	count := 0
	for _, r := range s {
		if r != ' ' && r != '\n' && r != '\t' {
			count++
		}
	}
	return count >= 3
}

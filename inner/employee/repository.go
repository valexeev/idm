package employee

import (
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

func (r *Repository) FindById(id int64) (res Entity, err error) {
	err = r.db.Get(&res, "select * from employee where id = $1", id)
	return res, err
}

func (r *Repository) Add(e *Entity) error {
	query := `INSERT INTO employee (name, created_at, updated_at) VALUES ($1, $2, $3) RETURNING id`
	return r.db.QueryRow(query, e.Name, e.CreatedAt, e.UpdatedAt).Scan(&e.Id)
}

func (r *Repository) FindAll() ([]Entity, error) {
	var res []Entity
	err := r.db.Select(&res, "SELECT * FROM employee")
	return res, err
}

func (r *Repository) FindByIds(ids []int64) ([]Entity, error) {
	query := `SELECT * FROM employee WHERE id = ANY($1)`
	var res []Entity
	err := r.db.Select(&res, query, pq.Array(ids))
	return res, err
}

func (r *Repository) DeleteById(id int64) error {
	_, err := r.db.Exec("DELETE FROM employee WHERE id = $1", id)
	return err
}

func (r *Repository) DeleteByIds(ids []int64) error {
	_, err := r.db.Exec("DELETE FROM employee WHERE id = ANY($1)", pq.Array(ids))
	return err
}

// BeginTransaction начинает новую транзакцию
func (r *Repository) BeginTransaction() (Transaction, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	return &TxWrapper{tx: tx}, nil
}

// FindByNameTx проверяет наличие в базе данных сотрудника с заданным именем в рамках транзакции
func (r *Repository) FindByNameTx(tx Transaction, name string) (bool, error) {
	var exists bool
	err := tx.Get(
		&exists,
		"SELECT EXISTS(SELECT 1 FROM employee WHERE name = $1)",
		name,
	)
	return exists, err
}

// AddTx добавляет нового сотрудника в рамках транзакции
func (r *Repository) AddTx(tx Transaction, e *Entity) error {
	query := `INSERT INTO employee (name, created_at, updated_at) VALUES ($1, $2, $3) RETURNING id`
	return tx.QueryRow(query, e.Name, e.CreatedAt, e.UpdatedAt).Scan(&e.Id)
}

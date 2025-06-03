package employee

import (
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

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

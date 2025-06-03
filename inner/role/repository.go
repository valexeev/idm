package role

import (
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// Repository представляет репозиторий для работы с ролями
type Repository struct {
	db *sqlx.DB
}

// NewRepository создает новый экземпляр Repository
func NewRepository(database *sqlx.DB) *Repository {
	return &Repository{db: database}
}

func (r *Repository) Add(e *Entity) error {
	query := `insert into role (name, created_at, updated_at) values ($1, $2, $3) returning id`
	return r.db.QueryRow(query, e.Name, e.CreatedAt, e.UpdatedAt).Scan(&e.Id)
}

func (r *Repository) FindById(id int64) (res Entity, err error) {
	err = r.db.Get(&res, "select * from role where id = $1", id)
	return res, err
}

func (r *Repository) FindAll() (res []Entity, err error) {
	err = r.db.Select(&res, "select * from role")
	return res, err
}

func (r *Repository) FindByIds(ids []int64) (res []Entity, err error) {
	query := `select * from role where id = any($1)`
	err = r.db.Select(&res, query, pq.Array(ids))
	return res, err
}

func (r *Repository) DeleteById(id int64) error {
	_, err := r.db.Exec("delete from role where id = $1", id)
	return err
}

func (r *Repository) DeleteByIds(ids []int64) error {
	_, err := r.db.Exec("delete from role where id = any($1)", pq.Array(ids))
	return err
}

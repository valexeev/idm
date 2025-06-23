package role

import (
	"context"

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

func (r *Repository) Add(ctx context.Context, e *Entity) error {
	query := `insert into role (name, created_at, updated_at) values ($1, $2, $3) returning id`
	return r.db.QueryRowContext(ctx, query, e.Name, e.CreatedAt, e.UpdatedAt).Scan(&e.Id)
}

func (r *Repository) FindById(ctx context.Context, id int64) (res Entity, err error) {
	err = r.db.GetContext(ctx, &res, "select * from role where id = $1", id)
	return res, err
}

func (r *Repository) FindAll(ctx context.Context) (res []Entity, err error) {
	err = r.db.SelectContext(ctx, &res, "select * from role")
	return res, err
}

func (r *Repository) FindByIds(ctx context.Context, ids []int64) (res []Entity, err error) {
	query := `select * from role where id = any($1)`
	err = r.db.SelectContext(ctx, &res, query, pq.Array(ids))
	return res, err
}

func (r *Repository) DeleteById(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "delete from role where id = $1", id)
	return err
}

func (r *Repository) DeleteByIds(ctx context.Context, ids []int64) error {
	_, err := r.db.ExecContext(ctx, "delete from role where id = any($1)", pq.Array(ids))
	return err
}

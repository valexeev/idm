package role

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type RoleRepository struct {
	db *sqlx.DB
}

func NewRoleRepository(database *sqlx.DB) *RoleRepository {
	return &RoleRepository{db: database}
}

type RoleEntity struct {
	Id        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (r *RoleRepository) Add(e *RoleEntity) error {
	query := `insert into role (name, created_at, updated_at) values ($1, $2, $3) returning id`
	return r.db.QueryRow(query, e.Name, e.CreatedAt, e.UpdatedAt).Scan(&e.Id)
}

func (r *RoleRepository) FindById(id int64) (res RoleEntity, err error) {
	err = r.db.Get(&res, "select * from role where id = $1", id)
	return res, err
}

func (r *RoleRepository) FindAll() (res []RoleEntity, err error) {
	err = r.db.Select(&res, "select * from role")
	return res, err
}

func (r *RoleRepository) FindByIds(ids []int64) (res []RoleEntity, err error) {
	query := `select * from role where id = any($1)`
	err = r.db.Select(&res, query, ids)
	return res, err
}

func (r *RoleRepository) DeleteById(id int64) error {
	_, err := r.db.Exec("delete from role where id = $1", id)
	return err
}

func (r *RoleRepository) DeleteByIds(ids []int64) error {
	_, err := r.db.Exec("delete from role where id = any($1)", ids)
	return err
}

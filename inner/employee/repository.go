package employee

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type EmployeeRepository struct {
	db *sqlx.DB
}

func NewEmployeeRepository(database *sqlx.DB) *EmployeeRepository {
	return &EmployeeRepository{db: database}
}

type EmployeeEntity struct {
	Id        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (r *EmployeeRepository) FindById(id int64) (res EmployeeEntity, err error) {
	err = r.db.Get(&res, "select * from employee where id = $1", id)
	return res, err
}

func (r *EmployeeRepository) Add(e *EmployeeEntity) error {
	query := `INSERT INTO employee (name, created_at, updated_at) VALUES ($1, $2, $3) RETURNING id`
	return r.db.QueryRow(query, e.Name, e.CreatedAt, e.UpdatedAt).Scan(&e.Id)
}

func (r *EmployeeRepository) FindAll() ([]EmployeeEntity, error) {
	var res []EmployeeEntity
	err := r.db.Select(&res, "SELECT * FROM employee")
	return res, err
}

func (r *EmployeeRepository) FindByIds(ids []int64) ([]EmployeeEntity, error) {
	query := `SELECT * FROM employee WHERE id = ANY($1)`
	var res []EmployeeEntity
	err := r.db.Select(&res, query, ids)
	return res, err
}

func (r *EmployeeRepository) DeleteById(id int64) error {
	_, err := r.db.Exec("DELETE FROM employee WHERE id = $1", id)
	return err
}

func (r *EmployeeRepository) DeleteByIds(ids []int64) error {
	_, err := r.db.Exec("DELETE FROM employee WHERE id = ANY($1)", ids)
	return err
}

package idm_test

import (
	"context"
	"fmt"
	"idm/inner/employee"
	"idm/inner/role"
	"time"

	"github.com/jmoiron/sqlx"
)

// Fixture — вспомогательная структура для подготовки данных в тестах
type Fixture struct {
	employees *employee.Repository
	roles     *role.Repository
	db        *sqlx.DB
}

// NewFixture — создаёт фикстуру и инициализирует таблицы, если нужно
func NewFixture(employees *employee.Repository, roles *role.Repository, db *sqlx.DB) (*Fixture, error) {
	f := &Fixture{employees: employees, roles: roles, db: db}
	if err := f.setupDatabase(); err != nil {
		return nil, fmt.Errorf("setup database: %w", err)
	}
	return f, nil
}

// setupDatabase — создаёт таблицы, если они ещё не созданы
func (f *Fixture) setupDatabase() error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS role (
			id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT now(),
			updated_at TIMESTAMPTZ DEFAULT now()
		)`,
		`CREATE TABLE IF NOT EXISTS employee (
			id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT now(),
			updated_at TIMESTAMPTZ DEFAULT now()
		)`,
	}
	for _, q := range tables {
		if _, err := f.db.Exec(q); err != nil {
			return fmt.Errorf("create table: %w", err)
		}
	}
	return nil
}

// Очистка таблиц после каждого теста
func (f *Fixture) CleanupDatabase() error {
	return f.ClearAllTables()
}

func (f *Fixture) ClearAllTables() error {
	for _, table := range []string{"employee", "role"} {
		if err := f.ClearTable(table); err != nil {
			return err
		}
	}
	return nil
}

func (f *Fixture) ClearTable(table string) error {
	allowed := map[string]bool{"employee": true, "role": true}
	if !allowed[table] {
		return fmt.Errorf("table %s not allowed", table)
	}
	query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)
	_, err := f.db.Exec(query)
	return err
}

// Employee — создаёт сотрудника, возвращает ID
func (f *Fixture) Employee(name string) (int64, error) {
	now := time.Now()
	e := employee.Entity{Name: name, CreatedAt: now, UpdatedAt: now}
	ctx := context.Background()
	if err := f.employees.Add(ctx, &e); err != nil {
		return 0, err
	}
	return e.Id, nil
}

// Role — создаёт роль, возвращает ID
func (f *Fixture) Role(name string) (int64, error) {
	now := time.Now()
	r := role.Entity{Name: name, CreatedAt: now, UpdatedAt: now}
	if err := f.roles.Add(context.Background(), &r); err != nil {
		return 0, err
	}
	return r.Id, nil
}

// MustEmployee — то же, что Employee, но паникует при ошибке (для тестов)
func (f *Fixture) MustEmployee(name string) int64 {
	id, err := f.Employee(name)
	if err != nil {
		panic("MustEmployee: " + err.Error())
	}
	return id
}

func (f *Fixture) MustRole(name string) int64 {
	id, err := f.Role(name)
	if err != nil {
		panic("MustRole: " + err.Error())
	}
	return id
}

func (f *Fixture) CreateMultipleEmployees(names []string) ([]int64, error) {
	var ids []int64
	for _, name := range names {
		id, err := f.Employee(name)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (f *Fixture) CreateMultipleRoles(names []string) ([]int64, error) {
	var ids []int64
	for _, name := range names {
		id, err := f.Role(name)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

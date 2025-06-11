package employee

import (
	"errors"
	"idm/inner/common/validator"
	"testing"
	"time"

	"github.com/78bits/go-sqlmock-sqlx"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestEmployeeService_AddTransactional_WithSqlMock(t *testing.T) {
	t.Run("should return error for empty name", func(t *testing.T) {
		a := assert.New(t)
		db, mock, err := sqlmock.New()
		a.NoError(err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		repo := NewRepository(sqlxDB)
		validator := validator.New()
		svc := NewService(repo, validator)

		// No mock expectations needed since we return early for empty name
		request := AddEmployeeRequest{Name: ""}
		response, err := svc.AddTransactional(request)

		// Debug logging
		t.Logf("Response: %+v", response)
		t.Logf("Error: %v", err)

		// Verify empty response
		expectedResponse := Response{}
		a.Equal(expectedResponse, response, "Response should be empty for empty name")

		// Verify error exists and contains expected message
		a.Error(err, "Should return error for empty name")
		if err != nil {
			a.Contains(err.Error(), "cannot be empty", "Error message should contain 'cannot be empty'")
		}

		// Since we return early, no database calls should be made
		a.NoError(mock.ExpectationsWereMet())
	})

	t.Run("should create employee successfully", func(t *testing.T) {
		a := assert.New(t)
		db, mock, err := sqlmock.New()
		a.NoError(err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		repo := NewRepository(sqlxDB)
		validator := validator.New()
		svc := NewService(repo, validator)

		// Настраиваем mock
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM employee WHERE name = \$1\)`).
			WithArgs("John Doe").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
		mock.ExpectQuery(`INSERT INTO employee \(name, created_at, updated_at\) VALUES \(\$1, \$2, \$3\) RETURNING id`).
			WithArgs("John Doe", sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		request := AddEmployeeRequest{Name: "John Doe"}
		response, err := svc.AddTransactional(request)

		a.NoError(err)
		a.Equal(int64(1), response.Id)
		a.Equal("John Doe", response.Name)

		a.NoError(mock.ExpectationsWereMet())
	})

	t.Run("should handle existing employee", func(t *testing.T) {
		a := assert.New(t)
		db, mock, err := sqlmock.New()
		a.NoError(err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		repo := NewRepository(sqlxDB)
		validator := validator.New()
		svc := NewService(repo, validator)

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM employee WHERE name = \$1\)`).
			WithArgs("Existing Employee").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
		mock.ExpectRollback()

		request := AddEmployeeRequest{Name: "Existing Employee"}
		response, err := svc.AddTransactional(request)

		a.Equal(Response{}, response)
		a.Error(err)
		a.Contains(err.Error(), "already exists")

		a.NoError(mock.ExpectationsWereMet())
	})

	t.Run("should rollback on insert error", func(t *testing.T) {
		a := assert.New(t)
		db, mock, err := sqlmock.New()
		a.NoError(err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		repo := NewRepository(sqlxDB)
		validator := validator.New()
		svc := NewService(repo, validator)

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM employee WHERE name = \$1\)`).
			WithArgs("John Doe").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
		mock.ExpectQuery(`INSERT INTO employee \(name, created_at, updated_at\) VALUES \(\$1, \$2, \$3\) RETURNING id`).
			WithArgs("John Doe", sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("insert failed"))
		mock.ExpectRollback()

		request := AddEmployeeRequest{Name: "John Doe"}
		response, err := svc.AddTransactional(request)

		a.Equal(Response{}, response)
		a.Error(err)
		a.Contains(err.Error(), "error adding employee")

		a.NoError(mock.ExpectationsWereMet())
	})

	t.Run("should handle commit error", func(t *testing.T) {
		a := assert.New(t)
		db, mock, err := sqlmock.New()
		a.NoError(err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		repo := NewRepository(sqlxDB)
		validator := validator.New()
		svc := NewService(repo, validator)

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM employee WHERE name = \$1\)`).
			WithArgs("John Doe").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
		mock.ExpectQuery(`INSERT INTO employee \(name, created_at, updated_at\) VALUES \(\$1, \$2, \$3\) RETURNING id`).
			WithArgs("John Doe", sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

		request := AddEmployeeRequest{Name: "John Doe"}
		response, err := svc.AddTransactional(request)

		a.Equal(Response{}, response)
		a.Error(err)
		a.Contains(err.Error(), "commiting transaction error")

		a.NoError(mock.ExpectationsWereMet())
	})

	t.Run("should handle transaction begin error", func(t *testing.T) {
		a := assert.New(t)
		db, mock, err := sqlmock.New()
		a.NoError(err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		repo := NewRepository(sqlxDB)
		validator := validator.New()
		svc := NewService(repo, validator)

		// Настраиваем ошибку при начале транзакции
		mock.ExpectBegin().WillReturnError(errors.New("failed to begin transaction"))

		request := AddEmployeeRequest{Name: "John Doe"}
		response, err := svc.AddTransactional(request)

		a.Equal(Response{}, response)
		a.Error(err)
		a.Contains(err.Error(), "error creating transaction")

		a.NoError(mock.ExpectationsWereMet())
	})

	t.Run("should rollback on check existence error", func(t *testing.T) {
		a := assert.New(t)
		db, mock, err := sqlmock.New()
		a.NoError(err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		repo := NewRepository(sqlxDB)
		validator := validator.New()
		svc := NewService(repo, validator)

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM employee WHERE name = \$1\)`).
			WithArgs("John Doe").
			WillReturnError(errors.New("database connection error"))
		mock.ExpectRollback()

		request := AddEmployeeRequest{Name: "John Doe"}
		response, err := svc.AddTransactional(request)

		a.Equal(Response{}, response)
		a.Error(err)
		a.Contains(err.Error(), "error checking employee existence")

		a.NoError(mock.ExpectationsWereMet())
	})

	t.Run("should handle rollback error on insert failure", func(t *testing.T) {
		a := assert.New(t)
		db, mock, err := sqlmock.New()
		a.NoError(err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		repo := NewRepository(sqlxDB)
		validator := validator.New()
		svc := NewService(repo, validator)

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM employee WHERE name = \$1\)`).
			WithArgs("John Doe").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
		mock.ExpectQuery(`INSERT INTO employee \(name, created_at, updated_at\) VALUES \(\$1, \$2, \$3\) RETURNING id`).
			WithArgs("John Doe", sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("insert failed"))
		mock.ExpectRollback().WillReturnError(errors.New("rollback failed"))

		request := AddEmployeeRequest{Name: "John Doe"}
		response, err := svc.AddTransactional(request)

		a.Equal(Response{}, response)
		a.Error(err)
		a.Contains(err.Error(), "rolling back transaction errors")
		a.Contains(err.Error(), "insert failed")
		a.Contains(err.Error(), "rollback failed")

		a.NoError(mock.ExpectationsWereMet())
	})
}

func TestEmployeeService_NonTransactionalMethods_WithSqlMock(t *testing.T) {
	t.Run("FindById should work with mock", func(t *testing.T) {
		a := assert.New(t)
		db, mock, err := sqlmock.New()
		a.NoError(err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		repo := NewRepository(sqlxDB)
		validator := validator.New()
		svc := NewService(repo, validator)

		// Use actual time values instead of sqlmock.AnyArg()
		createdAt := time.Date(2025, 6, 7, 10, 0, 0, 0, time.UTC)
		updatedAt := time.Date(2025, 6, 7, 10, 0, 0, 0, time.UTC)

		mock.ExpectQuery(`select \* from employee where id = \$1`).
			WithArgs(int64(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}).
				AddRow(1, "John Doe", createdAt, updatedAt))

		response, err := svc.FindById(1)

		a.NoError(err)
		a.Equal(int64(1), response.Id)
		a.Equal("John Doe", response.Name)

		a.NoError(mock.ExpectationsWereMet())
	})

	t.Run("Add should work with mock", func(t *testing.T) {
		a := assert.New(t)
		db, mock, err := sqlmock.New()
		a.NoError(err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		repo := NewRepository(sqlxDB)
		validator := validator.New()
		svc := NewService(repo, validator)

		mock.ExpectQuery(`INSERT INTO employee \(name, created_at, updated_at\) VALUES \(\$1, \$2, \$3\) RETURNING id`).
			WithArgs("John Doe", sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		response, err := svc.Add("John Doe")

		a.NoError(err)
		a.Equal(int64(1), response.Id)
		a.Equal("John Doe", response.Name)

		a.NoError(mock.ExpectationsWereMet())
	})
}

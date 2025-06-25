package employee

import (
	"context"
	"testing"
	"time"

	"github.com/78bits/go-sqlmock-sqlx"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestRepository_TransactionalMethods(t *testing.T) {
	a := assert.New(t)

	t.Run("should execute queries within same transaction", func(t *testing.T) {
		// Создаем mock базы данных
		db, mock, err := sqlmock.New()
		a.NoError(err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		repo := NewRepository(sqlxDB)
		ctx := context.Background()

		// Настраиваем mock для начала транзакции
		mock.ExpectBegin()

		// Настраиваем mock для проверки существования сотрудника
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM employee WHERE name = \$1\)`).
			WithArgs("John Doe").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		// Настраиваем mock для создания сотрудника
		mock.ExpectQuery(`INSERT INTO employee \(name, created_at, updated_at\) VALUES \(\$1, \$2, \$3\) RETURNING id`).
			WithArgs("John Doe", sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// Настраиваем mock для коммита транзакции
		mock.ExpectCommit()

		// Начинаем транзакцию
		tx, err := repo.BeginTransaction(ctx)
		a.NoError(err)
		a.NotNil(tx)

		// Проверяем существование сотрудника
		exists, err := repo.FindByNameTx(ctx, tx, "John Doe")
		a.NoError(err)
		a.False(exists)

		// Создаем сотрудника
		now := time.Now()
		entity := &Entity{
			Name:      "John Doe",
			CreatedAt: now,
			UpdatedAt: now,
		}

		err = repo.AddTx(ctx, tx, entity)
		a.NoError(err)
		a.Equal(int64(1), entity.Id)

		// Коммитим транзакцию
		err = tx.Commit()
		a.NoError(err)

		// Проверяем, что все ожидания выполнены
		a.NoError(mock.ExpectationsWereMet())
	})

	t.Run("should find existing employee by name in transaction", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		a.NoError(err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		repo := NewRepository(sqlxDB)
		ctx := context.Background()

		// Настраиваем mock для начала транзакции
		mock.ExpectBegin()

		// Настраиваем mock для проверки существования сотрудника
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM employee WHERE name = \$1\)`).
			WithArgs("Existing Employee").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		// Настраиваем mock для отката транзакции
		mock.ExpectRollback()

		// Начинаем транзакцию
		tx, err := repo.BeginTransaction(ctx)
		a.NoError(err)

		// Проверяем существование сотрудника
		exists, err := repo.FindByNameTx(ctx, tx, "Existing Employee")
		a.NoError(err)
		a.True(exists)

		// Откатываем транзакцию
		err = tx.Rollback()
		a.NoError(err)

		// Проверяем, что все ожидания выполнены
		a.NoError(mock.ExpectationsWereMet())
	})
}

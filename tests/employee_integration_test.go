package idm_test

import (
	"idm/inner/common"
	"idm/inner/database"
	"idm/inner/employee"
	"idm/inner/role"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEmployee_TransactionalMethods_Integration(t *testing.T) {
	a := assert.New(t)
	cfg := common.GetConfig(".env.tests")

	db := database.ConnectDbWithCfg(cfg)
	defer db.Close()

	employeeRepo := employee.NewRepository(db)
	roleRepo := role.NewRepository(db)

	// Создаем и настраиваем fixture
	fixture, err := NewFixture(employeeRepo, roleRepo, db)
	if err != nil {
		t.Fatal("Не удалось создать fixture:", err)
	}

	// Очищаем базу данных перед всеми тестами
	err = fixture.CleanupDatabase()
	if err != nil {
		t.Fatal("Не удалось очистить базу данных:", err)
	}

	t.Run("should execute transactional operations with real database", func(t *testing.T) {
		defer func() {
			if err := fixture.CleanupDatabase(); err != nil {
				t.Errorf("Failed to cleanup database: %v", err)
			}
		}()

		// Начинаем транзакцию
		tx, err := employeeRepo.BeginTransaction()
		a.NoError(err)
		a.NotNil(tx)

		// Проверяем, что сотрудника с таким именем не существует
		exists, err := employeeRepo.FindByNameTx(tx, "John Transaction")
		a.NoError(err)
		a.False(exists)

		// Создаем сотрудника в рамках транзакции
		now := time.Now()
		entity := &employee.Entity{
			Name:      "John Transaction",
			CreatedAt: now,
			UpdatedAt: now,
		}

		err = employeeRepo.AddTx(tx, entity)
		a.NoError(err)
		a.Greater(entity.Id, int64(0))

		// Коммитим транзакцию
		err = tx.Commit()
		a.NoError(err)

		// Проверяем, что сотрудник действительно создан
		savedEmployee, err := employeeRepo.FindById(entity.Id)
		a.NoError(err)
		a.Equal("John Transaction", savedEmployee.Name)
	})

	t.Run("should rollback transaction on error", func(t *testing.T) {
		defer func() {
			if err := fixture.CleanupDatabase(); err != nil {
				t.Errorf("Failed to cleanup database: %v", err)
			}
		}()

		// Начинаем транзакцию
		tx, err := employeeRepo.BeginTransaction()
		a.NoError(err)

		// Создаем сотрудника
		now := time.Now()
		entity := &employee.Entity{
			Name:      "Rollback Test",
			CreatedAt: now,
			UpdatedAt: now,
		}

		err = employeeRepo.AddTx(tx, entity)
		a.NoError(err)

		// Откатываем транзакцию
		err = tx.Rollback()
		a.NoError(err)

		// Проверяем, что сотрудник не был сохранен
		_, err = employeeRepo.FindById(entity.Id)
		a.Error(err) // Должна быть ошибка, так как записи нет
	})

	t.Run("should detect existing employee in transaction", func(t *testing.T) {
		defer func() {
			if err := fixture.CleanupDatabase(); err != nil {
				t.Errorf("Failed to cleanup database: %v", err)
			}
		}()

		// Сначала создаем сотрудника обычным способом
		now := time.Now()
		existingEmployee := &employee.Entity{
			Name:      "Existing Employee",
			CreatedAt: now,
			UpdatedAt: now,
		}
		err = employeeRepo.Add(existingEmployee)
		a.NoError(err)

		// Теперь проверяем в транзакции
		tx, err := employeeRepo.BeginTransaction()
		a.NoError(err)
		defer func() {
			// Обрабатываем ошибку rollback корректно
			if err := tx.Rollback(); err != nil && err.Error() != "sql: transaction has already been committed or rolled back" {
				t.Logf("tx.Rollback failed: %v", err)
			}
		}()

		exists, err := employeeRepo.FindByNameTx(tx, "Existing Employee")
		a.NoError(err)
		a.True(exists)
	})

	t.Run("should create employee using AddTx method", func(t *testing.T) {
		defer func() {
			if err := fixture.CleanupDatabase(); err != nil {
				t.Errorf("Failed to cleanup database: %v", err)
			}
		}()

		tx, err := employeeRepo.BeginTransaction()
		a.NoError(err)

		now := time.Now()
		entity := &employee.Entity{
			Name:      "AddTx Test",
			CreatedAt: now,
			UpdatedAt: now,
		}

		err = employeeRepo.AddTx(tx, entity)
		a.NoError(err)
		a.Greater(entity.Id, int64(0))

		err = tx.Commit()
		a.NoError(err)

		// Проверяем, что сотрудник создан
		savedEmployee, err := employeeRepo.FindById(entity.Id)
		a.NoError(err)
		a.Equal("AddTx Test", savedEmployee.Name)
	})

	t.Run("should handle multiple operations in single transaction", func(t *testing.T) {
		defer func() {
			if err := fixture.CleanupDatabase(); err != nil {
				t.Errorf("Failed to cleanup database: %v", err)
			}
		}()

		tx, err := employeeRepo.BeginTransaction()
		a.NoError(err)

		now := time.Now()

		// Создаем первого сотрудника
		employee1 := &employee.Entity{Name: "Employee 1", CreatedAt: now, UpdatedAt: now}
		err = employeeRepo.AddTx(tx, employee1)
		a.NoError(err)
		a.Greater(employee1.Id, int64(0))

		// Создаем второго сотрудника
		employee2 := &employee.Entity{Name: "Employee 2", CreatedAt: now, UpdatedAt: now}
		err = employeeRepo.AddTx(tx, employee2)
		a.NoError(err)
		a.Greater(employee2.Id, int64(0))

		// Проверяем существование в рамках той же транзакции
		exists1, err := employeeRepo.FindByNameTx(tx, "Employee 1")
		a.NoError(err)
		a.True(exists1)

		exists2, err := employeeRepo.FindByNameTx(tx, "Employee 2")
		a.NoError(err)
		a.True(exists2)

		err = tx.Commit()
		a.NoError(err)

		// Проверяем, что оба сотрудника созданы
		saved1, err := employeeRepo.FindById(employee1.Id)
		a.NoError(err)
		a.Equal("Employee 1", saved1.Name)

		saved2, err := employeeRepo.FindById(employee2.Id)
		a.NoError(err)
		a.Equal("Employee 2", saved2.Name)
	})
}

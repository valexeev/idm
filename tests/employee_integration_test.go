package idm_test

import (
	"context"
	"encoding/json"
	"idm/inner/common"
	"idm/inner/database"
	"idm/inner/employee"
	"idm/inner/role"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEmployee_TransactionalMethods_Integration(t *testing.T) {
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

		ctx := context.Background()
		// Начинаем транзакцию
		tx, err := employeeRepo.BeginTransaction(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, tx)

		// Проверяем, что сотрудника с таким именем не существует
		exists, err := employeeRepo.FindByNameTx(ctx, tx, "John Transaction")
		assert.NoError(t, err)
		assert.False(t, exists)

		// Создаем сотрудника в рамках транзакции
		now := time.Now()
		entity := &employee.Entity{
			Name:      "John Transaction",
			CreatedAt: now,
			UpdatedAt: now,
		}

		err = employeeRepo.AddTx(ctx, tx, entity)
		assert.NoError(t, err)
		assert.Greater(t, entity.Id, int64(0))

		// Коммитим транзакцию
		err = tx.Commit()
		assert.NoError(t, err)

		// Проверяем, что сотрудник действительно создан
		savedEmployee, err := employeeRepo.FindById(ctx, entity.Id)
		assert.NoError(t, err)
		assert.Equal(t, "John Transaction", savedEmployee.Name)
	})

	t.Run("should rollback transaction on error", func(t *testing.T) {
		defer func() {
			if err := fixture.CleanupDatabase(); err != nil {
				t.Errorf("Failed to cleanup database: %v", err)
			}
		}()

		ctx := context.Background()
		// Начинаем транзакцию
		tx, err := employeeRepo.BeginTransaction(ctx)
		assert.NoError(t, err)

		// Создаем сотрудника
		now := time.Now()
		entity := &employee.Entity{
			Name:      "Rollback Test",
			CreatedAt: now,
			UpdatedAt: now,
		}

		err = employeeRepo.AddTx(ctx, tx, entity)
		assert.NoError(t, err)

		// Откатываем транзакцию
		err = tx.Rollback()
		assert.NoError(t, err)

		// Проверяем, что сотрудник не был сохранен
		_, err = employeeRepo.FindById(ctx, entity.Id)
		assert.Error(t, err) // Должна быть ошибка, так как записи нет
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
		ctx := context.Background()
		err = employeeRepo.Add(ctx, existingEmployee)
		assert.NoError(t, err)

		// Теперь проверяем в транзакции
		tx, err := employeeRepo.BeginTransaction(ctx)
		assert.NoError(t, err)

		defer func() {
			// Обрабатываем ошибку rollback корректно
			if err := tx.Rollback(); err != nil && err.Error() != "sql: transaction has already been committed or rolled back" {
				t.Logf("tx.Rollback failed: %v", err)
			}
		}()

		exists, err := employeeRepo.FindByNameTx(ctx, tx, "Existing Employee")
		assert.NoError(t, err)
		assert.True(t, exists)

	})

	t.Run("should create employee using AddTx method", func(t *testing.T) {
		defer func() {
			if err := fixture.CleanupDatabase(); err != nil {
				t.Errorf("Failed to cleanup database: %v", err)
			}
		}()

		ctx := context.Background()
		tx, err := employeeRepo.BeginTransaction(ctx)
		assert.NoError(t, err)

		now := time.Now()
		entity := &employee.Entity{
			Name:      "AddTx Test",
			CreatedAt: now,
			UpdatedAt: now,
		}

		err = employeeRepo.AddTx(ctx, tx, entity)
		assert.NoError(t, err)
		assert.Greater(t, entity.Id, int64(0))

		err = tx.Commit()
		assert.NoError(t, err)

		// Проверяем, что сотрудник создан
		savedEmployee, err := employeeRepo.FindById(ctx, entity.Id)
		assert.NoError(t, err)
		assert.Equal(t, "AddTx Test", savedEmployee.Name)
	})

	t.Run("should handle multiple operations in single transaction", func(t *testing.T) {
		defer func() {
			if err := fixture.CleanupDatabase(); err != nil {
				t.Errorf("Failed to cleanup database: %v", err)
			}
		}()

		ctx := context.Background()
		tx, err := employeeRepo.BeginTransaction(ctx)
		assert.NoError(t, err)
		now := time.Now()

		// Создаем первого сотрудника
		employee1 := &employee.Entity{Name: "Employee 1", CreatedAt: now, UpdatedAt: now}
		err = employeeRepo.AddTx(ctx, tx, employee1)
		assert.NoError(t, err)
		assert.Greater(t, employee1.Id, int64(0))

		// Создаем второго сотрудника
		employee2 := &employee.Entity{Name: "Employee 2", CreatedAt: now, UpdatedAt: now}
		err = employeeRepo.AddTx(ctx, tx, employee2)
		assert.NoError(t, err)
		assert.Greater(t, employee2.Id, int64(0))

		// Проверяем существование в рамках той же транзакции
		exists1, err := employeeRepo.FindByNameTx(ctx, tx, "Employee 1")
		assert.NoError(t, err)
		assert.True(t, exists1)

		exists2, err := employeeRepo.FindByNameTx(ctx, tx, "Employee 2")
		assert.NoError(t, err)
		assert.True(t, exists2)

		err = tx.Commit()
		assert.NoError(t, err)

		// Проверяем, что оба сотрудника созданы
		saved1, err := employeeRepo.FindById(ctx, employee1.Id)
		assert.NoError(t, err)
		assert.Equal(t, "Employee 1", saved1.Name)
		saved2, err := employeeRepo.FindById(ctx, employee2.Id)
		assert.NoError(t, err)
		assert.Equal(t, "Employee 2", saved2.Name)
	})
}

// Тест пагинации
func TestEmployee_Pagination_Integration(t *testing.T) {
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
	defer func() {
		if err := fixture.CleanupDatabase(); err != nil {
			t.Errorf("Failed to cleanup database: %v", err)
		}
	}()

	// Очищаем и наполняем БД 5 сотрудниками
	err = fixture.CleanupDatabase()
	if err != nil {
		t.Fatal("Не удалось очистить базу данных:", err)
	}
	names := []string{"Alice", "Bob", "Charlie", "David", "Eve"}
	_, err = fixture.CreateMultipleEmployees(names)
	if err != nil {
		t.Fatal("Не удалось создать сотрудников:", err)
	}

	// Дополнительная проверка количества записей
	var count int64
	err = db.Get(&count, "SELECT COUNT(*) FROM employee")
	if err != nil {
		t.Fatalf("Ошибка при подсчете сотрудников: %v", err)
	}

	app := setupTestApp(db)

	// Все проверки только внутри t.Run!
	type WrappedPageResponse struct {
		Success bool                  `json:"success"`
		Data    employee.PageResponse `json:"data"`
	}

	t.Run("should return 3 employees on first page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/page?pageNumber=0&pageSize=3", nil)
		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var wrappedResp WrappedPageResponse
		err = json.NewDecoder(resp.Body).Decode(&wrappedResp)
		assert.NoError(t, err)
		assert.Len(t, wrappedResp.Data.Result, 3)
		assert.Equal(t, int64(5), wrappedResp.Data.Total)
	})

	t.Run("should return 2 employees on second page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/page?pageNumber=1&pageSize=3", nil)
		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var wrappedResp WrappedPageResponse
		err = json.NewDecoder(resp.Body).Decode(&wrappedResp)
		assert.NoError(t, err)
		assert.Len(t, wrappedResp.Data.Result, 2)
		assert.Equal(t, int64(5), wrappedResp.Data.Total)
	})

	t.Run("should return 0 employees on third page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/page?pageNumber=2&pageSize=3", nil)
		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var wrappedResp WrappedPageResponse
		err = json.NewDecoder(resp.Body).Decode(&wrappedResp)
		assert.NoError(t, err)
		assert.Len(t, wrappedResp.Data.Result, 0)
		assert.Equal(t, int64(5), wrappedResp.Data.Total)
	})

	t.Run("should return error for invalid pageSize", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/page?pageNumber=0&pageSize=0", nil)
		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)

		var errResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&errResp)
		assert.NoError(t, err)
		assert.Contains(t, errResp["error"], "pagesize must be at least 1")
	})

	t.Run("should use default PageNumber if not provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/page?pageSize=3", nil)
		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var wrappedResp WrappedPageResponse
		err = json.NewDecoder(resp.Body).Decode(&wrappedResp)
		assert.NoError(t, err)
		assert.Equal(t, 0, wrappedResp.Data.PageNumber)
	})

	t.Run("should use default PageSize if not provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/page?pageNumber=0", nil)
		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var wrappedResp WrappedPageResponse
		err = json.NewDecoder(resp.Body).Decode(&wrappedResp)
		assert.NoError(t, err)
		assert.Equal(t, 20, wrappedResp.Data.PageSize)
	})
}

package employee

import (
	"errors"
	"fmt"
	"idm/inner/common"
	"idm/inner/common/validator"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepo - mock-объект репозитория
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) FindById(id int64) (Entity, error) {
	args := m.Called(id)
	return args.Get(0).(Entity), args.Error(1)
}

func (m *MockRepo) Add(e *Entity) error {
	args := m.Called(e)
	return args.Error(0)
}

func (m *MockRepo) FindAll() ([]Entity, error) {
	args := m.Called()
	return args.Get(0).([]Entity), args.Error(1)
}

func (m *MockRepo) FindByIds(ids []int64) ([]Entity, error) {
	args := m.Called(ids)
	return args.Get(0).([]Entity), args.Error(1)
}

func (m *MockRepo) DeleteById(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRepo) DeleteByIds(ids []int64) error {
	args := m.Called(ids)
	return args.Error(0)
}

func (m *MockRepo) BeginTransaction() (Transaction, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(Transaction), args.Error(1)
}

func (m *MockRepo) FindByNameTx(tx Transaction, name string) (bool, error) {
	args := m.Called(tx, name)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepo) AddTx(tx Transaction, e *Entity) error {
	args := m.Called(tx, e)
	return args.Error(0)
}

func (m *MockRepo) CommitTransaction(tx *sqlx.Tx) error {
	args := m.Called(tx)
	return args.Error(0)
}

func (m *MockRepo) RollbackTransaction(tx *sqlx.Tx) error {
	args := m.Called(tx)
	return args.Error(0)
}

func TestEmployeeService_FindById(t *testing.T) {
	a := assert.New(t)

	t.Run("should return found employee", func(t *testing.T) {
		// Создаем mock репозитория
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)

		// Тестовые данные
		entity := Entity{
			Id:        1,
			Name:      "John Doe",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		want := entity.toResponse()

		// Настраиваем поведение mock
		repo.On("FindById", int64(1)).Return(entity, nil)

		// Вызываем тестируемый метод
		got, err := svc.FindById(1)

		// Проверяем результат
		a.Nil(err)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "FindById", 1))
	})

	t.Run("should return error for invalid id", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)

		// Тестируем валидацию
		response, err := svc.FindById(0)

		a.Empty(response)
		a.NotNil(err)
		a.Contains(err.Error(), "invalid employee id")
		a.True(repo.AssertNumberOfCalls(t, "FindById", 0)) // репозиторий не должен вызываться
	})

	t.Run("should return wrapped error from repository", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)

		// Тестовые данные
		repoErr := errors.New("database error")
		entity := Entity{}
		want := fmt.Errorf("error finding employee with id 1: %w", repoErr)

		// Настраиваем mock для возврата ошибки
		repo.On("FindById", int64(1)).Return(entity, repoErr)

		// Вызываем тестируемый метод
		response, got := svc.FindById(1)

		// Проверяем результат
		a.Empty(response)
		a.NotNil(got)
		a.Equal(want.Error(), got.Error())
		a.True(repo.AssertNumberOfCalls(t, "FindById", 1))
	})
}

func TestEmployeeService_Add(t *testing.T) {
	a := assert.New(t)

	t.Run("should add employee successfully", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)

		// Настраиваем mock - при вызове Add модифицируем Entity
		repo.On("Add", mock.AnythingOfType("*employee.Entity")).Return(nil).Run(func(args mock.Arguments) {
			entity := args.Get(0).(*Entity)
			entity.Id = 1 // симулируем присвоение ID в БД
		})

		// Вызываем тестируемый метод
		got, err := svc.Add("John Doe")

		// Проверяем результат
		a.Nil(err)
		a.Equal(int64(1), got.Id)
		a.Equal("John Doe", got.Name)
		a.True(repo.AssertNumberOfCalls(t, "Add", 1))
	})

	t.Run("should return error for empty name", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)

		// Вызываем с пустым именем
		response, err := svc.Add("")

		// Проверяем результат
		a.Empty(response)
		a.NotNil(err)
		a.Contains(err.Error(), "name cannot be empty")
		a.True(repo.AssertNumberOfCalls(t, "Add", 0)) // репозиторий не должен вызываться
	})

	t.Run("should return wrapped error from repository", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)

		repoErr := errors.New("database error")
		repo.On("Add", mock.AnythingOfType("*employee.Entity")).Return(repoErr)

		// Вызываем тестируемый метод
		response, got := svc.Add("John Doe")

		// Проверяем результат
		a.Empty(response)
		a.NotNil(got)
		a.Contains(got.Error(), "error adding employee")
		a.True(repo.AssertNumberOfCalls(t, "Add", 1))
	})
}

func TestEmployeeService_FindAll(t *testing.T) {
	a := assert.New(t)

	t.Run("should return all employees", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)
		// Тестовые данные
		entities := []Entity{
			{Id: 1, Name: "John Doe", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{Id: 2, Name: "Jane Smith", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}

		repo.On("FindAll").Return(entities, nil)

		// Вызываем тестируемый метод
		got, err := svc.FindAll()

		// Проверяем результат
		a.Nil(err)
		a.Len(got, 2)
		a.Equal("John Doe", got[0].Name)
		a.Equal("Jane Smith", got[1].Name)
		a.True(repo.AssertNumberOfCalls(t, "FindAll", 1))
	})

	t.Run("should return wrapped error from repository", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)

		repoErr := errors.New("database error")
		repo.On("FindAll").Return([]Entity{}, repoErr)

		// Вызываем тестируемый метод
		got, err := svc.FindAll()

		// Проверяем результат
		a.Nil(got)
		a.NotNil(err)
		a.Contains(err.Error(), "error finding all employees")
		a.True(repo.AssertNumberOfCalls(t, "FindAll", 1))
	})
}

func TestEmployeeService_DeleteById(t *testing.T) {
	a := assert.New(t)

	t.Run("should delete employee successfully", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)
		repo.On("DeleteById", int64(1)).Return(nil)

		// Вызываем тестируемый метод
		err := svc.DeleteById(1)

		// Проверяем результат
		a.Nil(err)
		a.True(repo.AssertNumberOfCalls(t, "DeleteById", 1))
	})

	t.Run("should return error for invalid id", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)

		// Вызываем с невалидным ID
		err := svc.DeleteById(0)

		// Проверяем результат
		a.NotNil(err)
		a.Contains(err.Error(), "invalid employee id")
		a.True(repo.AssertNumberOfCalls(t, "DeleteById", 0)) // репозиторий не должен вызываться
	})
}

// StubRepo - stub-объект репозитория (созданный вручную)
type StubRepo struct {
	findByIdFunc func(id int64) (Entity, error)
	addFunc      func(e *Entity) error
	// Остальные методы можем не реализовывать для простоты
}

func (s *StubRepo) FindById(id int64) (Entity, error) {
	if s.findByIdFunc != nil {
		return s.findByIdFunc(id)
	}
	return Entity{}, errors.New("not implemented")
}

func (s *StubRepo) Add(e *Entity) error {
	if s.addFunc != nil {
		return s.addFunc(e)
	}
	return errors.New("not implemented")
}

func (s *StubRepo) FindAll() ([]Entity, error) {
	return nil, errors.New("not implemented")
}

func (s *StubRepo) FindByIds(ids []int64) ([]Entity, error) {
	return nil, errors.New("not implemented")
}

func (s *StubRepo) DeleteById(id int64) error {
	return errors.New("not implemented")
}

func (s *StubRepo) DeleteByIds(ids []int64) error {
	return errors.New("not implemented")
}

func (s *StubRepo) BeginTransaction() (Transaction, error) {
	return nil, errors.New("not implemented")
}

func (s *StubRepo) FindByNameTx(tx Transaction, name string) (bool, error) {
	return false, errors.New("not implemented")
}

func (s *StubRepo) AddTx(tx Transaction, e *Entity) error {
	return errors.New("not implemented")
}

func TestEmployeeService_FindById_WithStub(t *testing.T) {
	a := assert.New(t)

	t.Run("should return found employee using stub", func(t *testing.T) {
		// Создаем stub с предопределенным поведением
		stub := &StubRepo{
			findByIdFunc: func(id int64) (Entity, error) {
				return Entity{
					Id:        id,
					Name:      "Stubbed Employee",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil
			},
		}

		validator := validator.New()
		svc := NewService(stub, validator)

		// Вызываем тестируемый метод
		got, err := svc.FindById(42)

		// Проверяем результат
		a.Nil(err)
		a.Equal(int64(42), got.Id)
		a.Equal("Stubbed Employee", got.Name)
	})

	t.Run("should return error from stub", func(t *testing.T) {
		// Создаем stub, который возвращает ошибку
		stub := &StubRepo{
			findByIdFunc: func(id int64) (Entity, error) {
				return Entity{}, errors.New("stub database error")
			},
		}
		validator := validator.New()
		svc := NewService(stub, validator)

		// Вызываем тестируемый метод
		response, err := svc.FindById(1)

		// Проверяем результат
		a.Empty(response)
		a.NotNil(err)
		a.Contains(err.Error(), "error finding employee with id 1")
		a.Contains(err.Error(), "stub database error")
	})
}

func TestEmployeeService_FindByIds(t *testing.T) {
	a := assert.New(t)

	t.Run("should return employees by ids", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)

		// Тестовые данные
		entities := []Entity{
			{Id: 1, Name: "John Doe", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{Id: 2, Name: "Jane Smith", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}
		ids := []int64{1, 2}

		repo.On("FindByIds", ids).Return(entities, nil)

		// Вызываем тестируемый метод
		got, err := svc.FindByIds(ids)

		// Проверяем результат
		a.Nil(err)
		a.Len(got, 2)
		a.Equal("John Doe", got[0].Name)
		a.Equal("Jane Smith", got[1].Name)
		a.True(repo.AssertNumberOfCalls(t, "FindByIds", 1))
	})

	t.Run("should return wrapped error from repository", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)

		ids := []int64{1, 2}
		repoErr := errors.New("database error")
		repo.On("FindByIds", ids).Return([]Entity{}, repoErr)

		// Вызываем тестируемый метод
		got, err := svc.FindByIds(ids)

		// Проверяем результат
		a.Nil(got)
		a.NotNil(err)
		a.Contains(err.Error(), "error finding employees by ids")
		a.True(repo.AssertNumberOfCalls(t, "FindByIds", 1))
	})
}

func TestEmployeeService_DeleteByIds(t *testing.T) {
	a := assert.New(t)

	t.Run("should delete employees by ids successfully", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)

		ids := []int64{1, 2}
		repo.On("DeleteByIds", ids).Return(nil)

		// Вызываем тестируемый метод
		err := svc.DeleteByIds(ids)

		// Проверяем результат
		a.Nil(err)
		a.True(repo.AssertNumberOfCalls(t, "DeleteByIds", 1))
	})

	t.Run("should return wrapped error from repository", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)

		ids := []int64{1, 2}
		repoErr := errors.New("database error")
		repo.On("DeleteByIds", ids).Return(repoErr)

		// Вызываем тестируемый метод
		err := svc.DeleteByIds(ids)

		// Проверяем результат
		a.NotNil(err)
		a.Contains(err.Error(), "error deleting employees by ids")
		a.True(repo.AssertNumberOfCalls(t, "DeleteByIds", 1))
	})
}

// MockTransaction - мок для sqlx.Tx
type MockTransaction struct {
	mock.Mock
}

func (m *MockTransaction) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTransaction) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTransaction) Get(dest interface{}, query string, args ...interface{}) error {
	mockArgs := m.Called(dest, query, args)
	return mockArgs.Error(0)
}

func (m *MockTransaction) QueryRow(query string, args ...interface{}) Row {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(*MockRow)
}

// MockRow - мок для sql.Row
type MockRow struct {
	mock.Mock
}

func (m *MockRow) Scan(dest ...interface{}) error {
	args := m.Called(dest)
	// Если нужно симулировать присвоение ID
	if len(dest) == 1 {
		if idPtr, ok := dest[0].(*int64); ok {
			*idPtr = 1 // симулируем присвоение ID = 1
		}
	}
	return args.Error(0)
}

// TestValidateRequest - тесты для валидации запросов
func TestService_ValidateRequest(t *testing.T) {
	a := assert.New(t)
	repo := new(MockRepo)
	validator := validator.New()
	svc := NewService(repo, validator)

	t.Run("should validate AddEmployeeRequest successfully", func(t *testing.T) {
		request := AddEmployeeRequest{Name: "John Doe"}

		err := svc.ValidateRequest(request)

		a.Nil(err, "Valid request should not return error")
	})

	t.Run("should return validation error for empty name", func(t *testing.T) {
		request := AddEmployeeRequest{Name: ""}

		err := svc.ValidateRequest(request)

		a.NotNil(err, "Empty name should return validation error")
		validationErr, ok := err.(common.RequestValidationError)
		a.True(ok, "Error should be RequestValidationError type")
		a.Contains(validationErr.Message, "name cannot be empty")
	})

	t.Run("should return validation error for short name", func(t *testing.T) {
		request := AddEmployeeRequest{Name: "J"}

		err := svc.ValidateRequest(request)

		a.NotNil(err)
		validationErr, ok := err.(common.RequestValidationError)
		a.True(ok)
		a.Contains(validationErr.Message, "name must be at least 2 characters long")
	})

	t.Run("should return validation error for long name", func(t *testing.T) {
		longName := string(make([]byte, 101)) // 101 characters
		for i := range longName {
			longName = longName[:i] + "a" + longName[i+1:]
		}
		request := AddEmployeeRequest{Name: longName}

		err := svc.ValidateRequest(request)

		a.NotNil(err)
		validationErr, ok := err.(common.RequestValidationError)
		a.True(ok)
		a.Contains(validationErr.Message, "name must be at most 100 characters long")
	})

	t.Run("should validate FindByIdRequest successfully", func(t *testing.T) {
		request := FindByIdRequest{Id: 1}

		err := svc.ValidateRequest(request)

		a.Nil(err)
	})

	t.Run("should return validation error for invalid ID", func(t *testing.T) {
		request := FindByIdRequest{Id: 0}

		err := svc.ValidateRequest(request)

		a.NotNil(err)
		validationErr, ok := err.(common.RequestValidationError)
		a.True(ok)
		a.Contains(validationErr.Message, "id must be greater than 0")
	})

	t.Run("should validate FindByIdsRequest successfully", func(t *testing.T) {
		request := FindByIdsRequest{Ids: []int64{1, 2, 3}}

		err := svc.ValidateRequest(request)

		a.Nil(err)
	})

	t.Run("should return validation error for empty IDs array", func(t *testing.T) {
		request := FindByIdsRequest{Ids: []int64{}}

		err := svc.ValidateRequest(request)

		a.NotNil(err)
		validationErr, ok := err.(common.RequestValidationError)
		a.True(ok)
		a.Contains(validationErr.Message, "ids")
	})

	t.Run("should return validation error for invalid ID in array", func(t *testing.T) {
		request := FindByIdsRequest{Ids: []int64{1, 0, 3}}

		err := svc.ValidateRequest(request)

		a.NotNil(err)
		validationErr, ok := err.(common.RequestValidationError)
		a.True(ok)
		a.Contains(validationErr.Message, "ids[1] must be greater than 0")
	})
}

// TestAddTransactional - тесты для транзакционного добавления с валидацией
func TestService_AddTransactional(t *testing.T) {
	a := assert.New(t)

	t.Run("should return validation error and not call repository", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)

		// Тестируем с пустым именем
		request := AddEmployeeRequest{Name: ""}

		response, err := svc.AddTransactional(request)

		// Проверяем, что возвращается ошибка валидации
		a.Empty(response)
		a.NotNil(err)
		validationErr, ok := err.(common.RequestValidationError)
		a.True(ok, "Should return RequestValidationError")
		a.Contains(validationErr.Message, "name cannot be empty")

		// Проверяем, что репозиторий не вызывался
		repo.AssertNotCalled(t, "BeginTransaction")
		repo.AssertNotCalled(t, "FindByNameTx")
		repo.AssertNotCalled(t, "AddTx")
	})

	t.Run("should return validation error for short name and not call repository", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)

		request := AddEmployeeRequest{Name: "J"}

		response, err := svc.AddTransactional(request)

		a.Empty(response)
		a.NotNil(err)
		validationErr, ok := err.(common.RequestValidationError)
		a.True(ok)
		a.Contains(validationErr.Message, "name must be at least 2 characters long")

		// Репозиторий не должен вызываться при ошибке валидации
		repo.AssertNotCalled(t, "BeginTransaction")
		repo.AssertNotCalled(t, "FindByNameTx")
		repo.AssertNotCalled(t, "AddTx")
	})

	t.Run("should handle transaction creation error", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)

		request := AddEmployeeRequest{Name: "John Doe"}
		transactionErr := errors.New("failed to create transaction")

		repo.On("BeginTransaction").Return(nil, transactionErr)

		response, err := svc.AddTransactional(request)

		a.Empty(response)
		a.NotNil(err)
		transactionError, ok := err.(common.TransactionError)
		a.True(ok, "Should return TransactionError")
		a.Contains(transactionError.Message, "error creating transaction")
		a.Equal(transactionErr, transactionError.Err)

		// FindByNameTx и AddTx не должны вызываться при ошибке создания транзакции
		repo.AssertNotCalled(t, "FindByNameTx")
		repo.AssertNotCalled(t, "AddTx")
	})

	t.Run("should rollback transaction when employee already exists", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)
		mockTx := new(MockTransaction)

		request := AddEmployeeRequest{Name: "John Doe"}

		repo.On("BeginTransaction").Return(mockTx, nil)
		repo.On("FindByNameTx", mockTx, "John Doe").Return(true, nil)
		mockTx.On("Rollback").Return(nil)

		response, err := svc.AddTransactional(request)

		a.Empty(response)
		a.NotNil(err)
		alreadyExistsErr, ok := err.(common.AlreadyExistsError)
		a.True(ok, "Should return AlreadyExistsError")
		a.Contains(alreadyExistsErr.Message, "employee with name 'John Doe' already exists")

		// Проверяем, что транзакция была откачена
		mockTx.AssertCalled(t, "Rollback")
		// AddTx не должен вызываться, если сотрудник уже существует
		repo.AssertNotCalled(t, "AddTx")
	})

	t.Run("should rollback transaction on AddTx error", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)
		mockTx := new(MockTransaction)

		request := AddEmployeeRequest{Name: "John Doe"}
		addErr := errors.New("database insert error")

		repo.On("BeginTransaction").Return(mockTx, nil)
		repo.On("FindByNameTx", mockTx, "John Doe").Return(false, nil)
		repo.On("AddTx", mockTx, mock.AnythingOfType("*employee.Entity")).Return(addErr)
		mockTx.On("Rollback").Return(nil)

		response, err := svc.AddTransactional(request)

		a.Empty(response)
		a.NotNil(err)
		a.Contains(err.Error(), "error adding employee")
		a.Contains(err.Error(), "database insert error")

		mockTx.AssertCalled(t, "Rollback")
	})

	t.Run("should handle rollback error properly", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)
		mockTx := new(MockTransaction)

		request := AddEmployeeRequest{Name: "John Doe"}
		addErr := errors.New("database insert error")
		rollbackErr := errors.New("rollback failed")

		repo.On("BeginTransaction").Return(mockTx, nil)
		repo.On("FindByNameTx", mockTx, "John Doe").Return(false, nil)
		repo.On("AddTx", mockTx, mock.AnythingOfType("*employee.Entity")).Return(addErr)
		mockTx.On("Rollback").Return(rollbackErr)

		response, err := svc.AddTransactional(request)

		a.Empty(response)
		a.NotNil(err)
		a.Contains(err.Error(), "rolling back transaction errors")
		a.Contains(err.Error(), "database insert error")
		a.Contains(err.Error(), "rollback failed")
	})

	t.Run("should handle commit error and reset response", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)
		mockTx := new(MockTransaction)

		request := AddEmployeeRequest{Name: "John Doe"}
		commitErr := errors.New("commit failed")

		repo.On("BeginTransaction").Return(mockTx, nil)
		repo.On("FindByNameTx", mockTx, "John Doe").Return(false, nil)
		repo.On("AddTx", mockTx, mock.AnythingOfType("*employee.Entity")).Return(nil).Run(func(args mock.Arguments) {
			entity := args.Get(1).(*Entity)
			entity.Id = 1
		})
		mockTx.On("Commit").Return(commitErr)

		response, err := svc.AddTransactional(request)

		a.Empty(response, "Response should be empty on commit error")
		a.NotNil(err)
		a.Contains(err.Error(), "creating employee: commiting transaction error")
		a.Contains(err.Error(), "commit failed")
	})

	t.Run("should successfully add employee with valid data", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)
		mockTx := new(MockTransaction)

		request := AddEmployeeRequest{Name: "John Doe"}

		repo.On("BeginTransaction").Return(mockTx, nil)
		repo.On("FindByNameTx", mockTx, "John Doe").Return(false, nil)
		repo.On("AddTx", mockTx, mock.AnythingOfType("*employee.Entity")).Return(nil).Run(func(args mock.Arguments) {
			entity := args.Get(1).(*Entity)
			entity.Id = 1
		})
		mockTx.On("Commit").Return(nil)

		response, err := svc.AddTransactional(request)

		a.Nil(err)
		a.NotEmpty(response)
		a.Equal(int64(1), response.Id)
		a.Equal("John Doe", response.Name)

		mockTx.AssertCalled(t, "Commit")
		mockTx.AssertNotCalled(t, "Rollback")
	})
}

// TestServiceMethodsWithInvalidData - тесты для проверки обработки некорректных данных в других методах
func TestService_MethodsWithInvalidData(t *testing.T) {
	a := assert.New(t)

	t.Run("FindById should handle validation internally and not call repository", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)

		// Тестируем с невалидным ID
		response, err := svc.FindById(-1)

		a.Empty(response)
		a.NotNil(err)
		validationErr, ok := err.(common.RequestValidationError)
		a.True(ok, "Should return RequestValidationError")
		a.Contains(validationErr.Message, "invalid employee id: -1")

		// Репозиторий не должен вызываться при невалидном ID
		repo.AssertNotCalled(t, "FindById")
	})

	t.Run("Add should validate name and not call repository on empty name", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)

		response, err := svc.Add("")

		a.Empty(response)
		a.NotNil(err)
		validationErr, ok := err.(common.RequestValidationError)
		a.True(ok)
		a.Contains(validationErr.Message, "name cannot be empty")

		// Репозиторий не должен вызываться при невалидном имени
		repo.AssertNotCalled(t, "Add")
	})

	t.Run("Add should validate name length and not call repository", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)

		// Слишком короткое имя
		response, err := svc.Add("J")

		a.Empty(response)
		a.NotNil(err)
		validationErr, ok := err.(common.RequestValidationError)
		a.True(ok)
		a.Contains(validationErr.Message, "name must be at least 2 characters long")

		repo.AssertNotCalled(t, "Add")
	})

	t.Run("DeleteById should validate ID and not call repository", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)

		err := svc.DeleteById(0)

		a.NotNil(err)
		a.Contains(err.Error(), "invalid employee id: 0")

		repo.AssertNotCalled(t, "DeleteById")
	})

	t.Run("DeleteById should validate negative ID and not call repository", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)

		err := svc.DeleteById(-5)

		a.NotNil(err)
		a.Contains(err.Error(), "invalid employee id: -5")

		repo.AssertNotCalled(t, "DeleteById")
	})
}

// TestValidationDoesNotReachDatabase - интеграционные тесты, подтверждающие, что некорректные данные не достигают базы данных
func TestService_ValidationDoesNotReachDatabase(t *testing.T) {
	a := assert.New(t)

	testCases := []struct {
		name        string
		testFunc    func(*MockRepo, *Service) error
		description string
	}{
		{
			name: "empty_name_add",
			testFunc: func(repo *MockRepo, svc *Service) error {
				_, err := svc.Add("")
				return err
			},
			description: "Add with empty name should not reach database",
		},
		{
			name: "short_name_add",
			testFunc: func(repo *MockRepo, svc *Service) error {
				_, err := svc.Add("J")
				return err
			},
			description: "Add with short name should not reach database",
		},
		{
			name: "empty_name_transactional",
			testFunc: func(repo *MockRepo, svc *Service) error {
				_, err := svc.AddTransactional(AddEmployeeRequest{Name: ""})
				return err
			},
			description: "AddTransactional with empty name should not reach database",
		},
		{
			name: "invalid_id_find",
			testFunc: func(repo *MockRepo, svc *Service) error {
				_, err := svc.FindById(0)
				return err
			},
			description: "FindById with invalid ID should not reach database",
		},
		{
			name: "invalid_id_delete",
			testFunc: func(repo *MockRepo, svc *Service) error {
				return svc.DeleteById(-1)
			},
			description: "DeleteById with invalid ID should not reach database",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(MockRepo)
			validator := validator.New()
			svc := NewService(repo, validator)

			err := tc.testFunc(repo, svc)

			// Проверяем, что возвращается ошибка
			a.NotNil(err, tc.description)

			// Проверяем, что никакие методы репозитория не были вызваны
			repo.AssertNotCalled(t, "Add")
			repo.AssertNotCalled(t, "FindById")
			repo.AssertNotCalled(t, "DeleteById")
			repo.AssertNotCalled(t, "BeginTransaction")
			repo.AssertNotCalled(t, "FindByNameTx")
			repo.AssertNotCalled(t, "AddTx")
		})
	}
}

// TestComplexValidationScenarios - тесты сложных сценариев валидации
func TestService_ComplexValidationScenarios(t *testing.T) {
	a := assert.New(t)

	t.Run("multiple validation errors should be handled properly", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)

		// Создаем запрос с множественными ошибками валидации
		request := FindByIdsRequest{Ids: []int64{}} // пустой массив

		err := svc.ValidateRequest(request)

		a.NotNil(err)
		validationErr, ok := err.(common.RequestValidationError)
		a.True(ok)
		// Проверяем, что сообщение содержит информацию об ошибке
		a.Contains(validationErr.Message, "ids")
	})

	t.Run("boundary value testing for name length", func(t *testing.T) {
		repo := new(MockRepo)
		validator := validator.New()
		svc := NewService(repo, validator)

		// Тест с именем ровно 2 символа
		validName := "Jo"
		repo.On("Add", mock.AnythingOfType("*employee.Entity")).Return(nil).Once()
		_, err := svc.Add(validName)
		require.NoError(t, err)
		// Это должно пройти валидацию, но может упасть на уровне репозитория
		// Главное - валидация должна пройти успешно

		// Тест с именем 1 символ (invalid)
		repo2 := new(MockRepo)
		svc2 := NewService(repo2, validator)

		_, err = svc2.Add("J")
		a.NotNil(err)
		validationErr, ok := err.(common.RequestValidationError)
		a.True(ok)
		a.Contains(validationErr.Message, "name must be at least 2 characters long")
		repo2.AssertNotCalled(t, "Add")

		// Тест с именем ровно 100 символов (maximum valid)
		maxValidName := string(make([]rune, 100))
		for i := range maxValidName {
			maxValidName = maxValidName[:i] + "a" + maxValidName[i+1:]
		}
		repo3 := new(MockRepo)
		repo3.On("Add", mock.AnythingOfType("*employee.Entity")).Return(nil).Once()
		svc3 := NewService(repo3, validator)

		_, err = svc3.Add(maxValidName)
		require.NoError(t, err)
		// Валидация должна пройти успешно

		// Тест с именем 101 символ (invalid)
		repo4 := new(MockRepo)

		svc4 := NewService(repo4, validator)

		tooLongName := maxValidName + "a"
		_, err = svc4.Add(tooLongName)
		a.NotNil(err)
		validationErr, ok = err.(common.RequestValidationError)
		a.True(ok)
		a.Contains(validationErr.Message, "name must be at most 100 characters long")
		repo4.AssertNotCalled(t, "Add")
	})
}

// TestBusinessLogicProtectionDetailed - детальные тесты защиты бизнес-логики
func TestBusinessLogicProtectionDetailed(t *testing.T) {
	a := assert.New(t)

	protectionTests := []struct {
		name         string
		testFunc     func(*MockRepo, *Service) error
		description  string
		shouldCallDB bool
	}{
		{
			name: "whitespace_only_name",
			testFunc: func(repo *MockRepo, svc *Service) error {
				_, err := svc.Add("   ")
				return err
			},
			description:  "Whitespace-only name validation",
			shouldCallDB: true, // validator не триммит пробелы по умолчанию
		},
		{
			name: "unicode_name_valid",
			testFunc: func(repo *MockRepo, svc *Service) error {
				repo.On("Add", mock.AnythingOfType("*employee.Entity")).Return(nil).Run(func(args mock.Arguments) {
					entity := args.Get(0).(*Entity)
					entity.Id = 1
					entity.Name = "Владимир"
				})
				_, err := svc.Add("Владимир")
				return err
			},
			description:  "Unicode name should reach database",
			shouldCallDB: true,
		},
		{
			name: "boundary_max_id",
			testFunc: func(repo *MockRepo, svc *Service) error {
				maxId := int64(9223372036854775807)
				repo.On("FindById", maxId).Return(
					Entity{}, common.NotFoundError{Message: "not found"})
				_, err := svc.FindById(maxId)
				return err
			},
			description:  "Maximum int64 ID should reach database",
			shouldCallDB: true,
		},
		{
			name: "multiple_validation_errors",
			testFunc: func(repo *MockRepo, svc *Service) error {
				// Пустой список ID - должен провалить валидацию
				_, err := svc.FindByIds([]int64{})
				return err
			},
			description:  "Multiple validation errors should not reach database",
			shouldCallDB: false,
		},
	}

	for _, pt := range protectionTests {
		t.Run(pt.name, func(t *testing.T) {
			repo := new(MockRepo)

			// ✅ Вместо фиксированного списка On() — ставим switch:
			switch pt.name {
			case "whitespace_only_name":
				repo.On("Add", mock.AnythingOfType("*employee.Entity")).Return(nil)
			case "unicode_name_valid":
				repo.On("Add", mock.AnythingOfType("*employee.Entity")).Return(nil)
			case "boundary_max_id":
				maxId := int64(9223372036854775807)
				repo.On("FindById", maxId).Return(
					Entity{}, common.NotFoundError{Message: "not found"})
			case "multiple_validation_errors":

			}

			validator := validator.New()
			svc := NewService(repo, validator)
			err := pt.testFunc(repo, svc)

			if pt.shouldCallDB {
				// Если должны вызвать БД, то repo.AssertExpectations пройдет
				repo.AssertExpectations(t)
			} else {
				// Если не должны вызывать БД, проверяем что ошибка есть
				a.Error(err, pt.description)
				// И что никакие методы репозитория не вызывались
				repo.AssertNotCalled(t, "Add")
				repo.AssertNotCalled(t, "FindById")
				repo.AssertNotCalled(t, "FindByIds")
				repo.AssertNotCalled(t, "DeleteById")
				repo.AssertNotCalled(t, "DeleteByIds")
			}
		})
	}
}

// TestValidationErrorTypes - тесты типов ошибок валидации
func TestValidationErrorTypes(t *testing.T) {
	a := assert.New(t)
	repo := new(MockRepo)
	validator := validator.New()
	svc := NewService(repo, validator)

	validationErrorTests := []struct {
		name     string
		testFunc func() error
	}{
		{
			name: "add_validation_error",
			testFunc: func() error {
				_, err := svc.Add("")
				return err
			},
		},
		{
			name: "add_transactional_validation_error",
			testFunc: func() error {
				_, err := svc.AddTransactional(AddEmployeeRequest{Name: ""})
				return err
			},
		},
		{
			name: "find_by_id_validation_error",
			testFunc: func() error {
				_, err := svc.FindById(0)
				return err
			},
		},
		{
			name: "delete_by_id_validation_error",
			testFunc: func() error {
				return svc.DeleteById(-1)
			},
		},
	}

	for _, test := range validationErrorTests {
		t.Run(test.name, func(t *testing.T) {
			err := test.testFunc()
			a.Error(err)

			// Проверяем, что это RequestValidationError
			var validationErr common.RequestValidationError
			a.True(errors.As(err, &validationErr), "Should be RequestValidationError")

			// Убеждаемся, что репозиторий не вызывался
			repo.AssertNotCalled(t, "Add")
			repo.AssertNotCalled(t, "FindById")
			repo.AssertNotCalled(t, "DeleteById")
		})
	}
}

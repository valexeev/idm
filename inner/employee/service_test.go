package employee

import (
	"errors"
	"fmt"
	"idm/inner/common/validator"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func (m *MockRepo) BeginTransaction() (*sqlx.Tx, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqlx.Tx), args.Error(1)
}

func (m *MockRepo) FindByNameTx(tx *sqlx.Tx, name string) (bool, error) {
	args := m.Called(tx, name)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepo) AddTx(tx *sqlx.Tx, e *Entity) error {
	args := m.Called(tx, e)
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

func (s *StubRepo) BeginTransaction() (*sqlx.Tx, error) {
	return nil, errors.New("not implemented")
}

func (s *StubRepo) FindByNameTx(tx *sqlx.Tx, name string) (bool, error) {
	return false, errors.New("not implemented")
}

func (s *StubRepo) AddTx(tx *sqlx.Tx, e *Entity) error {
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

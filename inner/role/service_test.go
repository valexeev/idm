package role

import (
	"errors"
	"testing"
	"time"

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

func TestRoleService_FindById(t *testing.T) {
	a := assert.New(t)

	t.Run("should return found role", func(t *testing.T) {
		repo := new(MockRepo)
		svc := NewService(repo)

		entity := Entity{
			Id:        1,
			Name:      "Admin",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		want := entity.toResponse()

		repo.On("FindById", int64(1)).Return(entity, nil)

		got, err := svc.FindById(1)

		a.Nil(err)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "FindById", 1))
	})

	t.Run("should return error for invalid id", func(t *testing.T) {
		repo := new(MockRepo)
		svc := NewService(repo)

		response, err := svc.FindById(-1)

		a.Empty(response)
		a.NotNil(err)
		a.Contains(err.Error(), "invalid role id")
		a.True(repo.AssertNumberOfCalls(t, "FindById", 0))
	})
}

func TestRoleService_Add(t *testing.T) {
	a := assert.New(t)

	t.Run("should add role successfully", func(t *testing.T) {
		repo := new(MockRepo)
		svc := NewService(repo)

		repo.On("Add", mock.AnythingOfType("*role.Entity")).Return(nil).Run(func(args mock.Arguments) {
			entity := args.Get(0).(*Entity)
			entity.Id = 1
		})

		got, err := svc.Add("Admin")

		a.Nil(err)
		a.Equal(int64(1), got.Id)
		a.Equal("Admin", got.Name)
		a.True(repo.AssertNumberOfCalls(t, "Add", 1))
	})

	t.Run("should return error for empty name", func(t *testing.T) {
		repo := new(MockRepo)
		svc := NewService(repo)

		response, err := svc.Add("")

		a.Empty(response)
		a.NotNil(err)
		a.Contains(err.Error(), "name cannot be empty")
		a.True(repo.AssertNumberOfCalls(t, "Add", 0))
	})
}

func TestRoleService_FindAll(t *testing.T) {
	a := assert.New(t)

	t.Run("should return all roles", func(t *testing.T) {
		repo := new(MockRepo)
		svc := NewService(repo)

		entities := []Entity{
			{Id: 1, Name: "Admin", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{Id: 2, Name: "User", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}

		repo.On("FindAll").Return(entities, nil)

		got, err := svc.FindAll()

		a.Nil(err)
		a.Len(got, 2)
		a.Equal("Admin", got[0].Name)
		a.Equal("User", got[1].Name)
		a.True(repo.AssertNumberOfCalls(t, "FindAll", 1))
	})
}

func TestRoleService_DeleteById(t *testing.T) {
	a := assert.New(t)

	t.Run("should delete role successfully", func(t *testing.T) {
		repo := new(MockRepo)
		svc := NewService(repo)

		repo.On("DeleteById", int64(1)).Return(nil)

		err := svc.DeleteById(1)

		a.Nil(err)
		a.True(repo.AssertNumberOfCalls(t, "DeleteById", 1))
	})

	t.Run("should return error for invalid id", func(t *testing.T) {
		repo := new(MockRepo)
		svc := NewService(repo)

		err := svc.DeleteById(0)

		a.NotNil(err)
		a.Contains(err.Error(), "invalid role id")
		a.True(repo.AssertNumberOfCalls(t, "DeleteById", 0))
	})
}

func TestRoleService_FindByIds(t *testing.T) {
	a := assert.New(t)

	t.Run("should return roles by ids", func(t *testing.T) {
		repo := new(MockRepo)
		svc := NewService(repo)

		// Тестовые данные
		entities := []Entity{
			{Id: 1, Name: "Admin", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{Id: 2, Name: "User", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}
		ids := []int64{1, 2}

		repo.On("FindByIds", ids).Return(entities, nil)

		// Вызываем тестируемый метод
		got, err := svc.FindByIds(ids)

		// Проверяем результат
		a.Nil(err)
		a.Len(got, 2)
		a.Equal("Admin", got[0].Name)
		a.Equal("User", got[1].Name)
		a.True(repo.AssertNumberOfCalls(t, "FindByIds", 1))
	})

	t.Run("should return wrapped error from repository", func(t *testing.T) {
		repo := new(MockRepo)
		svc := NewService(repo)

		ids := []int64{1, 2}
		repoErr := errors.New("database error")
		repo.On("FindByIds", ids).Return([]Entity{}, repoErr)

		// Вызываем тестируемый метод
		got, err := svc.FindByIds(ids)

		// Проверяем результат
		a.Nil(got)
		a.NotNil(err)
		a.Contains(err.Error(), "error finding roles by ids")
		a.True(repo.AssertNumberOfCalls(t, "FindByIds", 1))
	})
}

func TestRoleService_DeleteByIds(t *testing.T) {
	a := assert.New(t)

	t.Run("should delete roles by ids successfully", func(t *testing.T) {
		repo := new(MockRepo)
		svc := NewService(repo)

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
		svc := NewService(repo)

		ids := []int64{1, 2}
		repoErr := errors.New("database error")
		repo.On("DeleteByIds", ids).Return(repoErr)

		// Вызываем тестируемый метод
		err := svc.DeleteByIds(ids)

		// Проверяем результат
		a.NotNil(err)
		a.Contains(err.Error(), "error deleting roles by ids")
		a.True(repo.AssertNumberOfCalls(t, "DeleteByIds", 1))
	})
}

// Добавляем недостающие тесты с обработкой ошибок
func TestRoleService_FindById_ErrorHandling(t *testing.T) {
	a := assert.New(t)

	t.Run("should return wrapped error from repository", func(t *testing.T) {
		repo := new(MockRepo)
		svc := NewService(repo)

		entity := Entity{}
		repoErr := errors.New("database error")
		repo.On("FindById", int64(1)).Return(entity, repoErr)

		response, err := svc.FindById(1)

		a.Empty(response)
		a.NotNil(err)
		a.Contains(err.Error(), "error finding role with id 1")
		a.Contains(err.Error(), "database error")
		a.True(repo.AssertNumberOfCalls(t, "FindById", 1))
	})
}

func TestRoleService_Add_ErrorHandling(t *testing.T) {
	a := assert.New(t)

	t.Run("should return wrapped error from repository", func(t *testing.T) {
		repo := new(MockRepo)
		svc := NewService(repo)

		repoErr := errors.New("database error")
		repo.On("Add", mock.AnythingOfType("*role.Entity")).Return(repoErr)

		response, err := svc.Add("Admin")

		a.Empty(response)
		a.NotNil(err)
		a.Contains(err.Error(), "error adding role")
		a.True(repo.AssertNumberOfCalls(t, "Add", 1))
	})
}

func TestRoleService_FindAll_ErrorHandling(t *testing.T) {
	a := assert.New(t)

	t.Run("should return wrapped error from repository", func(t *testing.T) {
		repo := new(MockRepo)
		svc := NewService(repo)

		repoErr := errors.New("database error")
		repo.On("FindAll").Return([]Entity{}, repoErr)

		got, err := svc.FindAll()

		a.Nil(got)
		a.NotNil(err)
		a.Contains(err.Error(), "error finding all roles")
		a.True(repo.AssertNumberOfCalls(t, "FindAll", 1))
	})
}

func TestRoleService_DeleteById_ErrorHandling(t *testing.T) {
	a := assert.New(t)

	t.Run("should return wrapped error from repository", func(t *testing.T) {
		repo := new(MockRepo)
		svc := NewService(repo)

		repoErr := errors.New("database error")
		repo.On("DeleteById", int64(1)).Return(repoErr)

		err := svc.DeleteById(1)

		a.NotNil(err)
		a.Contains(err.Error(), "error deleting role with id 1")
		a.True(repo.AssertNumberOfCalls(t, "DeleteById", 1))
	})
}

// StubRepo - stub-объект репозитория для ролей (созданный вручную)
type StubRepo struct {
	findByIdFunc func(id int64) (Entity, error)
	addFunc      func(e *Entity) error
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

func TestRoleService_FindById_WithStub(t *testing.T) {
	a := assert.New(t)

	t.Run("should return found role using stub", func(t *testing.T) {
		// Создаем stub с предопределенным поведением
		stub := &StubRepo{
			findByIdFunc: func(id int64) (Entity, error) {
				return Entity{
					Id:        id,
					Name:      "Stubbed Role",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil
			},
		}

		svc := NewService(stub)

		// Вызываем тестируемый метод
		got, err := svc.FindById(42)

		// Проверяем результат
		a.Nil(err)
		a.Equal(int64(42), got.Id)
		a.Equal("Stubbed Role", got.Name)
	})

	t.Run("should return error from stub", func(t *testing.T) {
		// Создаем stub, который возвращает ошибку
		stub := &StubRepo{
			findByIdFunc: func(id int64) (Entity, error) {
				return Entity{}, errors.New("stub database error")
			},
		}

		svc := NewService(stub)

		// Вызываем тестируемый метод
		response, err := svc.FindById(1)

		// Проверяем результат
		a.Empty(response)
		a.NotNil(err)
		a.Contains(err.Error(), "error finding role with id 1")
		a.Contains(err.Error(), "stub database error")
	})
}

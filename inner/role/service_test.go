package role

import (
	"context"
	"errors"
	"idm/inner/common"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockValidator struct {
	mock.Mock
}

// MockRepo - mock-объект репозитория
type MockRepo struct {
	mock.Mock
}

func (m *MockValidator) Validate(request any) error {
	args := m.Called(request)
	return args.Error(0)
}

func (m *MockValidator) ValidateWithCustomMessages(request any) error {
	args := m.Called(request)
	return args.Error(0)
}

func (m *MockRepo) FindById(ctx context.Context, id int64) (Entity, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(Entity), args.Error(1)
}

func (m *MockRepo) Add(ctx context.Context, e *Entity) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *MockRepo) FindAll(ctx context.Context) ([]Entity, error) {
	args := m.Called(ctx)
	return args.Get(0).([]Entity), args.Error(1)
}

func (m *MockRepo) FindByIds(ctx context.Context, ids []int64) ([]Entity, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]Entity), args.Error(1)
}

func (m *MockRepo) DeleteById(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepo) DeleteByIds(ctx context.Context, ids []int64) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func TestRoleService_FindById(t *testing.T) {
	a := assert.New(t)

	t.Run("should return found role", func(t *testing.T) {
		repo := new(MockRepo)
		validator := new(MockValidator)
		svc := NewService(repo, validator)

		entity := Entity{
			Id:        1,
			Name:      "Admin",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		want := entity.toResponse()

		// Правильно: валидатор вызывается с int64(1)
		validator.On("ValidateWithCustomMessages", int64(1)).Return(nil)
		repo.On("FindById", mock.Anything, int64(1)).Return(entity, nil)

		got, err := svc.FindById(context.Background(), 1)

		a.Nil(err)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "FindById", 1))

		// Проверка, что мок ожидания выполнены
		validator.AssertExpectations(t)
		repo.AssertExpectations(t)
	})

	t.Run("should return error for invalid id", func(t *testing.T) {
		repo := new(MockRepo)
		validator := new(MockValidator)
		svc := NewService(repo, validator)

		// Для invalid id валидатор не вызывается, репозиторий тоже
		response, err := svc.FindById(context.Background(), -1)

		a.Empty(response)
		a.NotNil(err)
		a.Contains(err.Error(), "invalid role id")
		a.True(repo.AssertNumberOfCalls(t, "FindById", 0))
		a.True(validator.AssertNumberOfCalls(t, "ValidateWithCustomMessages", 0))
	})
}

func TestRoleService_Add(t *testing.T) {
	a := assert.New(t)

	t.Run("should add role successfully", func(t *testing.T) {
		repo := new(MockRepo)
		validator := new(MockValidator)
		svc := NewService(repo, validator)

		// Вызов с конкретным значением структуры AddRoleRequest
		validator.On("ValidateWithCustomMessages", AddRoleRequest{Name: "Admin"}).Return(nil)
		repo.On("Add", mock.Anything, mock.AnythingOfType("*role.Entity")).Return(nil).Run(func(args mock.Arguments) {
			entity := args.Get(1).(*Entity)
			entity.Id = 1
		})

		got, err := svc.Add(context.Background(), "Admin")

		a.Nil(err)
		a.Equal(int64(1), got.Id)
		a.Equal("Admin", got.Name)
		a.True(repo.AssertNumberOfCalls(t, "Add", 1))

		validator.AssertExpectations(t)
		repo.AssertExpectations(t)
	})

	t.Run("should return error for empty name", func(t *testing.T) {
		repo := new(MockRepo)
		validator := new(MockValidator)
		svc := NewService(repo, validator)

		// Ошибка валидации для пустого имени
		validator.On("ValidateWithCustomMessages", AddRoleRequest{Name: ""}).Return(errors.New("name cannot be empty"))

		response, err := svc.Add(context.Background(), "")

		a.Empty(response)
		a.NotNil(err)
		a.Contains(err.Error(), "name cannot be empty")

		validator.AssertExpectations(t)
		repo.AssertExpectations(t)
	})
}

func TestRoleService_FindAll(t *testing.T) {
	a := assert.New(t)

	t.Run("should return all roles", func(t *testing.T) {
		repo := new(MockRepo)
		validator := new(MockValidator) // валидатор не нужен, можно не создавать
		svc := NewService(repo, validator)

		entities := []Entity{
			{Id: 1, Name: "Admin", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{Id: 2, Name: "User", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}

		repo.On("FindAll", mock.Anything).Return(entities, nil)

		got, err := svc.FindAll(context.Background())

		a.Nil(err)
		a.Len(got, 2)
		a.Equal("Admin", got[0].Name)
		a.Equal("User", got[1].Name)
		a.True(repo.AssertNumberOfCalls(t, "FindAll", 1))

		repo.AssertExpectations(t)
	})
}
func TestRoleService_DeleteById(t *testing.T) {
	a := assert.New(t)

	t.Run("should delete role successfully", func(t *testing.T) {
		repo := new(MockRepo)
		validator := new(MockValidator)
		svc := NewService(repo, validator)

		validator.On("ValidateWithCustomMessages", int64(1)).Return(nil)
		repo.On("DeleteById", mock.Anything, int64(1)).Return(nil)

		err := svc.DeleteById(context.Background(), 1)

		a.Nil(err)
		a.True(repo.AssertNumberOfCalls(t, "DeleteById", 1))

		validator.AssertExpectations(t)
		repo.AssertExpectations(t)
	})

	t.Run("should return error for invalid id", func(t *testing.T) {
		repo := new(MockRepo)
		validator := new(MockValidator)
		svc := NewService(repo, validator)

		validator.On("ValidateWithCustomMessages", int64(0)).Return(nil)

		err := svc.DeleteById(context.Background(), 0)

		a.NotNil(err)
		a.Contains(err.Error(), "invalid role id")
		a.True(repo.AssertNumberOfCalls(t, "DeleteById", 0))
		a.True(validator.AssertNumberOfCalls(t, "ValidateWithCustomMessages", 1))
	})
}

func TestRoleService_FindByIds(t *testing.T) {
	a := assert.New(t)

	t.Run("should return roles by ids", func(t *testing.T) {
		repo := new(MockRepo)
		validator := new(MockValidator)
		svc := NewService(repo, validator)

		// Тестовые данные
		entities := []Entity{
			{Id: 1, Name: "Admin", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{Id: 2, Name: "User", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}
		ids := []int64{1, 2}
		validator.On("ValidateWithCustomMessages", mock.MatchedBy(func(arg any) bool {
			s, ok := arg.([]int64)
			if !ok {
				return false
			}
			if len(s) != 2 {
				return false
			}
			return s[0] == 1 && s[1] == 2
		})).Return(nil)

		repo.On("FindByIds", mock.Anything, ids).Return(entities, nil)

		// Вызываем тестируемый метод
		got, err := svc.FindByIds(context.Background(), ids)

		// Проверяем результат
		a.Nil(err)
		a.Len(got, 2)
		a.Equal("Admin", got[0].Name)
		a.Equal("User", got[1].Name)
		a.True(repo.AssertNumberOfCalls(t, "FindByIds", 1))
	})

	t.Run("should return wrapped error from repository", func(t *testing.T) {
		repo := new(MockRepo)
		validator := new(MockValidator)
		svc := NewService(repo, validator)

		ids := []int64{1, 2}
		repoErr := errors.New("database error")
		validator.On("ValidateWithCustomMessages", mock.MatchedBy(func(arg any) bool {
			s, ok := arg.([]int64)
			if !ok {
				return false
			}
			if len(s) != 2 {
				return false
			}
			return s[0] == 1 && s[1] == 2
		})).Return(nil)
		repo.On("FindByIds", mock.Anything, ids).Return([]Entity{}, repoErr)

		// Вызываем тестируемый метод
		got, err := svc.FindByIds(context.Background(), ids)

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
		validator := new(MockValidator)
		svc := NewService(repo, validator)

		ids := []int64{1, 2}
		validator.On("ValidateWithCustomMessages", mock.MatchedBy(func(arg any) bool {
			s, ok := arg.([]int64)
			if !ok {
				return false
			}
			if len(s) != 2 {
				return false
			}
			return s[0] == 1 && s[1] == 2
		})).Return(nil)

		repo.On("DeleteByIds", mock.Anything, ids).Return(nil)

		err := svc.DeleteByIds(context.Background(), ids)

		a.Nil(err)
		a.True(repo.AssertNumberOfCalls(t, "DeleteByIds", 1))
	})

	t.Run("should return wrapped error from repository", func(t *testing.T) {
		repo := new(MockRepo)
		validator := new(MockValidator)
		svc := NewService(repo, validator)

		ids := []int64{1, 2}
		repoErr := errors.New("database error")
		validator.On("ValidateWithCustomMessages", mock.MatchedBy(func(arg any) bool {
			s, ok := arg.([]int64)
			if !ok {
				return false
			}
			if len(s) != 2 {
				return false
			}
			return s[0] == 1 && s[1] == 2
		})).Return(nil)
		repo.On("DeleteByIds", mock.Anything, ids).Return(repoErr)

		// Вызываем тестируемый метод
		err := svc.DeleteByIds(context.Background(), ids)

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
		validator := new(MockValidator)
		svc := NewService(repo, validator)

		entity := Entity{}
		repoErr := errors.New("database error")
		validator.On("ValidateWithCustomMessages", int64(1)).Return(nil)
		repo.On("FindById", mock.Anything, int64(1)).Return(entity, repoErr)

		response, err := svc.FindById(context.Background(), 1)

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
		validator := new(MockValidator)
		svc := NewService(repo, validator)

		repoErr := errors.New("database error")
		repo.On("Add", mock.Anything, mock.AnythingOfType("*role.Entity")).Return(repoErr)
		validator.On("ValidateWithCustomMessages", AddRoleRequest{Name: "Admin"}).Return(nil)

		response, err := svc.Add(context.Background(), "Admin")

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
		validator := new(MockValidator)
		svc := NewService(repo, validator)

		repoErr := errors.New("database error")
		repo.On("FindAll", mock.Anything).Return([]Entity{}, repoErr)

		got, err := svc.FindAll(context.Background())

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
		validator := new(MockValidator)
		svc := NewService(repo, validator)

		validator.On("ValidateWithCustomMessages", int64(1)).Return(nil)
		repo.On("DeleteById", mock.Anything, int64(1)).Return(errors.New("db error"))

		err := svc.DeleteById(context.Background(), 1)

		a.NotNil(err)
		a.Contains(err.Error(), "error deleting role with id 1")
		a.True(repo.AssertNumberOfCalls(t, "DeleteById", 1))
	})
}

// StubRepo - stub-объект репозитория для ролей (созданный вручную)
type StubRepo struct {
	findByIdFunc func(ctx context.Context, id int64) (Entity, error)
	addFunc      func(ctx context.Context, e *Entity) error
}

func (s *StubRepo) FindById(ctx context.Context, id int64) (Entity, error) {
	if s.findByIdFunc != nil {
		return s.findByIdFunc(ctx, id)
	}
	return Entity{}, errors.New("not implemented")
}

func (s *StubRepo) Add(ctx context.Context, e *Entity) error {
	if s.addFunc != nil {
		return s.addFunc(ctx, e)
	}
	return errors.New("not implemented")
}

func (s *StubRepo) FindAll(ctx context.Context) ([]Entity, error) {
	return nil, errors.New("not implemented")
}

func (s *StubRepo) FindByIds(ctx context.Context, ids []int64) ([]Entity, error) {
	return nil, errors.New("not implemented")
}

func (s *StubRepo) DeleteById(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}

func (s *StubRepo) DeleteByIds(ctx context.Context, ids []int64) error {
	return errors.New("not implemented")
}

type StubValidator struct{}

func (s *StubValidator) Validate(request any) error {
	return nil
}

func (s *StubValidator) ValidateWithCustomMessages(request any) error {
	return nil
}

func TestRoleService_FindById_WithStub(t *testing.T) {
	a := assert.New(t)

	t.Run("should return found role using stub", func(t *testing.T) {
		// Создаем stub с предопределенным поведением
		stub := &StubRepo{
			findByIdFunc: func(ctx context.Context, id int64) (Entity, error) {
				return Entity{
					Id:        id,
					Name:      "Stubbed Role",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil
			},
		}

		validator := &StubValidator{}
		svc := NewService(stub, validator)

		// Вызываем тестируемый метод
		got, err := svc.FindById(context.Background(), 42)

		// Проверяем результат
		a.Nil(err)
		a.Equal(int64(42), got.Id)
		a.Equal("Stubbed Role", got.Name)
	})

	t.Run("should return error from stub", func(t *testing.T) {
		// Создаем stub, который возвращает ошибку
		stub := &StubRepo{
			findByIdFunc: func(ctx context.Context, id int64) (Entity, error) {
				return Entity{}, errors.New("stub database error")
			},
		}

		validator := &StubValidator{}
		svc := NewService(stub, validator)

		// Вызываем тестируемый метод
		response, err := svc.FindById(context.Background(), 1)

		// Проверяем результат
		a.Empty(response)
		a.NotNil(err)
		a.Contains(err.Error(), "error finding role with id 1")
		a.Contains(err.Error(), "stub database error")
	})
}

// TestRoleBusinessLogicProtection - тесты защиты бизнес-логики от некорректных данных
func TestRoleBusinessLogicProtection(t *testing.T) {
	a := assert.New(t)

	testCases := []struct {
		name        string
		testFunc    func(*MockRepo, *MockValidator, *Service) error
		description string
	}{
		{
			name: "empty_name_add",
			testFunc: func(repo *MockRepo, validator *MockValidator, svc *Service) error {
				validator.On(
					"ValidateWithCustomMessages",
					AddRoleRequest{Name: ""},
				).Return(common.RequestValidationError{Message: "name cannot be empty"}).Once()

				_, err := svc.Add(context.Background(), "")
				validator.AssertExpectations(t)
				return err
			},
			description: "Add with empty name should not reach database",
		},
		{
			name: "short_name_add",
			testFunc: func(repo *MockRepo, validator *MockValidator, svc *Service) error {
				validator.On(
					"ValidateWithCustomMessages",
					AddRoleRequest{Name: "A"},
				).Return(common.RequestValidationError{Message: "name too short"}).Once()

				_, err := svc.Add(context.Background(), "A")
				validator.AssertExpectations(t)
				return err
			},
			description: "Add with short name should not reach database",
		},
		{
			name: "long_name_add",
			testFunc: func(repo *MockRepo, validator *MockValidator, svc *Service) error {
				longName := strings.Repeat("R", 101)
				validator.On(
					"ValidateWithCustomMessages",
					AddRoleRequest{Name: longName},
				).Return(common.RequestValidationError{Message: "name too long"}).Once()

				_, err := svc.Add(context.Background(), longName)
				validator.AssertExpectations(t)
				return err
			},
			description: "Add with long name should not reach database",
		},
		{
			name: "invalid_id_find",
			testFunc: func(repo *MockRepo, validator *MockValidator, svc *Service) error {
				_, err := svc.FindById(context.Background(), 0)
				return err
			},
			description: "FindById with invalid ID should not reach database",
		},
		{
			name: "negative_id_find",
			testFunc: func(repo *MockRepo, validator *MockValidator, svc *Service) error {
				_, err := svc.FindById(context.Background(), -1)
				return err
			},
			description: "FindById with negative ID should not reach database",
		},
		{
			name: "invalid_id_delete",
			testFunc: func(repo *MockRepo, validator *MockValidator, svc *Service) error {
				validator.On(
					"ValidateWithCustomMessages",
					int64(0),
				).Return(common.RequestValidationError{Message: "invalid id"}).Once()

				err := svc.DeleteById(context.Background(), 0)
				validator.AssertExpectations(t)
				return err
			},
			description: "DeleteById with invalid ID should not reach database",
		},
		{
			name: "negative_id_delete",
			testFunc: func(repo *MockRepo, validator *MockValidator, svc *Service) error {
				validator.On(
					"ValidateWithCustomMessages",
					int64(-5),
				).Return(common.RequestValidationError{Message: "negative id"}).Once()

				err := svc.DeleteById(context.Background(), -5)
				validator.AssertExpectations(t)
				return err
			},
			description: "DeleteById with negative ID should not reach database",
		},
		{
			name: "empty_ids_find",
			testFunc: func(repo *MockRepo, validator *MockValidator, svc *Service) error {
				validator.On(
					"ValidateWithCustomMessages",
					[]int64{},
				).Return(common.RequestValidationError{Message: "empty ids"}).Once()

				_, err := svc.FindByIds(context.Background(), []int64{})
				validator.AssertExpectations(t)
				return err
			},
			description: "FindByIds with empty list should not reach database",
		},
		{
			name: "nil_ids_find",
			testFunc: func(repo *MockRepo, validator *MockValidator, svc *Service) error {
				validator.On(
					"ValidateWithCustomMessages",
					([]int64)(nil),
				).Return(common.RequestValidationError{Message: "nil ids"}).Once()

				_, err := svc.FindByIds(context.Background(), nil)
				validator.AssertExpectations(t)
				return err
			},
			description: "FindByIds with nil list should not reach database",
		},
		{
			name: "invalid_ids_find",
			testFunc: func(repo *MockRepo, validator *MockValidator, svc *Service) error {
				validator.On(
					"ValidateWithCustomMessages",
					[]int64{1, 0, 3},
				).Return(common.RequestValidationError{Message: "invalid ids"}).Once()

				_, err := svc.FindByIds(context.Background(), []int64{1, 0, 3})
				validator.AssertExpectations(t)
				return err
			},
			description: "FindByIds with invalid ID should not reach database",
		},
		{
			name: "empty_ids_delete",
			testFunc: func(repo *MockRepo, validator *MockValidator, svc *Service) error {
				validator.On(
					"ValidateWithCustomMessages",
					[]int64{},
				).Return(common.RequestValidationError{Message: "empty ids"}).Once()

				err := svc.DeleteByIds(context.Background(), []int64{})
				validator.AssertExpectations(t)
				return err
			},
			description: "DeleteByIds with empty list should not reach database",
		},
		{
			name: "invalid_ids_delete",
			testFunc: func(repo *MockRepo, validator *MockValidator, svc *Service) error {
				validator.On(
					"ValidateWithCustomMessages",
					mock.Anything,
				).Return(common.RequestValidationError{Message: "invalid ids"}).Once()

				err := svc.DeleteByIds(context.Background(), []int64{1, 2, 3})
				validator.AssertExpectations(t)
				return err
			},
			description: "DeleteByIds with invalid ID should not reach database",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(MockRepo)
			validator := new(MockValidator)
			svc := NewService(repo, validator)

			err := tc.testFunc(repo, validator, svc)

			a.NotNil(err, tc.description)

			repo.AssertNotCalled(t, "Add")
			repo.AssertNotCalled(t, "FindById")
			repo.AssertNotCalled(t, "FindByIds")
			repo.AssertNotCalled(t, "DeleteById")
			repo.AssertNotCalled(t, "DeleteByIds")
		})
	}
}

// TestRoleValidationErrorTypes - тесты типов ошибок валидации
func TestRoleValidationErrorTypes(t *testing.T) {
	a := assert.New(t)
	repo := new(MockRepo)
	validator := new(MockValidator)
	svc := NewService(repo, validator)

	validationErrorTests := []struct {
		name     string
		testFunc func() error
	}{
		{
			name: "add_validation_error",
			testFunc: func() error {
				_, err := svc.Add(context.Background(), "")
				return err
			},
		},
		{
			name: "find_by_id_validation_error",
			testFunc: func() error {
				_, err := svc.FindById(context.Background(), 0)
				return err
			},
		},
		{
			name: "delete_by_id_validation_error",
			testFunc: func() error {
				return svc.DeleteById(context.Background(), -1)
			},
		},
		{
			name: "find_by_ids_validation_error",
			testFunc: func() error {
				_, err := svc.FindByIds(context.Background(), []int64{})
				return err
			},
		},
		{
			name: "delete_by_ids_validation_error",
			testFunc: func() error {
				err := svc.DeleteByIds(context.Background(), []int64{1, 2})
				return err
			},
		},
		{
			name: "find_by_id_validation_error_context",
			testFunc: func() error {
				_, err := svc.FindById(context.Background(), 123)
				return err
			},
		},
	}

	for _, test := range validationErrorTests {
		t.Run(test.name, func(t *testing.T) {
			switch test.name {
			case "add_validation_error":
				validator.On("ValidateWithCustomMessages", AddRoleRequest{Name: ""}).Return(common.RequestValidationError{Message: "invalid"}).Once()
			case "find_by_id_validation_error":
				validator.On("ValidateWithCustomMessages", int64(0)).Return(common.RequestValidationError{Message: "invalid"}).Once()
			case "find_by_id_validation_error_context":
				validator.On("ValidateWithCustomMessages", mock.Anything).Return(common.RequestValidationError{Message: "invalid"}).Once()
			case "delete_by_id_validation_error":
				validator.On("ValidateWithCustomMessages", int64(-1)).Return(common.RequestValidationError{Message: "invalid"}).Once()
			case "find_by_ids_validation_error":
				validator.On("ValidateWithCustomMessages", []int64{}).Return(common.RequestValidationError{Message: "invalid"}).Once()
			case "delete_by_ids_validation_error":
				validator.On("ValidateWithCustomMessages", mock.Anything).Return(common.RequestValidationError{Message: "invalid"}).Once()
			}
			err := test.testFunc()
			a.Error(err)

			// Проверяем, что это ошибка валидации
			a.Contains(err.Error(), "invalid", "Should contain validation error message")

			// Убеждаемся, что репозиторий не вызывался
			repo.AssertNotCalled(t, "Add")
			repo.AssertNotCalled(t, "FindById")
			repo.AssertNotCalled(t, "FindByIds")
			repo.AssertNotCalled(t, "DeleteById")
			repo.AssertNotCalled(t, "DeleteByIds")
		})
	}
}

// TestRoleBusinessLogicProtectionDetailed - детальные тесты защиты бизнес-логики
func TestRoleBusinessLogicProtectionDetailed(t *testing.T) {
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
				validator := svc.validator.(*MockValidator)
				validator.On("ValidateWithCustomMessages", AddRoleRequest{Name: "   "}).Return(common.RequestValidationError{Message: "name cannot be only whitespace"}).Once()
				_, err := svc.Add(context.Background(), "   ")
				validator.AssertExpectations(t)
				return err
			},
			description:  "Whitespace-only name validation",
			shouldCallDB: false,
		},
		{
			name: "unicode_name_valid",
			testFunc: func(repo *MockRepo, svc *Service) error {
				validator := svc.validator.(*MockValidator)
				validator.On("ValidateWithCustomMessages", AddRoleRequest{Name: "Администратор"}).Return(nil).Once()
				repo.On("Add", mock.Anything, mock.AnythingOfType("*role.Entity")).Return(nil).Run(func(args mock.Arguments) {
					entity := args.Get(1).(*Entity)
					entity.Id = 1
				})
				_, err := svc.Add(context.Background(), "Администратор")
				validator.AssertExpectations(t)
				return err
			},
			description:  "Unicode name should reach database",
			shouldCallDB: true,
		},
		{
			name: "boundary_max_id",
			testFunc: func(repo *MockRepo, svc *Service) error {
				maxId := int64(9223372036854775807)
				validator := svc.validator.(*MockValidator)
				validator.On("ValidateWithCustomMessages", maxId).Return(nil).Once()
				repo.On("FindById", mock.Anything, maxId).Return(Entity{}, errors.New("not found"))
				_, err := svc.FindById(context.Background(), maxId)
				validator.AssertExpectations(t)
				return err
			},
			description:  "Maximum int64 ID should reach database",
			shouldCallDB: true,
		},
		{
			name: "boundary_min_valid_name",
			testFunc: func(repo *MockRepo, svc *Service) error {
				validator := svc.validator.(*MockValidator)
				validator.On("ValidateWithCustomMessages", AddRoleRequest{Name: "AB"}).Return(nil).Once()
				repo.On("Add", mock.Anything, mock.AnythingOfType("*role.Entity")).Return(nil).Run(func(args mock.Arguments) {
					entity := args.Get(1).(*Entity)
					entity.Id = 1
				})
				_, err := svc.Add(context.Background(), "AB") // Минимально допустимое имя
				validator.AssertExpectations(t)
				return err
			},
			description:  "Minimum valid name should reach database",
			shouldCallDB: true,
		},
		{
			name: "boundary_max_valid_name",
			testFunc: func(repo *MockRepo, svc *Service) error {
				maxName := strings.Repeat("R", 100) // Максимально допустимое имя
				validator := svc.validator.(*MockValidator)
				validator.On("ValidateWithCustomMessages", AddRoleRequest{Name: maxName}).Return(nil).Once()
				repo.On("Add", mock.Anything, mock.AnythingOfType("*role.Entity")).Return(nil).Run(func(args mock.Arguments) {
					entity := args.Get(1).(*Entity)
					entity.Id = 1
				})
				_, err := svc.Add(context.Background(), maxName)
				validator.AssertExpectations(t)
				return err
			},
			description:  "Maximum valid name should reach database",
			shouldCallDB: true,
		},
		{
			name: "empty_ids_validation",
			testFunc: func(repo *MockRepo, svc *Service) error {
				validator := svc.validator.(*MockValidator)
				validator.On("ValidateWithCustomMessages", []int64{}).Return(common.RequestValidationError{Message: "ids list cannot be empty"}).Once()
				_, err := svc.FindByIds(context.Background(), []int64{})
				validator.AssertExpectations(t)
				return err
			},
			description:  "Empty IDs list should not reach database",
			shouldCallDB: false,
		},
		{
			name: "mixed_valid_invalid_ids",
			testFunc: func(repo *MockRepo, svc *Service) error {
				validator := svc.validator.(*MockValidator)
				validator.On("ValidateWithCustomMessages", mock.Anything).Return(common.RequestValidationError{Message: "invalid id in list"}).Once()
				_, err := svc.FindByIds(context.Background(), []int64{1, 0, 3})
				validator.AssertExpectations(t)
				return err
			},
			description:  "Mixed valid/invalid IDs should not reach database",
			shouldCallDB: false,
		},
	}

	for _, pt := range protectionTests {
		t.Run(pt.name, func(t *testing.T) {
			repo := new(MockRepo)
			validator := new(MockValidator)
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

// TestRoleEdgeCasesValidation - тесты граничных случаев валидации
func TestRoleEdgeCasesValidation(t *testing.T) {
	a := assert.New(t)
	repo := new(MockRepo)
	validator := new(MockValidator)
	svc := NewService(repo, validator)

	t.Run("special_characters_names", func(t *testing.T) {
		specialNames := []string{
			"Admin-Role",      // Дефис
			"Super User",      // Пробел
			"Role_1",          // Подчеркивание
			"Admin (Primary)", // Скобки
			"Level-2 User",    // Комбинация
		}

		for _, name := range specialNames {
			// Настраиваем мок для валидатора и репозитория для каждого имени
			validator.On("ValidateWithCustomMessages", AddRoleRequest{Name: name}).Return(nil).Once()
			repo.On("Add", mock.Anything, mock.AnythingOfType("*role.Entity")).Return(nil).Run(func(args mock.Arguments) {
				entity := args.Get(1).(*Entity)
				entity.Id = 1
			}).Once()

			_, err := svc.Add(context.Background(), name)
			validator.AssertExpectations(t)
			a.NoError(err, "Name with special characters should be valid: %s", name)
		}

		repo.AssertExpectations(t)
	})

	t.Run("large_ids_list", func(t *testing.T) {
		// Создаем большой список валидных ID
		largeIdsList := make([]int64, 1000)
		for i := range largeIdsList {
			largeIdsList[i] = int64(i + 1)
		}

		validator.On("ValidateWithCustomMessages", largeIdsList).Return(nil).Once()
		repo.On("FindByIds", mock.Anything, largeIdsList).Return([]Entity{}, nil)

		_, err := svc.FindByIds(context.Background(), largeIdsList)
		validator.AssertExpectations(t)
		a.NoError(err, "Large list of valid IDs should be processed")

		repo.AssertExpectations(t)
	})

	t.Run("boundary_ids", func(t *testing.T) {
		boundaryIds := []int64{
			1,                   // Минимальное валидное значение
			9223372036854775807, // Максимальное значение int64
		}

		for _, id := range boundaryIds {
			validator.On("ValidateWithCustomMessages", id).Return(nil).Once()
			repo.On("FindById", mock.Anything, id).Return(Entity{}, errors.New("not found")).Once()

			_, err := svc.FindById(context.Background(), id)
			validator.AssertExpectations(t)
			a.Error(err, "Boundary ID should be processed")
		}

		repo.AssertExpectations(t)
	})
}

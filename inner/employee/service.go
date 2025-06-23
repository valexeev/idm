package employee

import (
	"context"
	"fmt"
	"idm/inner/common"
	"time"
)

// Transaction interface for testability
type Transaction interface {
	Rollback() error
	Commit() error
	Get(dest interface{}, query string, args ...interface{}) error
	QueryRow(query string, args ...interface{}) Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) Row
}

// Row interface for testability
type Row interface {
	Scan(dest ...interface{}) error
}

// Service структура, которая инкапсулирует бизнес-логику
type Service struct {
	repo      Repo
	validator Validator
}

// Repo интерфейс репозитория для сотрудников
type Repo interface {
	FindById(ctx context.Context, id int64) (Entity, error)
	Add(ctx context.Context, e *Entity) error
	FindAll(ctx context.Context) ([]Entity, error)
	FindByIds(ctx context.Context, ids []int64) ([]Entity, error)
	DeleteById(ctx context.Context, id int64) error
	DeleteByIds(ctx context.Context, ids []int64) error
	BeginTransaction(ctx context.Context) (Transaction, error)
	FindByNameTx(ctx context.Context, tx Transaction, name string) (bool, error)
	AddTx(ctx context.Context, tx Transaction, e *Entity) error
}

type Validator interface {
	Validate(any) error
	ValidateWithCustomMessages(any) error
}

func (svc *Service) ValidateRequest(request interface{}) error {
	err := svc.validator.ValidateWithCustomMessages(request)
	if err != nil {
		return common.RequestValidationError{Message: err.Error()}
	}
	return nil
}

// AddTransactional транзакционно добавляет нового сотрудника
// Проверяет наличие сотрудника с таким именем и создает нового, если его нет
func (svc *Service) AddTransactional(ctx context.Context, request AddEmployeeRequest) (response Response, err error) {
	err = svc.validator.ValidateWithCustomMessages(request)
	if err != nil {
		return Response{}, common.RequestValidationError{Message: err.Error()}
	}

	// Начинаем транзакцию
	tx, err := svc.repo.BeginTransaction(ctx)
	if err != nil {
		return Response{}, common.TransactionError{Message: "error creating transaction", Err: err}
	}

	// Отложенная функция завершения транзакции
	defer func() {
		// Проверяем, не было ли паники
		if r := recover(); r != nil {
			panicErr := fmt.Errorf("creating employee panic: %v", r)
			// Если была паника, то откатываем транзакцию
			errTx := tx.Rollback()
			if errTx != nil {
				err = fmt.Errorf("creating employee: rolling back transaction errors: %w, %w", panicErr, errTx)
			} else {
				err = panicErr
			}
		} else if err != nil {
			// Если произошла другая ошибка (не паника), то откатываем транзакцию
			errTx := tx.Rollback()
			if errTx != nil {
				// Формируем сообщение в ожидаемом тестом формате
				err = fmt.Errorf("rolling back transaction errors: %w, rollback error: %w", err, errTx)
			}
		} else {
			// Если ошибок нет, то коммитим транзакцию
			errTx := tx.Commit()
			if errTx != nil {
				err = fmt.Errorf("creating employee: commiting transaction error: %w", errTx)
				response = Response{} // Reset response to empty value on commit error
			}
		}
	}()

	// Проверяем, существует ли сотрудник с таким именем
	exists, err := svc.repo.FindByNameTx(ctx, tx, request.Name)
	if err != nil {
		return Response{}, fmt.Errorf("error checking employee existence: %w", err)
	}

	if exists {
		return Response{}, common.AlreadyExistsError{Message: fmt.Sprintf("employee with name '%s' already exists", request.Name)}
	}

	// Создаем нового сотрудника
	now := time.Now()
	entity := &Entity{
		Name:      request.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = svc.repo.AddTx(ctx, tx, entity)
	if err != nil {
		// Возвращаем ошибку с нужным текстом для теста
		return Response{}, fmt.Errorf("error adding employee: %w", err)
	}

	return entity.toResponse(), nil
}

// NewService функция-конструктор для Service
func NewService(repo Repo, validator Validator) *Service {
	return &Service{
		repo:      repo,
		validator: validator,
	}
}

// FindById возвращает сотрудника по ID
func (svc *Service) FindById(ctx context.Context, id int64) (Response, error) {
	if id <= 0 {
		return Response{}, common.RequestValidationError{Message: fmt.Sprintf("invalid employee id: %d", id)}
	}

	entity, err := svc.repo.FindById(ctx, id)
	if err != nil {
		return Response{}, common.RepositoryError{Message: fmt.Sprintf("error finding employee with id %d", id), Err: err}
	}

	return entity.toResponse(), nil
}

// Add добавляет нового сотрудника
func (svc *Service) Add(ctx context.Context, name string) (Response, error) {
	req := AddEmployeeRequest{Name: name}
	err := svc.validator.ValidateWithCustomMessages(req)
	if err != nil {
		return Response{}, common.RequestValidationError{Message: err.Error()}
	}

	now := time.Now()
	entity := &Entity{
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = svc.repo.Add(ctx, entity)
	if err != nil {
		return Response{}, fmt.Errorf("error adding employee: %w", err)
	}

	return entity.toResponse(), nil
}

// FindAll возвращает всех сотрудников
func (svc *Service) FindAll(ctx context.Context) ([]Response, error) {
	entities, err := svc.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("error finding all employees: %w", err)
	}

	responses := make([]Response, len(entities))
	for i, entity := range entities {
		responses[i] = entity.toResponse()
	}

	return responses, nil
}

// FindByIds возвращает сотрудников по списку ID
func (svc *Service) FindByIds(ctx context.Context, ids []int64) ([]Response, error) {
	// Validate the input before proceeding
	req := FindByIdsRequest{Ids: ids}
	if err := svc.ValidateRequest(req); err != nil {
		return nil, err
	}

	entities, err := svc.repo.FindByIds(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("error finding employees by ids: %w", err)
	}

	responses := make([]Response, len(entities))
	for i, entity := range entities {
		responses[i] = entity.toResponse()
	}
	return responses, nil
}

// DeleteById удаляет сотрудника по ID
func (svc *Service) DeleteById(ctx context.Context, id int64) error {
	if id <= 0 {
		return common.RequestValidationError{Message: fmt.Sprintf("invalid employee id: %d", id)}
	}

	err := svc.repo.DeleteById(ctx, id)
	if err != nil {
		return fmt.Errorf("error deleting employee with id %d: %w", id, err)
	}

	return nil
}

// DeleteByIds удаляет сотрудников по списку ID
func (svc *Service) DeleteByIds(ctx context.Context, ids []int64) error {
	err := svc.repo.DeleteByIds(ctx, ids)
	if err != nil {
		return fmt.Errorf("error deleting employees by ids: %w", err)
	}

	return nil
}

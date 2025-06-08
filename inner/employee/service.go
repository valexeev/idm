package employee

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// Service структура, которая инкапсулирует бизнес-логику
type Service struct {
	repo Repo
}

// Repo интерфейс репозитория для сотрудников
type Repo interface {
	FindById(id int64) (Entity, error)
	Add(e *Entity) error
	FindAll() ([]Entity, error)
	FindByIds(ids []int64) ([]Entity, error)
	DeleteById(id int64) error
	DeleteByIds(ids []int64) error
	BeginTransaction() (*sqlx.Tx, error)
	FindByNameTx(tx *sqlx.Tx, name string) (bool, error)
	AddTx(tx *sqlx.Tx, e *Entity) error
}

// AddTransactional транзакционно добавляет нового сотрудника
// Проверяет наличие сотрудника с таким именем и создает нового, если его нет

func (svc *Service) AddTransactional(name string) (response Response, err error) {
	if name == "" {
		return Response{}, fmt.Errorf("employee name cannot be empty")
	}

	// Начинаем транзакцию
	tx, err := svc.repo.BeginTransaction()
	if err != nil {
		return Response{}, fmt.Errorf("error creating transaction: %w", err)
	}

	// Отложенная функция завершения транзакции
	defer func() {
		// Проверяем, не было ли паники
		if r := recover(); r != nil {
			err = fmt.Errorf("creating employee panic: %v", r)
			// Если была паника, то откатываем транзакцию
			errTx := tx.Rollback()
			if errTx != nil {
				err = fmt.Errorf("creating employee: rolling back transaction errors: %w, %w", err, errTx)
			}
		} else if err != nil {
			// Если произошла другая ошибка (не паника), то откатываем транзакцию
			errTx := tx.Rollback()
			if errTx != nil {
				err = fmt.Errorf("creating employee: rolling back transaction errors: %w, %w", err, errTx)
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
	exists, err := svc.repo.FindByNameTx(tx, name)
	if err != nil {
		return Response{}, fmt.Errorf("error checking employee existence: %w", err)
	}

	if exists {
		return Response{}, fmt.Errorf("employee with name '%s' already exists", name)
	}

	// Создаем нового сотрудника
	now := time.Now()
	entity := &Entity{
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = svc.repo.AddTx(tx, entity)
	if err != nil {
		return Response{}, fmt.Errorf("error adding employee: %w", err)
	}

	return entity.toResponse(), nil
}

// NewService функция-конструктор для Service
func NewService(repo Repo) *Service {
	return &Service{
		repo: repo,
	}
}

// FindById возвращает сотрудника по ID
func (svc *Service) FindById(id int64) (Response, error) {
	if id <= 0 {
		return Response{}, fmt.Errorf("invalid employee id: %d", id)
	}

	entity, err := svc.repo.FindById(id)
	if err != nil {
		return Response{}, fmt.Errorf("error finding employee with id %d: %w", id, err)
	}

	return entity.toResponse(), nil
}

// Add добавляет нового сотрудника
func (svc *Service) Add(name string) (Response, error) {
	if name == "" {
		return Response{}, fmt.Errorf("employee name cannot be empty")
	}

	now := time.Now()
	entity := &Entity{
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := svc.repo.Add(entity)
	if err != nil {
		return Response{}, fmt.Errorf("error adding employee: %w", err)
	}

	return entity.toResponse(), nil
}

// FindAll возвращает всех сотрудников
func (svc *Service) FindAll() ([]Response, error) {
	entities, err := svc.repo.FindAll()
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
func (svc *Service) FindByIds(ids []int64) ([]Response, error) {
	entities, err := svc.repo.FindByIds(ids)
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
func (svc *Service) DeleteById(id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid employee id: %d", id)
	}

	err := svc.repo.DeleteById(id)
	if err != nil {
		return fmt.Errorf("error deleting employee with id %d: %w", id, err)
	}

	return nil
}

// DeleteByIds удаляет сотрудников по списку ID
func (svc *Service) DeleteByIds(ids []int64) error {
	err := svc.repo.DeleteByIds(ids)
	if err != nil {
		return fmt.Errorf("error deleting employees by ids: %w", err)
	}

	return nil
}

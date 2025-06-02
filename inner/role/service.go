package role

import (
	"fmt"
	"time"
)

// Service структура, которая инкапсулирует бизнес-логику
type Service struct {
	repo Repo
}

// Repo интерфейс репозитория для ролей
type Repo interface {
	FindById(id int64) (Entity, error)
	Add(e *Entity) error
	FindAll() ([]Entity, error)
	FindByIds(ids []int64) ([]Entity, error)
	DeleteById(id int64) error
	DeleteByIds(ids []int64) error
}

// NewService функция-конструктор для Service
func NewService(repo Repo) *Service {
	return &Service{
		repo: repo,
	}
}

// FindById возвращает роль по ID
func (svc *Service) FindById(id int64) (Response, error) {
	if id <= 0 {
		return Response{}, fmt.Errorf("invalid role id: %d", id)
	}

	entity, err := svc.repo.FindById(id)
	if err != nil {
		return Response{}, fmt.Errorf("error finding role with id %d: %w", id, err)
	}

	return entity.toResponse(), nil
}

// Add добавляет новую роль
func (svc *Service) Add(name string) (Response, error) {
	if name == "" {
		return Response{}, fmt.Errorf("role name cannot be empty")
	}

	now := time.Now()
	entity := &Entity{
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := svc.repo.Add(entity)
	if err != nil {
		return Response{}, fmt.Errorf("error adding role: %w", err)
	}

	return entity.toResponse(), nil
}

// FindAll возвращает все роли
func (svc *Service) FindAll() ([]Response, error) {
	entities, err := svc.repo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("error finding all roles: %w", err)
	}

	responses := make([]Response, len(entities))
	for i, entity := range entities {
		responses[i] = entity.toResponse()
	}

	return responses, nil
}

// FindByIds возвращает роли по списку ID
func (svc *Service) FindByIds(ids []int64) ([]Response, error) {
	entities, err := svc.repo.FindByIds(ids)
	if err != nil {
		return nil, fmt.Errorf("error finding roles by ids: %w", err)
	}

	responses := make([]Response, len(entities))
	for i, entity := range entities {
		responses[i] = entity.toResponse()
	}

	return responses, nil
}

// DeleteById удаляет роль по ID
func (svc *Service) DeleteById(id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid role id: %d", id)
	}

	err := svc.repo.DeleteById(id)
	if err != nil {
		return fmt.Errorf("error deleting role with id %d: %w", id, err)
	}

	return nil
}

// DeleteByIds удаляет роли по списку ID
func (svc *Service) DeleteByIds(ids []int64) error {
	err := svc.repo.DeleteByIds(ids)
	if err != nil {
		return fmt.Errorf("error deleting roles by ids: %w", err)
	}

	return nil
}

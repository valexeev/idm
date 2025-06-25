package role

import (
	"context"
	"fmt"
	"idm/inner/common"
	"time"
)

// Service структура, которая инкапсулирует бизнес-логику
type Service struct {
	repo      Repo
	validator Validator
}

// Repo интерфейс репозитория для ролей
type Repo interface {
	FindById(ctx context.Context, id int64) (Entity, error)
	Add(ctx context.Context, e *Entity) error
	FindAll(ctx context.Context) ([]Entity, error)
	FindByIds(ctx context.Context, ids []int64) ([]Entity, error)
	DeleteById(ctx context.Context, id int64) error
	DeleteByIds(ctx context.Context, ids []int64) error
}
type Validator interface {
	Validate(any) error
	ValidateWithCustomMessages(any) error
}

// NewService функция-конструктор для Service
func NewService(repo Repo, validator Validator) *Service {
	return &Service{
		repo:      repo,
		validator: validator,
	}
}

func (svc *Service) ValidateRequest(request any) error {
	err := svc.validator.ValidateWithCustomMessages(request)
	if err != nil {
		return common.RequestValidationError{Message: err.Error()}
	}
	return nil
}

// FindById возвращает роль по ID
func (svc *Service) FindById(ctx context.Context, id int64) (Response, error) {
	if id <= 0 {
		return Response{}, common.RequestValidationError{Message: "invalid role id"}
	}

	if err := svc.validator.ValidateWithCustomMessages(id); err != nil {
		return Response{}, err
	}

	entity, err := svc.repo.FindById(ctx, id)
	if err != nil {
		return Response{}, fmt.Errorf("error finding role with id %d: %w", id, err)
	}

	return entity.toResponse(), nil
}

// Add добавляет новую роль
func (svc *Service) Add(ctx context.Context, name string) (Response, error) {
	req := AddRoleRequest{Name: name}
	if err := svc.validator.ValidateWithCustomMessages(req); err != nil {
		return Response{}, err
	}

	now := time.Now()
	entity := &Entity{
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := svc.repo.Add(ctx, entity); err != nil {
		return Response{}, fmt.Errorf("error adding role: %w", err)
	}

	return entity.toResponse(), nil
}

// FindAll возвращает все роли
func (svc *Service) FindAll(ctx context.Context) ([]Response, error) {
	entities, err := svc.repo.FindAll(ctx)
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
func (svc *Service) FindByIds(ctx context.Context, ids []int64) ([]Response, error) {

	if err := svc.validator.ValidateWithCustomMessages(ids); err != nil {
		return nil, err
	}

	// 2) Если всё ок — работаешь с repo
	entities, err := svc.repo.FindByIds(ctx, ids)
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
func (svc *Service) DeleteById(ctx context.Context, id int64) error {
	if err := svc.validator.ValidateWithCustomMessages(id); err != nil {
		return err
	}

	if id <= 0 {
		return common.RequestValidationError{Message: "invalid role id"}
	}
	err := svc.repo.DeleteById(ctx, id)
	if err != nil {
		return fmt.Errorf("error deleting role with id %d: %w", id, err)
	}

	return nil
}

// DeleteByIds удаляет роли по списку ID
func (svc *Service) DeleteByIds(ctx context.Context, ids []int64) error {

	if err := svc.validator.ValidateWithCustomMessages(ids); err != nil {
		return err
	}

	err := svc.repo.DeleteByIds(ctx, ids)
	if err != nil {
		return fmt.Errorf("error deleting roles by ids: %w", err)
	}

	return nil
}

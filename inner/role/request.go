package role

import "time"

type AddRoleRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

func (req *AddRoleRequest) ToEntity() Entity {
	now := time.Now()
	return Entity{
		Name:      req.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type FindByIdRequest struct {
	Id int64 `json:"id" validate:"gt=0"`
}

type FindByIdsRequest struct {
	Ids []int64 `json:"ids" validate:"required,min=1,dive,gt=0"`
}

type DeleteByIdRequest struct {
	Id int64 `json:"id" validate:"gt=0"`
}

type DeleteByIdsRequest struct {
	Ids []int64 `json:"ids" validate:"required,min=1,dive,gt=0"`
}

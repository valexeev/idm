package employee

type AddEmployeeRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

func (req *AddEmployeeRequest) ToEntity() Entity {
	return Entity{Name: req.Name}
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

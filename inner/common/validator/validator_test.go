package validator

import (
	"testing"

	"idm/inner/employee"
	"idm/inner/role"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestValidator_AddEmployeeRequest(t *testing.T) {
	v := New()

	t.Run("valid request", func(t *testing.T) {
		req := employee.AddEmployeeRequest{
			Name: "John Doe",
		}

		err := v.Validate(req)
		assert.NoError(t, err)
	})

	t.Run("empty name", func(t *testing.T) {
		req := employee.AddEmployeeRequest{
			Name: "",
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Equal(t, "Name", validationErrs[0].Field())
		assert.Equal(t, "required", validationErrs[0].Tag())
	})

	t.Run("name too short", func(t *testing.T) {
		req := employee.AddEmployeeRequest{
			Name: "A",
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Equal(t, "Name", validationErrs[0].Field())
		assert.Equal(t, "min", validationErrs[0].Tag())
	})

	t.Run("name too long", func(t *testing.T) {
		longName := make([]byte, 101)
		for i := range longName {
			longName[i] = 'A'
		}

		req := employee.AddEmployeeRequest{
			Name: string(longName),
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Equal(t, "Name", validationErrs[0].Field())
		assert.Equal(t, "max", validationErrs[0].Tag())
	})

	t.Run("name at minimum length", func(t *testing.T) {
		req := employee.AddEmployeeRequest{
			Name: "AB",
		}

		err := v.Validate(req)
		assert.NoError(t, err)
	})

	t.Run("name at maximum length", func(t *testing.T) {
		maxName := make([]byte, 100)
		for i := range maxName {
			maxName[i] = 'A'
		}

		req := employee.AddEmployeeRequest{
			Name: string(maxName),
		}

		err := v.Validate(req)
		assert.NoError(t, err)
	})
}

func TestValidator_FindByIdRequest(t *testing.T) {
	v := New()

	t.Run("valid request", func(t *testing.T) {
		req := employee.FindByIdRequest{
			Id: 123,
		}

		err := v.Validate(req)
		assert.NoError(t, err)
	})

	t.Run("zero id", func(t *testing.T) {
		req := employee.FindByIdRequest{
			Id: 0,
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Equal(t, "Id", validationErrs[0].Field())
		assert.Equal(t, "gt", validationErrs[0].Tag())
	})

	t.Run("negative id", func(t *testing.T) {
		req := employee.FindByIdRequest{
			Id: -1,
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Equal(t, "Id", validationErrs[0].Field())
		assert.Equal(t, "gt", validationErrs[0].Tag())
	})
}

func TestValidator_FindByIdsRequest(t *testing.T) {
	v := New()

	t.Run("valid request", func(t *testing.T) {
		req := employee.FindByIdsRequest{
			Ids: []int64{1, 2, 3},
		}

		err := v.Validate(req)
		assert.NoError(t, err)
	})

	t.Run("empty ids array", func(t *testing.T) {
		req := employee.FindByIdsRequest{
			Ids: []int64{},
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Equal(t, "Ids", validationErrs[0].Field())
		assert.Equal(t, "min", validationErrs[0].Tag()) // Изменено с "required" на "min"
	})

	t.Run("nil ids array", func(t *testing.T) {
		req := employee.FindByIdsRequest{
			Ids: nil,
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Equal(t, "Ids", validationErrs[0].Field())
		assert.Equal(t, "required", validationErrs[0].Tag())
	})

	t.Run("ids with zero value", func(t *testing.T) {
		req := employee.FindByIdsRequest{
			Ids: []int64{1, 0, 3},
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Contains(t, validationErrs[0].Namespace(), "Ids[1]")
		assert.Equal(t, "gt", validationErrs[0].Tag())
	})

	t.Run("ids with negative value", func(t *testing.T) {
		req := employee.FindByIdsRequest{
			Ids: []int64{1, -5, 3},
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Contains(t, validationErrs[0].Namespace(), "Ids[1]")
		assert.Equal(t, "gt", validationErrs[0].Tag())
	})

	t.Run("multiple invalid ids", func(t *testing.T) {
		req := employee.FindByIdsRequest{
			Ids: []int64{0, -1, -2},
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 3) // все три ID невалидны

		for _, validationErr := range validationErrs {
			assert.Equal(t, "gt", validationErr.Tag())
		}
	})
}

func TestValidator_DeleteByIdRequest(t *testing.T) {
	v := New()

	t.Run("valid request", func(t *testing.T) {
		req := employee.DeleteByIdRequest{
			Id: 456,
		}

		err := v.Validate(req)
		assert.NoError(t, err)
	})

	t.Run("zero id", func(t *testing.T) {
		req := employee.DeleteByIdRequest{
			Id: 0,
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Equal(t, "Id", validationErrs[0].Field())
		assert.Equal(t, "gt", validationErrs[0].Tag())
	})

	t.Run("negative id", func(t *testing.T) {
		req := employee.DeleteByIdRequest{
			Id: -10,
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Equal(t, "Id", validationErrs[0].Field())
		assert.Equal(t, "gt", validationErrs[0].Tag())
	})
}

func TestValidator_DeleteByIdsRequest(t *testing.T) {
	v := New()

	t.Run("valid request", func(t *testing.T) {
		req := employee.DeleteByIdsRequest{
			Ids: []int64{10, 20, 30},
		}

		err := v.Validate(req)
		assert.NoError(t, err)
	})

	t.Run("empty ids array", func(t *testing.T) {
		req := employee.DeleteByIdsRequest{
			Ids: []int64{},
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Equal(t, "Ids", validationErrs[0].Field())
		assert.Equal(t, "min", validationErrs[0].Tag()) // Изменено с "required" на "min"
	})

	t.Run("nil ids array", func(t *testing.T) {
		req := employee.DeleteByIdsRequest{
			Ids: nil,
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Equal(t, "Ids", validationErrs[0].Field())
		assert.Equal(t, "required", validationErrs[0].Tag())
	})

	t.Run("ids with zero value", func(t *testing.T) {
		req := employee.DeleteByIdsRequest{
			Ids: []int64{10, 0, 30},
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Contains(t, validationErrs[0].Namespace(), "Ids[1]")
		assert.Equal(t, "gt", validationErrs[0].Tag())
	})

	t.Run("ids with negative value", func(t *testing.T) {
		req := employee.DeleteByIdsRequest{
			Ids: []int64{10, -15, 30},
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Contains(t, validationErrs[0].Namespace(), "Ids[1]")
		assert.Equal(t, "gt", validationErrs[0].Tag())
	})

	t.Run("single valid id", func(t *testing.T) {
		req := employee.DeleteByIdsRequest{
			Ids: []int64{1},
		}

		err := v.Validate(req)
		assert.NoError(t, err)
	})
}

// ===============================
// ТЕСТЫ ДЛЯ СТРУКТУР ИЗ ПАКЕТА ROLE
// ===============================

func TestValidator_Role_AddRoleRequest(t *testing.T) {
	v := New()

	t.Run("valid request", func(t *testing.T) {
		req := role.AddRoleRequest{
			Name: "Admin Role",
		}

		err := v.Validate(req)
		assert.NoError(t, err)
	})

	t.Run("empty name", func(t *testing.T) {
		req := role.AddRoleRequest{
			Name: "",
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Equal(t, "Name", validationErrs[0].Field())
		assert.Equal(t, "required", validationErrs[0].Tag())
	})

	t.Run("name too short", func(t *testing.T) {
		req := role.AddRoleRequest{
			Name: "A",
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Equal(t, "Name", validationErrs[0].Field())
		assert.Equal(t, "min", validationErrs[0].Tag())
	})

	t.Run("name too long", func(t *testing.T) {
		longName := make([]byte, 101)
		for i := range longName {
			longName[i] = 'A'
		}

		req := role.AddRoleRequest{
			Name: string(longName),
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Equal(t, "Name", validationErrs[0].Field())
		assert.Equal(t, "max", validationErrs[0].Tag())
	})

	t.Run("name at minimum length", func(t *testing.T) {
		req := role.AddRoleRequest{
			Name: "AB",
		}

		err := v.Validate(req)
		assert.NoError(t, err)
	})

	t.Run("name at maximum length", func(t *testing.T) {
		maxName := make([]byte, 100)
		for i := range maxName {
			maxName[i] = 'A'
		}

		req := role.AddRoleRequest{
			Name: string(maxName),
		}

		err := v.Validate(req)
		assert.NoError(t, err)
	})
}

func TestValidator_Role_FindByIdRequest(t *testing.T) {
	v := New()

	t.Run("valid request", func(t *testing.T) {
		req := role.FindByIdRequest{
			Id: 123,
		}

		err := v.Validate(req)
		assert.NoError(t, err)
	})

	t.Run("zero id", func(t *testing.T) {
		req := role.FindByIdRequest{
			Id: 0,
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Equal(t, "Id", validationErrs[0].Field())
		assert.Equal(t, "gt", validationErrs[0].Tag())
	})

	t.Run("negative id", func(t *testing.T) {
		req := role.FindByIdRequest{
			Id: -1,
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Equal(t, "Id", validationErrs[0].Field())
		assert.Equal(t, "gt", validationErrs[0].Tag())
	})
}

func TestValidator_Role_FindByIdsRequest(t *testing.T) {
	v := New()

	t.Run("valid request", func(t *testing.T) {
		req := role.FindByIdsRequest{
			Ids: []int64{1, 2, 3},
		}

		err := v.Validate(req)
		assert.NoError(t, err)
	})

	t.Run("empty ids array", func(t *testing.T) {
		req := role.FindByIdsRequest{
			Ids: []int64{},
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Equal(t, "Ids", validationErrs[0].Field())
		assert.Equal(t, "min", validationErrs[0].Tag()) // Изменено с "required" на "min"
	})

	t.Run("nil ids array", func(t *testing.T) {
		req := role.FindByIdsRequest{
			Ids: nil,
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Equal(t, "Ids", validationErrs[0].Field())
		assert.Equal(t, "required", validationErrs[0].Tag())
	})

	t.Run("ids with zero value", func(t *testing.T) {
		req := role.FindByIdsRequest{
			Ids: []int64{10, 0, 30},
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Contains(t, validationErrs[0].Namespace(), "Ids[1]")
		assert.Equal(t, "gt", validationErrs[0].Tag())
	})

	t.Run("ids with negative value", func(t *testing.T) {
		req := role.FindByIdsRequest{
			Ids: []int64{10, -15, 30},
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Contains(t, validationErrs[0].Namespace(), "Ids[1]")
		assert.Equal(t, "gt", validationErrs[0].Tag())
	})

	t.Run("multiple invalid ids", func(t *testing.T) {
		req := role.FindByIdsRequest{
			Ids: []int64{0, -1, -2},
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 3) // все три ID невалидны

		for _, validationErr := range validationErrs {
			assert.Equal(t, "gt", validationErr.Tag())
		}
	})
}

func TestValidator_Role_DeleteByIdRequest(t *testing.T) {
	v := New()

	t.Run("valid request", func(t *testing.T) {
		req := role.DeleteByIdRequest{
			Id: 456,
		}

		err := v.Validate(req)
		assert.NoError(t, err)
	})

	t.Run("zero id", func(t *testing.T) {
		req := role.DeleteByIdRequest{
			Id: 0,
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Equal(t, "Id", validationErrs[0].Field())
		assert.Equal(t, "gt", validationErrs[0].Tag())
	})

	t.Run("negative id", func(t *testing.T) {
		req := role.DeleteByIdRequest{
			Id: -10,
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Equal(t, "Id", validationErrs[0].Field())
		assert.Equal(t, "gt", validationErrs[0].Tag())
	})
}

func TestValidator_Role_DeleteByIdsRequest(t *testing.T) {
	v := New()

	t.Run("valid request", func(t *testing.T) {
		req := role.DeleteByIdsRequest{
			Ids: []int64{10, 20, 30},
		}

		err := v.Validate(req)
		assert.NoError(t, err)
	})

	t.Run("empty ids array", func(t *testing.T) {
		req := role.DeleteByIdsRequest{
			Ids: []int64{},
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Equal(t, "Ids", validationErrs[0].Field())
		assert.Equal(t, "min", validationErrs[0].Tag()) // Изменено с "required" на "min"
	})

	t.Run("nil ids array", func(t *testing.T) {
		req := role.DeleteByIdsRequest{
			Ids: nil,
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Equal(t, "Ids", validationErrs[0].Field())
		assert.Equal(t, "required", validationErrs[0].Tag())
	})

	t.Run("ids with zero value", func(t *testing.T) {
		req := role.DeleteByIdsRequest{
			Ids: []int64{10, 0, 30},
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Contains(t, validationErrs[0].Namespace(), "Ids[1]")
		assert.Equal(t, "gt", validationErrs[0].Tag())
	})

	t.Run("ids with negative value", func(t *testing.T) {
		req := role.DeleteByIdsRequest{
			Ids: []int64{10, -15, 30},
		}

		err := v.Validate(req)
		assert.Error(t, err)

		validationErrs, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Len(t, validationErrs, 1)
		assert.Contains(t, validationErrs[0].Namespace(), "Ids[1]")
		assert.Equal(t, "gt", validationErrs[0].Tag())
	})

	t.Run("single valid id", func(t *testing.T) {
		req := role.DeleteByIdsRequest{
			Ids: []int64{1},
		}

		err := v.Validate(req)
		assert.NoError(t, err)
	})
}

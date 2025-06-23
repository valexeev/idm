package idm_test

import (
	"context"
	"idm/inner/common"
	"idm/inner/database"
	"idm/inner/employee"
	"idm/inner/role"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmployeeRepository(t *testing.T) {
	a := assert.New(t)
	cfg := common.GetConfig(".env.tests")

	var db = database.ConnectDbWithCfg(cfg)
	defer db.Close()

	var clearDatabase = func() {
		db.MustExec("delete from employee")
		db.MustExec("delete from role")
	}

	defer func() {
		if r := recover(); r != nil {
			clearDatabase()
		}
	}()

	var employeeRepository = employee.NewRepository(db)
	var roleRepository = role.NewRepository(db)
	var fixture, fixtureErr = NewFixture(employeeRepository, roleRepository, db)
	a.Nil(fixtureErr)

	t.Run("add employee", func(t *testing.T) {
		emp := &employee.Entity{Name: "John Doe"}
		ctx := context.Background()
		err := employeeRepository.Add(ctx, emp)
		a.Nil(err)
		a.Greater(emp.Id, int64(0))
		clearDatabase()
	})

	t.Run("find employee by id", func(t *testing.T) {
		var newEmployeeId = fixture.MustEmployee("Test Name")
		ctx := context.Background()
		got, err := employeeRepository.FindById(ctx, newEmployeeId)
		a.Nil(err)
		a.Equal(newEmployeeId, got.Id)
		a.Equal("Test Name", got.Name)
		clearDatabase()
	})

	t.Run("find employee by id not found", func(t *testing.T) {
		ctx := context.Background()
		_, err := employeeRepository.FindById(ctx, 999999)
		a.Error(err)
		clearDatabase()
	})

	t.Run("find all employees", func(t *testing.T) {
		fixture.MustEmployee("Employee 1")
		fixture.MustEmployee("Employee 2")
		ctx := context.Background()
		got, err := employeeRepository.FindAll(ctx)
		a.Nil(err)
		a.Len(got, 2)
		clearDatabase()
	})

	t.Run("find employees by ids", func(t *testing.T) {
		id1 := fixture.MustEmployee("Employee 1")
		id2 := fixture.MustEmployee("Employee 2")
		ctx := context.Background()
		got, err := employeeRepository.FindByIds(ctx, []int64{id1, id2})
		a.Nil(err)
		a.Len(got, 2)
		clearDatabase()
	})

	t.Run("delete employee by id", func(t *testing.T) {
		empId := fixture.MustEmployee("To Delete")
		ctx := context.Background()
		err := employeeRepository.DeleteById(ctx, empId)
		a.Nil(err)
		_, err = employeeRepository.FindById(ctx, empId)
		a.Error(err)
		clearDatabase()
	})

	t.Run("delete employees by ids", func(t *testing.T) {
		id1 := fixture.MustEmployee("Delete 1")
		id2 := fixture.MustEmployee("Delete 2")
		ctx := context.Background()
		err := employeeRepository.DeleteByIds(ctx, []int64{id1, id2})
		a.Nil(err)
		all, err := employeeRepository.FindAll(ctx)
		a.Nil(err)
		a.Empty(all)
		clearDatabase()
	})
}

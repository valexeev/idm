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

func TestRoleRepository(t *testing.T) {
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

	var roleRepository = role.NewRepository(db)
	var employeeRepository = employee.NewRepository(db)
	var fixture, fixtureErr = NewFixture(employeeRepository, roleRepository, db)
	a.Nil(fixtureErr)

	t.Run("add role", func(t *testing.T) {
		roleEntity := &role.Entity{Name: "Administrator"}
		err := roleRepository.Add(context.Background(), roleEntity)
		a.Nil(err)
		a.Greater(roleEntity.Id, int64(0))
		clearDatabase()
	})

	t.Run("find role by id", func(t *testing.T) {
		var newRoleId = fixture.MustRole("Manager")
		got, err := roleRepository.FindById(context.Background(), newRoleId)
		a.Nil(err)
		a.Equal(newRoleId, got.Id)
		a.Equal("Manager", got.Name)
		clearDatabase()
	})

	t.Run("find role by id not found", func(t *testing.T) {
		_, err := roleRepository.FindById(context.Background(), 999999)
		a.Error(err)
		clearDatabase()
	})

	t.Run("find all roles", func(t *testing.T) {
		fixture.MustRole("Role 1")
		fixture.MustRole("Role 2")
		got, err := roleRepository.FindAll(context.Background())
		a.Nil(err)
		a.Len(got, 2)
		clearDatabase()
	})

	t.Run("find roles by ids", func(t *testing.T) {
		id1 := fixture.MustRole("Role 1")
		id2 := fixture.MustRole("Role 2")
		got, err := roleRepository.FindByIds(context.Background(), []int64{id1, id2})
		a.Nil(err)
		a.Len(got, 2)
		clearDatabase()
	})

	t.Run("delete role by id", func(t *testing.T) {
		id := fixture.MustRole("ToDelete")
		err := roleRepository.DeleteById(context.Background(), id)
		a.Nil(err)
		_, err = roleRepository.FindById(context.Background(), id)
		a.Error(err)
		clearDatabase()
	})

	t.Run("delete roles by ids", func(t *testing.T) {
		id1 := fixture.MustRole("ToDelete1")
		id2 := fixture.MustRole("ToDelete2")
		err := roleRepository.DeleteByIds(context.Background(), []int64{id1, id2})
		a.Nil(err)
		got, err := roleRepository.FindAll(context.Background())
		a.Nil(err)
		a.Len(got, 0)
		clearDatabase()
	})
}

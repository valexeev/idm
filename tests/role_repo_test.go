package idm_test

import (
	"idm/inner/common"
	"idm/inner/database"
	"idm/inner/employee"
	"idm/inner/role"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoleRepository(t *testing.T) {
	a := assert.New(t)

	cfg, err := common.GetConfig(".env.tests")
	if err != nil {
		t.Fatal("Не удалось загрузить конфиг:", err)
	}

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

	var roleRepository = role.NewRoleRepository(db)
	var employeeRepository = employee.NewEmployeeRepository(db)
	var fixture, fixtureErr = NewFixture(employeeRepository, roleRepository, db)
	a.Nil(fixtureErr)

	t.Run("add role", func(t *testing.T) {
		roleEntity := &role.RoleEntity{Name: "Administrator"}
		err := roleRepository.Add(roleEntity)
		a.Nil(err)
		a.Greater(roleEntity.Id, int64(0))
		clearDatabase()
	})

	t.Run("find role by id", func(t *testing.T) {
		var newRoleId = fixture.MustRole("Manager")
		got, err := roleRepository.FindById(newRoleId)
		a.Nil(err)
		a.Equal(newRoleId, got.Id)
		a.Equal("Manager", got.Name)
		clearDatabase()
	})

	t.Run("find role by id not found", func(t *testing.T) {
		_, err := roleRepository.FindById(999999)
		a.Error(err)
		clearDatabase()
	})

	t.Run("find all roles", func(t *testing.T) {
		fixture.MustRole("Role 1")
		fixture.MustRole("Role 2")
		got, err := roleRepository.FindAll()
		a.Nil(err)
		a.Len(got, 2)
		clearDatabase()
	})

	t.Run("find roles by ids", func(t *testing.T) {
		id1 := fixture.MustRole("Role 1")
		id2 := fixture.MustRole("Role 2")
		got, err := roleRepository.FindByIds([]int64{id1, id2})
		a.Nil(err)
		a.Len(got, 2)
		clearDatabase()
	})

	t.Run("delete role by id", func(t *testing.T) {
		roleId := fixture.MustRole("To Delete")
		err := roleRepository.DeleteById(roleId)
		a.Nil(err)
		_, err = roleRepository.FindById(roleId)
		a.Error(err)
		clearDatabase()
	})

	t.Run("delete roles by ids", func(t *testing.T) {
		id1 := fixture.MustRole("Delete 1")
		id2 := fixture.MustRole("Delete 2")
		err := roleRepository.DeleteByIds([]int64{id1, id2})
		a.Nil(err)
		all, err := roleRepository.FindAll()
		a.Nil(err)
		a.Empty(all)
		clearDatabase()
	})
}

package idm_test

import (
	"database/sql"
	"idm/inner/common"
	"idm/inner/database"
	"idm/inner/employee"
	"idm/inner/role"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmployeeRepository(t *testing.T) {
	a := assert.New(t)

	// Подключение к тестовой базе
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

	var employeeRepository = employee.NewEmployeeRepository(db)
	var roleRepository = role.NewRoleRepository(db)
	var fixture, fixtureErr = NewFixture(employeeRepository, roleRepository, db)
	a.Nil(fixtureErr)

	t.Run("add employee", func(t *testing.T) {
		emp := &employee.EmployeeEntity{Name: "John Doe"}
		err := employeeRepository.Add(emp)
		a.Nil(err)
		a.Greater(emp.Id, int64(0))
		clearDatabase()
	})

	t.Run("find employee by id", func(t *testing.T) {
		var newEmployeeId = fixture.MustEmployee("Test Name")
		got, err := employeeRepository.FindById(newEmployeeId)
		a.Nil(err)
		a.Equal(newEmployeeId, got.Id)
		a.Equal("Test Name", got.Name)
		clearDatabase()
	})

	t.Run("find employee by id not found", func(t *testing.T) {
		_, err := employeeRepository.FindById(999999)
		a.Equal(sql.ErrNoRows, err)
		clearDatabase()
	})

	t.Run("find all employees", func(t *testing.T) {
		fixture.MustEmployee("Employee 1")
		fixture.MustEmployee("Employee 2")
		got, err := employeeRepository.FindAll()
		a.Nil(err)
		a.Len(got, 2)
		clearDatabase()
	})

	t.Run("find employees by ids", func(t *testing.T) {
		id1 := fixture.MustEmployee("Employee 1")
		id2 := fixture.MustEmployee("Employee 2")
		got, err := employeeRepository.FindByIds([]int64{id1, id2})
		a.Nil(err)
		a.Len(got, 2)
		clearDatabase()
	})

	t.Run("delete employee by id", func(t *testing.T) {
		empId := fixture.MustEmployee("To Delete")
		err := employeeRepository.DeleteById(empId)
		a.Nil(err)
		_, err = employeeRepository.FindById(empId)
		a.Equal(sql.ErrNoRows, err)
		clearDatabase()
	})

	t.Run("delete employees by ids", func(t *testing.T) {
		id1 := fixture.MustEmployee("Delete 1")
		id2 := fixture.MustEmployee("Delete 2")
		err := employeeRepository.DeleteByIds([]int64{id1, id2})
		a.Nil(err)
		all, err := employeeRepository.FindAll()
		a.Nil(err)
		a.Empty(all)
		clearDatabase()
	})
}

package gorm_migration

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewMigrationManger_RegisterFunctions_Migrate(t *testing.T) {
	a := assert.New(t)

	db := newMemoryDB()
	mm := NewMigrationManger(db, Migrate)
	list := customerMigrations(mm)

	mm.RegisterFunctions(list...)
	mm.Migrate()
	var columns []SchemaMigration
	db.Order("id ASC").Find(&columns)
	a.Equal(len(list), len(columns))
	a.Equal("createUsersTable", columns[0].FuncName)
	a.Equal("addAvatarToUsers", columns[1].FuncName)
	a.Equal("addEmailIndexToUsers", columns[2].FuncName)

	mm.Type = Rollback
	mm.Migrate()
	var columns2 []SchemaMigration
	db.Order("id ASC").Find(&columns2)
	a.Equal(len(list)-1, len(columns2))

	mm.Type = Rollback
	mm.Migrate()
	columns = []SchemaMigration{}
	db.Order("id ASC").Find(&columns)
	a.Equal(len(list)-2, len(columns))
}

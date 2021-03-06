package gorm_migration

import (
	"gorm.io/gorm"
	"log"
	"reflect"
	"runtime"
	"strings"
)

type changeFunc func() error

// save to database
type SchemaMigration struct {
	gorm.Model
	FuncName string
	f        changeFunc
}

func (sm *SchemaMigration) getFuncName() (name string) {
	if sm.f == nil {
		return ""
	}
	// like: github.com/icbd/go_hightlights/db.(*MigrationManger).createUsersTable-fm
	name = runtime.FuncForPC(reflect.ValueOf(sm.f).Pointer()).Name()
	splits := strings.Split(name, ".")
	// like: createUsersTable-fm
	name = splits[len(splits)-1]
	splits = strings.Split(name, "-")
	// like: createUsersTable
	name = splits[0]
	return name
}

func (sm *SchemaMigration) equal(other *SchemaMigration) bool {
	return sm.FuncName == other.FuncName
}

type MigrationManger struct {
	DB         *gorm.DB
	Type       MigrateType
	Columns    []*SchemaMigration // saved in database
	Migrations []*SchemaMigration // registered by code
}

func NewMigrationManger(db *gorm.DB, t MigrateType) *MigrationManger {
	mm := MigrationManger{DB: db, Type: t}
	mm.checkTable()
	mm.DB.Order("id ASC").Find(&mm.Columns)
	mm.RegisterFunctions()

	return &mm
}

// RegisterFunctions fill it by user
func (mm *MigrationManger) RegisterFunctions(changeFunctions ...changeFunc) *MigrationManger {
	mm.Migrations = []*SchemaMigration{}
	for _, f := range changeFunctions {
		sm := SchemaMigration{f: f}
		sm.FuncName = sm.getFuncName()
		mm.Migrations = append(mm.Migrations, &sm)
	}
	return mm
}

// checkTable check and init schema_migrations table
func (mm *MigrationManger) checkTable() {
	sm := SchemaMigration{}
	migrator := mm.DB.Migrator()
	if !migrator.HasTable(&sm) {
		if err := migrator.CreateTable(&sm); err != nil {
			log.Fatal(err)
		}
	}
}

// IsCompleted fill mm Columns, and check Columns match with Migrations
func (mm *MigrationManger) IsCompleted() bool {
	mm.DB.Order("id ASC").Find(&mm.Columns)

	if len(mm.Columns) != len(mm.Migrations) {
		return false
	}
	for i, record := range mm.Columns {
		if !record.equal(mm.Migrations[i]) {
			return false
		}
	}
	return true
}

// Migrate trigger
func (mm *MigrationManger) Migrate() {
	isCompleted := mm.IsCompleted()
	switch mm.Type {
	case Migrate:
		if !isCompleted {
			mm.migrateUp()
		} else {
			log.Println("Nothing to migrate, all down")
		}
	case Rollback:
		if len(mm.Columns) == 0 {
			log.Println("No more migration to rollback")
		} else {
			mm.migrateDown()
		}
	default: // Check
		if !isCompleted {
			log.Fatal("Please run `-db=migrate` first")
		}
	}
}

func (mm *MigrationManger) migrateUp() {
	max := len(mm.Columns)
	for i, sm := range mm.Migrations {
		if i < max {
			// check
			if !sm.equal(mm.Columns[i]) {
				log.Fatal("Migration Conflict")
			}
		} else {
			// new migration
			if err := sm.f(); err != nil {
				log.Fatal(err)
			}
			mm.DB.Create(&sm)
		}
	}
}

func (mm *MigrationManger) migrateDown() {
	columnCount := len(mm.Columns)
	if columnCount == 0 {
		log.Fatal("No more migration to rollback")
	}

	if err := mm.Migrations[columnCount-1].f(); err != nil {
		log.Fatal(err)
	}
	lastColumn := mm.Columns[columnCount-1]
	mm.DB.Delete(&lastColumn)
	mm.Columns = mm.Columns[0 : columnCount-1]
}

// ChangeTable create or drop table
func (mm *MigrationManger) ChangeTable(dst interface{}) error {
	switch mm.Type {
	case Migrate:
		if !mm.DB.Migrator().HasTable(dst) {
			return mm.DB.Migrator().CreateTable(dst)
		}
	case Rollback:
		if mm.DB.Migrator().HasTable(dst) {
			return mm.DB.Migrator().DropTable(dst)
		}
	}
	return nil
}

// ChangeColumn create or drop table
func (mm *MigrationManger) ChangeColumn(dst interface{}, column string) error {
	hasColumn := mm.DB.Migrator().HasColumn(dst, column)
	switch mm.Type {
	case Migrate:
		if !hasColumn {
			return mm.DB.Migrator().AddColumn(dst, column)
		}
	case Rollback:
		if hasColumn {
			return mm.DB.Migrator().DropColumn(dst, column)
		}
	}
	return nil
}

// Change migrate up or migrate down
func (mm *MigrationManger) Change(up, down changeFunc) error {
	switch mm.Type {
	case Migrate:
		return up()
	case Rollback:
		return down()
	}
	return nil
}

func (mm *MigrationManger) ChangeFuncWrap(sqlLines ...string) changeFunc {
	return func() error {
		return mm.DB.Transaction(func(tx *gorm.DB) error {
			for _, sql := range sqlLines {
				if err := tx.Exec(sql).Error; err != nil {
					return err
				}
			}
			return nil
		})
	}
}

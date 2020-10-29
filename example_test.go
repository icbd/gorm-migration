package gorm_migration

import (
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
)

func (mm *MigrationManger) createUsersTable() error {
	type User struct {
		gorm.Model
		Status       uint8  `gorm:"type:tinyint;comment:enum"`
		Email        string `gorm:"type:varchar(255)"`
		PasswordHash string `gorm:"type:text;comment:BCrypt"`
	}
	return mm.ChangeTable(&User{})
}

func (mm *MigrationManger) addAvatarToUsers() error {
	type User struct {
		Avatar string `gorm:"type:text;comment:Avatar URL"`
	}
	return mm.ChangeColumn(&User{}, "Avatar")
}

func (mm *MigrationManger) addEmailIndexToUsers() error {
	up := mm.ChangeFuncWrap("CREATE UNIQUE INDEX idx_users_on_email ON users (email);")
	down := mm.ChangeFuncWrap("DROP INDEX idx_users_on_email;")
	return mm.Change(up, down)
}

func newMemoryDB() *gorm.DB {
	dbname := "file:" + uuid.New().String() + "?mode=memory"
	db, err := gorm.Open(sqlite.Open(dbname), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func customerMigrations(mm *MigrationManger) []changeFunc {
	return []changeFunc{
		mm.createUsersTable,
		mm.addAvatarToUsers,
		mm.addEmailIndexToUsers,
	}
}

func ExampleMigrate() {
	db := newMemoryDB()
	mm := NewMigrationManger(db, Migrate)
	mm.RegisterFunctions(
		mm.createUsersTable,
		mm.addAvatarToUsers,
		mm.addEmailIndexToUsers)
	mm.Migrate()
}

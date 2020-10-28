package gorm_migration

import "strings"

type MigrateType int8

const (
	Check MigrateType = iota
	Migrate
	Rollback
)

// Set for flag parse
// implement: type Value interface
func (t *MigrateType) Set(value string) error {
	switch strings.ToLower(value) {
	case "migrate":
		*t = Migrate
	case "rollback":
		*t = Rollback
	case "check":
		*t = Check
	default:
		*t = Check
	}
	return nil
}

// String for flag parse
// implement: type Value interface
func (t *MigrateType) String() string {
	if t == nil {
		return ""
	}

	switch *t {
	case Migrate:
		return "migrate"
	case Rollback:
		return "rollback"
	default:
		return "check"
	}
}

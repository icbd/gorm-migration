# GORM-based migration workaround

![GitHub Workflow Status](https://img.shields.io/github/workflow/status/icbd/gorm-migration/Test)
![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/icbd/gorm-migration)

1. Write migration functions with Golang
2. Register functions on `MigrationManger`
3. Execute the `migrate` or `rollback` command

## WHY

If you've used rails, you'll be impressed by its migration management.

`gorm-migration` is a clumsy parody, I've implemented basic migration management, including `Check`, `Migrate` and `Rollback`.

## Install

```shell script
go get github.com/icbd/gorm-migration
```

## How to use

For some types of databases, you need to create the `database` first.

### Migrate

It will perform all the tasks that haven't been run yet.

```go
db := newMemoryDB()
mm := NewMigrationManger(db, Migrate)
mm.RegisterFunctions(
    mm.createUsersTable,
    mm.addAvatarToUsers,
    mm.addEmailIndexToUsers)
mm.Migrate()
```

### Rollback

It will roll it back step by step.

```go
# ...
mm.Type = Rollback
mm.Migrate()
```

### Check

```go
# ...
mm.Type = Check
mm.Migrate()
```

## License

MIT, see [LICENSE](LICENSE)

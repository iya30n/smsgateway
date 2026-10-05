package migrator

import (
	"fmt"
	mysqlAdapter "smsgateway/adapter/mysql"

	migrate "github.com/rubenv/sql-migrate"
)

type Migrator struct {
	adapter    mysqlAdapter.Adapter
	dialect    string
	migrations *migrate.FileMigrationSource
}

func NewMigrator(adapter mysqlAdapter.Adapter, dialect string, migrationsPath string) Migrator {
	return Migrator{
		adapter: adapter,
		dialect: dialect,
		migrations: &migrate.FileMigrationSource{
			Dir: migrationsPath,
		},
	}
}

func (m Migrator) Up() {
	n, err := migrate.Exec(m.adapter.Client(), m.dialect, m.migrations, migrate.Up)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Applied %d migrations!\n", n)
}

func (m Migrator) Down() {
	n, err := migrate.Exec(m.adapter.Client(), m.dialect, m.migrations, migrate.Down)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Applied %d migrations!\n", n)
}

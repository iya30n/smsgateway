package migrator

import (
	mysqlAdapter "smsgateway/adapter/mysql"
	"smsgateway/pkg/logger"

	migrate "github.com/rubenv/sql-migrate"
	"go.uber.org/zap"
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
		logger.Logger.Fatal("failed to apply migrations", zap.Error(err))
	}

	logger.Logger.Info("migrations applied", zap.Int("count", n))
}

func (m Migrator) Down() {
	n, err := migrate.Exec(m.adapter.Client(), m.dialect, m.migrations, migrate.Down)
	if err != nil {
		logger.Logger.Fatal("failed to roll back migrations", zap.Error(err))
	}

	logger.Logger.Info("migrations rolled back", zap.Int("count", n))
}

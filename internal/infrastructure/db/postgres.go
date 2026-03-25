package DB

import (
	Config "autobill-service/internal/infrastructure/config"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PostgresDB struct {
	DB *gorm.DB
}

func CreatePostgresDb(config Config.DatabaseConfig) (*PostgresDB, error) {
	db, err := gorm.Open(postgres.Open(GetDSN(config)), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, err
	}

	if err := runMigrations(db, "./"); err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.MaxLifetime)

	return &PostgresDB{DB: db}, nil
}

func runMigrations(db *gorm.DB, migrationDir string) error {
	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("read migrations directory %q: %w", migrationDir, err)
	}

	migrationFiles := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".sql") {
			migrationFiles = append(migrationFiles, filepath.Join(migrationDir, name))
		}
	}

	sort.Strings(migrationFiles)

	for _, migrationFile := range migrationFiles {
		sqlBytes, readErr := os.ReadFile(migrationFile)
		if readErr != nil {
			return fmt.Errorf("read migration file %q: %w", migrationFile, readErr)
		}

		for _, statement := range splitSQLStatements(string(sqlBytes)) {
			if execErr := db.Exec(statement).Error; execErr != nil {
				return fmt.Errorf("execute migration statement from %q: %w", migrationFile, execErr)
			}
		}
	}

	return nil
}

func splitSQLStatements(content string) []string {
	rawStatements := strings.Split(content, ";")
	statements := make([]string, 0, len(rawStatements))

	for _, raw := range rawStatements {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}

		lines := strings.Split(trimmed, "\n")
		cleanLines := make([]string, 0, len(lines))
		for _, line := range lines {
			lineTrimmed := strings.TrimSpace(line)
			if strings.HasPrefix(lineTrimmed, "--") || lineTrimmed == "" {
				continue
			}
			cleanLines = append(cleanLines, line)
		}

		statement := strings.TrimSpace(strings.Join(cleanLines, "\n"))
		if statement == "" {
			continue
		}

		statements = append(statements, statement)
	}

	return statements
}

func GetDSN(config Config.DatabaseConfig) string {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Name, config.Password, config.SSLMode,
	)
	return dsn
}

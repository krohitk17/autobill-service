package DB

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	DefaultQueryTimeout = 30 * time.Second
)

type PostgresDB struct {
	DB *gorm.DB
}

type DBConfig struct {
	DSN          string
	QueryTimeout time.Duration
	MaxOpenConns int
	MaxIdleConns int
	ConnMaxLife  time.Duration
}

func CreatePostgresDb(dsn string) (*PostgresDB, error) {
	return CreatePostgresDbWithConfig(DBConfig{
		DSN:          dsn,
		QueryTimeout: DefaultQueryTimeout,
		MaxOpenConns: 25,
		MaxIdleConns: 5,
		ConnMaxLife:  5 * time.Minute,
	})
}

func CreatePostgresDbWithConfig(config DBConfig) (*PostgresDB, error) {
	db, err := gorm.Open(postgres.Open(config.DSN), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLife)

	MigrateDB(db)
	return &PostgresDB{DB: db}, nil
}

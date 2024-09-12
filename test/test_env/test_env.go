package test_env

import (
	"context"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	pg "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

type TestEnv struct {
	suite.Suite
	ctx                context.Context
	DB                 *gorm.DB
	pgContainer        *postgres.PostgresContainer
	pgConnectionString string
}

func (suite *TestEnv) SetupSuite() {
	suite.ctx = context.Background()
	pgContainer, err := postgres.Run(
		suite.ctx,
		"postgres:15.3-alpine",
		postgres.WithDatabase("gorm_wrap"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	suite.NoError(err)

	connStr, err := pgContainer.ConnectionString(suite.ctx, "sslmode=disable")
	suite.NoError(err)

	db, err := gorm.Open(pg.Open(connStr), &gorm.Config{})
	suite.NoError(err)

	suite.pgContainer = pgContainer
	suite.pgConnectionString = connStr
	suite.DB = db

	sqlDB, err := suite.DB.DB()
	suite.NoError(err)

	err = sqlDB.Ping()
	suite.NoError(err)

}

func (suite *TestEnv) TearDownSuite() {
	err := suite.pgContainer.Terminate(suite.ctx)
	suite.NoError(err)
}

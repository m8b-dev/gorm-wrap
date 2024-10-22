package main

import (
	"github.com/m8b-dev/gorm-wrap/ezg"
	"github.com/m8b-dev/gorm-wrap/test/test_env"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
)

type GormWrapTestSuite struct {
	test_env.TestEnv
}

type Foo struct {
	gorm.Model
	Length int `gorm:"not null"`
}

type Bar struct {
	gorm.Model
	Width int  `gorm:"not null"`
	FooId uint `gorm:""`
	Foo   Foo  `gorm:"foreignKey:FooId;references:ID"`
}

func (suite *GormWrapTestSuite) SetupTest() {
	suite.NoError(suite.DB.AutoMigrate(&Foo{}, &Bar{}))
}

func (suite *GormWrapTestSuite) TestJoin() {
	err := ezg.W(&Foo{Model: gorm.Model{ID: 50}, Length: 10}).Insert(suite.DB)
	suite.NoError(err)

	err = ezg.W(&Bar{Width: 20, FooId: 50}).Insert(suite.DB)
	suite.NoError(err)

	obj, err := ezg.W(&Foo{}).Join(suite.DB, "bars", "foos.id = bars.foo_id")
	suite.NoError(err)
	suite.NotNil(obj)
}

func TestGormWrap(t *testing.T) {
	suite.Run(t, new(GormWrapTestSuite))
}

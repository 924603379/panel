package services

import (
	"testing"

	"github.com/goravel/framework/testing/mock"
	"github.com/stretchr/testify/suite"
	"panel/app/models"
)

type SettingTestSuite struct {
	suite.Suite
	setting Setting
}

func TestSettingTestSuite(t *testing.T) {
	suite.Run(t, &SettingTestSuite{
		setting: NewSettingImpl(),
	})
}

func (s *SettingTestSuite) SetupTest() {

}

func (s *SettingTestSuite) TestGet() {
	mockOrm, mockDb, _, _ := mock.Orm()
	mockOrm.On("Query").Return(mockDb)
	mockDb.On("Where", "key", "test").Return(mockDb)
	mockDb.On("FirstOrFail", &models.Setting{}).Return(nil)
	a := s.setting.Get("test")
	s.Equal("", a)
	mockDb.On("FirstOrFail", &models.Setting{}).Return(nil)
	b := s.setting.Get("test", "test")
	s.Equal("test", b)
	mockOrm.AssertExpectations(s.T())
	mockDb.AssertExpectations(s.T())
}

func (s *SettingTestSuite) TestSet() {
	mockOrm, mockDb, _, _ := mock.Orm()
	mockOrm.On("Query").Return(mockDb)
	mockDb.On("Where", "key", "test").Return(mockDb)
	mockDb.On("UpdateOrCreate", &models.Setting{}, models.Setting{Key: "test"}, models.Setting{Value: "test"}).Return(nil)
	err := s.setting.Set("test", "test")
	s.Nil(err)
	mockOrm.AssertExpectations(s.T())
	mockDb.AssertExpectations(s.T())
}

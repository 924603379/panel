package tools

import (
	"testing"

	"github.com/goravel/framework/support/env"
	"github.com/stretchr/testify/suite"
)

type HelperTestSuite struct {
	suite.Suite
}

func TestHelperTestSuite(t *testing.T) {
	suite.Run(t, &HelperTestSuite{})
}

func (s *HelperTestSuite) TestGetMonitoringInfo() {
	s.NotNil(GetMonitoringInfo())
}

func (s *HelperTestSuite) TestVersionCompare() {
	s.True(VersionCompare("1.0.0", "1.0.0", "=="))
	s.True(VersionCompare("1.0.0", "1.0.0", ">="))
	s.True(VersionCompare("1.0.0", "1.0.0", "<="))
	s.False(VersionCompare("1.0.0", "1.0.0", ">"))
	s.False(VersionCompare("1.0.0", "1.0.0", "<"))
	s.False(VersionCompare("1.0.0", "1.0.0", "!="))

	s.True(VersionCompare("1.0.0", "1.0.1", "<"))
	s.True(VersionCompare("1.0.0", "1.0.1", "<="))
	s.True(VersionCompare("1.0.0", "1.0.1", "!="))
	s.False(VersionCompare("1.0.0", "1.0.1", "=="))
	s.False(VersionCompare("1.0.0", "1.0.1", ">="))
	s.False(VersionCompare("1.0.0", "1.0.1", ">"))

	s.True(VersionCompare("1.0.1", "1.0.0", ">"))
	s.True(VersionCompare("1.0.1", "1.0.0", ">="))
	s.True(VersionCompare("1.0.1", "1.0.0", "!="))
	s.False(VersionCompare("1.0.1", "1.0.0", "=="))
	s.False(VersionCompare("1.0.1", "1.0.0", "<="))
	s.False(VersionCompare("1.0.1", "1.0.0", "<"))

	s.True(VersionCompare("v1.0.0", "1.0.0", "=="))
	s.True(VersionCompare("1.0.0", "v1.0.0", "=="))
	s.True(VersionCompare("v1.0.0", "v1.0.0", "=="))
}

func (s *HelperTestSuite) TestGenerateVersions() {
	versions, err := GenerateVersions("1.0.0", "1.0.3")
	s.NoError(err)
	s.Equal([]string{"1.0.1", "1.0.2", "1.0.3"}, versions)

	versions, err = GenerateVersions("v1.0.0", "v1.0.3")
	s.NoError(err)
	s.Equal([]string{"1.0.1", "1.0.2", "1.0.3"}, versions)

	versions, err = GenerateVersions("1.0.0", "1.0.0")
	s.NoError(err)
	s.Equal([]string(nil), versions)

	versions, err = GenerateVersions("1.0.0", "1.1.1")
	s.NoError(err)
	s.Equal([]string{
		"1.0.1", "1.0.2", "1.0.3", "1.0.4", "1.0.5", "1.0.6", "1.0.7", "1.0.8", "1.0.9", "1.0.10",
		"1.0.11", "1.0.12", "1.0.13", "1.0.14", "1.0.15", "1.0.16", "1.0.17", "1.0.18", "1.0.19", "1.0.20",
		"1.0.21", "1.0.22", "1.0.23", "1.0.24", "1.0.25", "1.0.26", "1.0.27", "1.0.28", "1.0.29", "1.0.30",
		"1.0.31", "1.0.32", "1.0.33", "1.0.34", "1.0.35", "1.0.36", "1.0.37", "1.0.38", "1.0.39", "1.0.40",
		"1.0.41", "1.0.42", "1.0.43", "1.0.44", "1.0.45", "1.0.46", "1.0.47", "1.0.48", "1.0.49", "1.0.50",
		"1.0.51", "1.0.52", "1.0.53", "1.0.54", "1.0.55", "1.0.56", "1.0.57", "1.0.58", "1.0.59", "1.0.60",
		"1.0.61", "1.0.62", "1.0.63", "1.0.64", "1.0.65", "1.0.66", "1.0.67", "1.0.68", "1.0.69", "1.0.70",
		"1.0.71", "1.0.72", "1.0.73", "1.0.74", "1.0.75", "1.0.76", "1.0.77", "1.0.78", "1.0.79", "1.0.80",
		"1.0.81", "1.0.82", "1.0.83", "1.0.84", "1.0.85", "1.0.86", "1.0.87", "1.0.88", "1.0.89", "1.0.90",
		"1.0.91", "1.0.92", "1.0.93", "1.0.94", "1.0.95", "1.0.96", "1.0.97", "1.0.98", "1.0.99", "1.1.0",
		"1.1.1",
	}, versions)

	versions, err = GenerateVersions("1..0", "1.0.1")
	s.Error(err)
	s.Nil(versions)

	versions, err = GenerateVersions("1.0.0", "1..1")
	s.Error(err)
	s.Nil(versions)
}

func (s *HelperTestSuite) TestGetLatestPanelVersion() {
	if env.IsWindows() {
		return
	}
	version, err := GetLatestPanelVersion()
	s.NotEmpty(version)
	s.Nil(err)
}

func (s *HelperTestSuite) TestGetPanelVersion() {
	if env.IsWindows() {
		return
	}
	version, err := GetPanelVersion("v2.0.58")
	s.NotEmpty(version)
	s.Nil(err)
}

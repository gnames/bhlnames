package config_test

import (
	"path/filepath"
	"testing"

	"github.com/gnames/bhlnames/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	assert := assert.New(t)
	test := config.Config{
		BHLDumpURL:  "http://opendata.globalnames.org/bhlnames/bhl-data.zip",
		BHLNamesURL: "http://opendata.globalnames.org/bhlnames/names.zip",
		CoLDataURL:  "http://opendata.globalnames.org/bhlnames/col.zip",
		RootDir:     config.RootDir(),
		DbHost:      "0.0.0.0",
		DbUser:      "postgres",
		DbPass:      "postgres",
		DbDatabase:  "bhlnames",
		JobsNum:     4,
		PortREST:    8888,
	}
	test.DownloadBHLFile = filepath.Join(test.RootDir, "bhl-data.zip")
	test.DownloadNamesFile = filepath.Join(test.RootDir, "bhlindex-latest.zip")
	test.DownloadCoLFile = filepath.Join(test.RootDir, "col.zip")
	test.ExtractDir = filepath.Join(test.RootDir, "Data")

	cfg := config.New()
	assert.Equal(test, cfg)
}

func TestModifiedConfig(t *testing.T) {
	assert := assert.New(t)
	test := config.Config{
		BHLDumpURL:      "https://example.org",
		BHLNamesURL:     "https://example.org",
		CoLDataURL:      "https://example.org",
		RootDir:         "/tmp",
		DbHost:          "10.0.0.10",
		DbUser:          "john",
		DbPass:          "doe",
		DbDatabase:      "bhl",
		JobsNum:         100,
		PortREST:        80,
		WithRebuild:     true,
		WithCoLDataTrim: true,
	}

	test.DownloadBHLFile = filepath.Join(test.RootDir, "bhl-data.zip")
	test.DownloadNamesFile = filepath.Join(test.RootDir, "bhlindex-latest.zip")
	test.DownloadCoLFile = filepath.Join(test.RootDir, "col.zip")
	test.ExtractDir = filepath.Join(test.RootDir, "Data")
	cfg := modConfig()
	assert.Equal(test, cfg)
}

func modConfig() config.Config {
	opts := []config.Option{
		config.OptBHLDumpURL("https://example.org"),
		config.OptBHLNamesURL("https://example.org"),
		config.OptCoLDataURL("https://example.org"),
		config.OptRootDir("/tmp"),
		config.OptDbHost("10.0.0.10"),
		config.OptDbUser("john"),
		config.OptDbPass("doe"),
		config.OptDbDatabase("bhl"),
		config.OptJobsNum(100),
		config.OptPortREST(80),
		config.OptWithRebuild(true),
	}
	return config.New(opts...)
}

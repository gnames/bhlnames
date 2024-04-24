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
		BHLDumpURL:  "http://opendata.globalnames.org/dumps/bhl-data.zip",
		BHLNamesURL: "http://opendata.globalnames.org/dumps/bhl-col.zip",
		CoLDataURL:  "https://api.checklistbank.org/dataset/3LR/export?format=dwca",
		InputDir:    config.InputDir(),
		DbHost:      "0.0.0.0",
		DbUser:      "postgres",
		DbPass:      "postgres",
		DbDatabase:  "bhlnames",
		JobsNum:     4,
		PortREST:    8888,
	}
	test.DownloadBHLFile = filepath.Join(test.InputDir, "bhl-data.zip")
	test.DownloadNamesFile = filepath.Join(test.InputDir, "bhlindex-latest.zip")
	test.DownloadCoLFile = filepath.Join(test.InputDir, "col.zip")
	test.DownloadDir = filepath.Join(test.InputDir, "Data")

	cfg := config.New()
	assert.Equal(test, cfg)
}

func TestModifiedConfig(t *testing.T) {
	assert := assert.New(t)
	test := config.Config{
		BHLDumpURL:    "https://example.org",
		BHLNamesURL:   "https://example.org",
		CoLDataURL:    "https://example.org",
		InputDir:      "/tmp",
		DbHost:        "10.0.0.10",
		DbUser:        "john",
		DbPass:        "doe",
		DbDatabase:    "bhl",
		JobsNum:       100,
		PortREST:      80,
		WithRebuild:   true,
		WithCoLRecalc: true,
	}

	test.DownloadBHLFile = filepath.Join(test.InputDir, "bhl-data.zip")
	test.DownloadNamesFile = filepath.Join(test.InputDir, "bhlindex-latest.zip")
	test.DownloadCoLFile = filepath.Join(test.InputDir, "col.zip")
	test.DownloadDir = filepath.Join(test.InputDir, "Data")
	cfg := modConfig()
	assert.Equal(test, cfg)
}

func modConfig() config.Config {
	opts := []config.Option{
		config.OptBHLDumpURL("https://example.org"),
		config.OptBHLNamesURL("https://example.org"),
		config.OptCoLDataURL("https://example.org"),
		config.OptInputDir("/tmp"),
		config.OptDbHost("10.0.0.10"),
		config.OptDbUser("john"),
		config.OptDbPass("doe"),
		config.OptDbName("bhl"),
		config.OptJobsNum(100),
		config.OptPortREST(80),
		config.OptWithRebuild(true),
	}
	return config.New(opts...)
}

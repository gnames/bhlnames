package config_test

import (
	"path/filepath"
	"testing"

	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/gnfmt"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	test := config.Config{
		BHLDumpURL:   "http://opendata.globalnames.org/dumps/bhl-data.zip",
		BHLNamesURL:  "http://opendata.globalnames.org/dumps/bhl-col.zip",
		InputDir:     config.InputDir(),
		Delimiter:    ',',
		DbHost:       "0.0.0.0",
		DbUser:       "postgres",
		DbPass:       "postgres",
		DbDatabase:   "bhlnames",
		JobsNum:      4,
		PortREST:     8888,
		Format:       gnfmt.CSV,
		WithSynonyms: true,
	}
	test.DownloadBHLFile = filepath.Join(test.InputDir, "bhl-data.zip")
	test.DownloadNamesFile = filepath.Join(test.InputDir, "bhl-names.zip")
	test.DownloadDir = filepath.Join(test.InputDir, "Data")
	test.PageDir = filepath.Join(test.InputDir, "page")
	test.PageFileDir = filepath.Join(test.InputDir, "page-file")
	test.PartDir = filepath.Join(test.InputDir, "part")
	test.AhoCorasickDir = filepath.Join(test.InputDir, "ac")
	test.AhoCorKeyValDir = filepath.Join(test.InputDir, "ackv")

	cfg := config.New()
	assert.Equal(t, test, cfg)
}

func TestModifiedConfig(t *testing.T) {
	test := config.Config{
		BHLDumpURL:          "https://example.org",
		BHLNamesURL:         "https://example.org",
		InputDir:            "/tmp",
		Delimiter:           '\t',
		DbHost:              "10.0.0.10",
		DbUser:              "john",
		DbPass:              "doe",
		DbDatabase:          "bhl",
		JobsNum:             100,
		PortREST:            80,
		Format:              gnfmt.CompactJSON,
		WithSynonyms:        false,
		WithRebuild:         true,
		SortDesc:            true,
		WithShortenedOutput: true,
	}

	test.DownloadBHLFile = filepath.Join(test.InputDir, "bhl-data.zip")
	test.DownloadNamesFile = filepath.Join(test.InputDir, "bhl-names.zip")
	test.DownloadDir = filepath.Join(test.InputDir, "Data")
	test.PageDir = filepath.Join(test.InputDir, "page")
	test.PageFileDir = filepath.Join(test.InputDir, "page-file")
	test.PartDir = filepath.Join(test.InputDir, "part")
	test.AhoCorasickDir = filepath.Join(test.InputDir, "ac")
	test.AhoCorKeyValDir = filepath.Join(test.InputDir, "ackv")
	cfg := modConfig()
	assert.Equal(t, test, cfg)
}

func modConfig() config.Config {
	opts := []config.Option{
		config.OptBHLDumpURL("https://example.org"),
		config.OptBHLNamesURL("https://example.org"),
		config.OptInputDir("/tmp"),
		config.OptDbHost("10.0.0.10"),
		config.OptDbUser("john"),
		config.OptDbPass("doe"),
		config.OptDelimiter('\t'),
		config.OptDbName("bhl"),
		config.OptJobsNum(100),
		config.OptPortREST(80),
		config.OptFormat(gnfmt.CompactJSON),
		config.OptWithSynonyms(false),
		config.OptWithRebuild(true),
		config.OptSortDesc(true),
		config.OptWithShortenedOutput(true),
	}
	return config.New(opts...)
}

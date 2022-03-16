package config_test

import (
	"path/filepath"
	"testing"

	"github.com/gnames/bhlnames/config"
	"github.com/gnames/gnfmt"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	test := config.Config{
		BHLDumpURL:   "https://www.biodiversitylibrary.org/data/data.zip",
		BHLIndexURL:  "bhlrpc.globalnames.org:80",
		InputDir:     config.InputDir(),
		DbHost:       "localhost",
		DbUser:       "postgres",
		DbPass:       "",
		DbDatabase:   "bhlnames",
		Delimiter:    ',',
		JobsNum:      4,
		PortREST:     8888,
		Format:       gnfmt.CSV,
		WithSynonyms: true,
	}
	test.DownloadFile = filepath.Join(test.InputDir, "data.zip")
	test.DownloadDir = filepath.Join(test.InputDir, "Data")
	test.PageDir = filepath.Join(test.InputDir, "page")
	test.PartDir = filepath.Join(test.InputDir, "part")
	test.AhoCorasickDir = filepath.Join(test.InputDir, "ac")
	test.AhoCorKeyValDir = filepath.Join(test.InputDir, "ackv")

	cfg := config.New()
	assert.Equal(t, cfg, test)
}

func TestModifiedConfig(t *testing.T) {
	test := config.Config{
		BHLDumpURL:          "https://example.org",
		BHLIndexURL:         "https://example.org",
		InputDir:            "/tmp",
		DbHost:              "10.0.0.10",
		DbUser:              "john",
		DbPass:              "doe",
		Delimiter:           '\t',
		DbDatabase:          "bhl",
		JobsNum:             100,
		PortREST:            80,
		Format:              gnfmt.CompactJSON,
		WithSynonyms:        false,
		WithRebuild:         true,
		SortDesc:            true,
		WithShortenedOutput: true,
	}
	test.DownloadFile = filepath.Join(test.InputDir, "data.zip")
	test.DownloadDir = filepath.Join(test.InputDir, "Data")
	test.PageDir = filepath.Join(test.InputDir, "page")
	test.PartDir = filepath.Join(test.InputDir, "part")
	test.AhoCorasickDir = filepath.Join(test.InputDir, "ac")
	test.AhoCorKeyValDir = filepath.Join(test.InputDir, "ackv")
	cfg := modConfig()
	assert.Equal(t, cfg, test)
}

func modConfig() config.Config {
	opts := []config.Option{
		config.OptBHLDumpURL("https://example.org"),
		config.OptBHLIndexHost("https://example.org"),
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

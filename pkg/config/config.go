package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gnames/gnsys"
)

// Config contains data needed for BHLnames functionality.
type Config struct {
	// BHLDumpURL contains the URL containing Biodiversity Heritage Library
	// dump files. These files provide metadata necessary for connection of
	// names occurrences with BHL pages.
	BHLDumpURL string

	// BHLNamesURL provides URL to BHLindex Data. This data provides names
	// occurrences and verifications. Together with data from BHL dumps it allows
	// to connect a name to pages in BHL.
	BHLNamesURL string

	// CoLDataURL provides a URL to the Catalogue of Life data in Darwin Core
	// format.
	CoLDataURL string

	// DbDatabase is the name of the database to keep BHLnames data. By default
	// it is `bhlnames`.
	DbDatabase string

	// DbHost provides an IP or host name where PostgreSQL is located. The
	// database is used as the major data store for the project.
	DbHost string

	// DbPass is the password for DBUser.
	DbPass string

	// DbUser is the username in PostgreSQL database. The user must have
	// writing permissions to the database.
	DbUser string

	// DownloadBHLFile provides the path where BHL dump compressed file will be
	// stored.
	DownloadBHLFile string

	// DownloadCoLFile provides the path where CoL DwCA compressed file will be
	// stored.
	DownloadCoLFile string

	// DownloadDir is the directory where  BHLnames extracts data from
	// BHL dump.
	DownloadDir string

	// DownloadNamesFile provides the path where BHL dump compressed file will be
	// stored.
	DownloadNamesFile string

	// InputDir provides the `root` directory where all the BHLnames files are
	// created.
	InputDir string

	// JobsNum provides concurrency value for finding references that contain
	// specified names.
	JobsNum int

	// PortREST is used for BHLnames RESTful service port.
	PortREST int

	// WithCoLRecalc indicates that calculation of CoL nomenclatural events
	// tables will be emptied, and CoL nomenclatural data will be reimported
	// before linking to BHL data.
	WithCoLRecalc bool

	// WithRebuild determines if BHL dump data need to be uploaded again, or
	// the data from local cache can be used. If `true` then local cache is
	// ignored and data is downloaded from BHLDumpURL.
	WithRebuild bool
}

// Option type for changing GNfinder settings.
type Option func(*Config)

func OptBHLDumpURL(s string) Option {
	return func(cfg *Config) {
		cfg.BHLDumpURL = s
	}
}

func OptBHLNamesURL(s string) Option {
	return func(cfg *Config) {
		cfg.BHLNamesURL = s
	}
}

func OptCoLDataURL(s string) Option {
	return func(cfg *Config) {
		cfg.CoLDataURL = s
	}
}

func OptDbHost(s string) Option {
	return func(cfg *Config) {
		cfg.DbHost = s
	}
}

func OptDbName(s string) Option {
	return func(cfg *Config) {
		cfg.DbDatabase = s
	}
}

func OptDbPass(s string) Option {
	return func(cfg *Config) {
		cfg.DbPass = s
	}
}

func OptDbUser(s string) Option {
	return func(cfg *Config) {
		cfg.DbUser = s
	}
}

func OptInputDir(s string) Option {
	return func(cfg *Config) {
		var err error
		s, err = gnsys.ConvertTilda(s)
		if err != nil {
			err = fmt.Errorf("config.OptInputDir: %#w", err)
			slog.Error("Cannot convert tilda to path.", "error", err)
			os.Exit(1)
		}
		cfg.InputDir = s
	}
}

func OptJobsNum(i int) Option {
	return func(cfg *Config) {
		cfg.JobsNum = i
	}
}

func OptPortREST(i int) Option {
	return func(cfg *Config) {
		if i > 0 {
			cfg.PortREST = i
		}
	}
}

func OptWithCoLRecalc(b bool) Option {
	return func(cfg *Config) {
		cfg.WithCoLRecalc = b
	}
}

func OptWithRebuild(b bool) Option {
	return func(cfg *Config) {
		cfg.WithRebuild = b
	}
}

func InputDir() string {
	inputDir, err := os.UserCacheDir()
	if err != nil {
		inputDir = os.TempDir()
	}
	return filepath.Join(inputDir, "bhlnames")
}

func New(opts ...Option) Config {
	cfg := Config{
		BHLDumpURL:  "http://opendata.globalnames.org/dumps/bhl-data.zip",
		BHLNamesURL: "http://opendata.globalnames.org/dumps/bhl-col.zip",
		CoLDataURL:  "https://api.checklistbank.org/dataset/3LR/export?format=dwca",
		InputDir:    InputDir(),
		DbHost:      "0.0.0.0",
		DbUser:      "postgres",
		DbPass:      "postgres",
		DbDatabase:  "bhlnames",
		JobsNum:     4,
		PortREST:    8888,
		WithRebuild: false,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	// if we redownload CoL files, we always reimport data.
	if cfg.WithRebuild {
		cfg.WithCoLRecalc = true
	}

	cfg.DownloadBHLFile = filepath.Join(cfg.InputDir, "bhl-data.zip")
	cfg.DownloadNamesFile = filepath.Join(cfg.InputDir, "bhlindex-latest.zip")
	cfg.DownloadCoLFile = filepath.Join(cfg.InputDir, "col.zip")
	cfg.DownloadDir = filepath.Join(cfg.InputDir, "Data")
	return cfg
}

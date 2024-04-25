package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gnames/gnsys"
)

// Config defines the essential parameters needed for BHLnames functionality.
type Config struct {

	// BHLDumpURL specifies the source for Biodiversity Heritage Library dump
	// files.
	BHLDumpURL string

	// BHLNamesURL specifies the source for BHLindex data (name occurrences and
	// verifications).
	BHLNamesURL string

	// CoLDataURL specifies the source for Catalogue of Life data in Darwin Core
	// Archive format.
	CoLDataURL string

	// DbDatabase is the name of the PostgreSQL database for BHLnames data.
	DbDatabase string

	// DbHost is the hostname or IP address of the PostgreSQL server.
	DbHost string

	// DbUser is the PostgreSQL username with write permissions to the database.
	DbUser string

	// DbPass is the password for `DBUser`.
	DbPass string

	// JobsNum controls the concurrency level for finding references
	// containing specified names.
	JobsNum int

	// PortREST specifies the port number for the BHLnames RESTful service.
	PortREST int

	// RootDir is the base directory for all BHLnames downloaded and extracted
	// files.
	RootDir string

	// DownloadBHLFile is the full path where the downloaded BHL dump
	// (compressed) is stored.
	DownloadBHLFile string

	// DownloadCoLFile is the full path where the downloaded CoL DwCA file is
	// stored.
	DownloadCoLFile string

	// ExtractDir is the directory where BHLnames extracts the contents of the
	// compressed files.
	ExtractDir string

	// DownloadNamesFile is the full path where the downloaded BHLindex Data file
	// is stored.
	DownloadNamesFile string

	// WithCoLDataTrim indicates that calculation of CoL nomenclatural events
	// tables will be emptied, and CoL nomenclatural data will be reimported
	// before linking to BHL data.

	// WithCoLDataTrim indicates whether the CoL nomenclatural tables should be
	// cleared and repopulated with fresh CoL data or import will continue from
	// where it was paused.
	WithCoLDataTrim bool

	// WithRebuild determines if BHL or CoL data needs to be re-downloaded and
	// processed. If true, deletes any locally cached data.
	WithRebuild bool
}

// Option enables a functional approach for modifying Config settings.
type Option func(*Config)

// OptBHLDumpURL sets the URL for BHL dump files.
func OptBHLDumpURL(s string) Option {
	return func(cfg *Config) {
		cfg.BHLDumpURL = s
	}
}

// OptBHLNamesURL sets the URL for BHLindex data.
func OptBHLNamesURL(s string) Option {
	return func(cfg *Config) {
		cfg.BHLNamesURL = s
	}
}

// OptCoLDataURL sets the URL for the Catalogue of Life data.
func OptCoLDataURL(s string) Option {
	return func(cfg *Config) {
		cfg.CoLDataURL = s
	}
}

// OptDbDatabase sets the name of the PostgreSQL database for BHLnames data.
func OptDbDatabase(s string) Option {
	return func(cfg *Config) {
		cfg.DbDatabase = s
	}
}

// OptDbHost sets the hostname or IP address of the PostgreSQL server.
func OptDbHost(s string) Option {
	return func(cfg *Config) {
		cfg.DbHost = s
	}
}

// OptDbUser sets the PostgreSQL username with write permissions to the
// database.
func OptDbUser(s string) Option {
	return func(cfg *Config) {
		cfg.DbUser = s
	}
}

// OptDbPass sets the password for the PostgreSQL user.
func OptDbPass(s string) Option {
	return func(cfg *Config) {
		cfg.DbPass = s
	}
}

// OptJobsNum sets the concurrency level for finding references containing
func OptJobsNum(i int) Option {
	return func(cfg *Config) {
		cfg.JobsNum = i
	}
}

// OptPortREST sets the port number for the BHLnames RESTful service.
func OptPortREST(i int) Option {
	return func(cfg *Config) {
		if i > 0 {
			cfg.PortREST = i
		}
	}
}

// OptRootDir sets the base directory for all BHLnames downloaded and
// extracted files.
func OptRootDir(s string) Option {
	return func(cfg *Config) {
		var err error
		s, err = gnsys.ConvertTilda(s)
		if err != nil {
			err = fmt.Errorf("config.OptInputDir: %#w", err)
			slog.Error("Cannot convert tilda to path.", "error", err)
			os.Exit(1)
		}
		cfg.RootDir = s
	}
}

// OptWithCoLDataTrim sets the CoL data trim option.
func OptWithCoLDataTrim(b bool) Option {
	return func(cfg *Config) {
		cfg.WithCoLDataTrim = b
	}
}

// OptWithRebuild sets the rebuild option.
func OptWithRebuild(b bool) Option {
	return func(cfg *Config) {
		cfg.WithRebuild = b
	}
}

// RootDir returns the default root directory for BHLnames data.
func RootDir() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}
	return filepath.Join(cacheDir, "bhlnames")
}

// New creates a new Config instance with default values.
func New(opts ...Option) Config {
	cfg := Config{
		BHLDumpURL:      "http://opendata.globalnames.org/bhlnames/bhl-data.zip",
		BHLNamesURL:     "http://opendata.globalnames.org/bhlnames/names.zip",
		CoLDataURL:      "http://opendata.globalnames.org/bhlnames/col.zip",
		DbDatabase:      "bhlnames",
		DbHost:          "0.0.0.0",
		DbUser:          "postgres",
		DbPass:          "postgres",
		JobsNum:         4,
		PortREST:        8888,
		RootDir:         RootDir(),
		WithRebuild:     false,
		WithCoLDataTrim: false,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	// if we redownload CoL files, we always reimport data.
	if cfg.WithRebuild {
		cfg.WithCoLDataTrim = true
	}

	cfg.DownloadBHLFile = filepath.Join(cfg.RootDir, "bhl-data.zip")
	cfg.DownloadNamesFile = filepath.Join(cfg.RootDir, "bhlindex-latest.zip")
	cfg.DownloadCoLFile = filepath.Join(cfg.RootDir, "col.zip")
	cfg.ExtractDir = filepath.Join(cfg.RootDir, "Data")
	return cfg
}

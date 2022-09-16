package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnames/gnfmt"
	"github.com/gnames/gnsys"
	"github.com/rs/zerolog/log"
)

// Config contains data needed for BHLnames functionality.
type Config struct {
	// BHLDumpURL contains the URL containing Biodiversity Heritage Library
	// dump files. These files provide metadata necessary for connection of
	// names occurrences with BHL pages.
	BHLDumpURL string

	// BHLNamesURL provides URL to BHLindex RESTful API. This API provides
	// names occurrences data. Together with data from BHL dumps it allows
	// to connect a name to pages in BHL.
	BHLNamesURL string

	// DbHost provides an IP or host name where PostgreSQL is located. The
	// database is used as the major data store for the project.
	DbHost string

	// DbUser is the username in PostgreSQL database. The user must have
	// writing permissions to the database.
	DbUser string

	// DbPass is the password for DBUser.
	DbPass string

	// DbDatabase is the name of the database to keep BHLnames data. By default
	// it is `bhlnames`.
	DbDatabase string

	// JobsNum provides concurrency value for finding references that contain
	// specified names.
	JobsNum int

	// PortREST is used for BHLnames RESTful service port.
	PortREST int

	// Format determines format of the output data.
	Format gnfmt.Format

	// Delimiter allows to set a delimiter for ingesting input CSV files. These
	// files contain names and other metadata to use for matching names,
	// citations to BHL pages.
	Delimiter rune

	// WithSynonyms determines if to provide synonyms of a name in the output.
	WithSynonyms bool

	// WithRebuild determines if BHL dump data need to be uploaded again, or
	// the data from local cache can be used. If `true` then local cache is
	// ignored and data is downloaded from BHLDumpURL.
	WithRebuild bool

	// SortDesc determines the order of sorting the output data. If `true`
	// data are sorted by year from latest to earliest. If `false` then from
	// earliest to latest.
	SortDesc bool

	// WithShortenedOutput determines if references details will be provided.
	// If it is `true`, found references are not provided, only the metadata
	// about them.
	WithShortenedOutput bool

	// InputDir provides the `root` directory where all the BHLnames files are
	// created.
	InputDir string

	// DownloadBHLFile provides the path where BHL dump compressed file will be
	// stored.
	DownloadBHLFile string

	// DownloadNamesFile provides the path where BHL dump compressed file will be
	// stored.
	DownloadNamesFile string

	// DownloadDir is the directory where  BHLnames extracts data from
	// BHL dump.
	DownloadDir string

	// PageDir provides the directory where BHLnames keeps key-value database for
	// pages information. We do not have file name of a page connected to page ID
	// in the BHL data dump. So we have to calculate this ID by using page
	// sequence in a title. We find out page id by concatenation of
	// "FileNum|TitleID" fields.
	//
	// This key-value store is generated using data dump from BHL databse.
	PageDir string

	// PageFileDir provides the directory to a key-value store database that
	// connects BHL's PageID to the page's file name in the BHL corpus
	// directory structure.
	//
	// It is generated using bhlindex page dump and key-value store from
	// PageDir
	PageFileDir string

	// PartDir is another key-value database to keep data about BHL's `parts`.
	// A `part` is usually a distinct entity in `item`, for example it can be
	// an scientific paper.
	PartDir string

	// AhoCorasickDir provides a directory where Aho-Corasick algorithm stores
	// its cached data.
	AhoCorasickDir string

	// AhoCorKeyValDir provides a directory to keep a Key-Value store used by
	// AhoCorasic library.
	AhoCorKeyValDir string
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

func OptDelimiter(r rune) Option {
	return func(cfg *Config) {
		cfg.Delimiter = r
	}
}

func OptInputDir(s string) Option {
	return func(cfg *Config) {
		var err error
		s, err = gnsys.ConvertTilda(s)
		if err != nil {
			err = fmt.Errorf("config.OptInputDir: %#w", err)
			log.Fatal().Err(err)
		}
		cfg.InputDir = s
	}
}

func OptDbHost(s string) Option {
	return func(cfg *Config) {
		cfg.DbHost = s
	}
}

func OptDbUser(s string) Option {
	return func(cfg *Config) {
		cfg.DbUser = s
	}
}

func OptDbPass(s string) Option {
	return func(cfg *Config) {
		cfg.DbPass = s
	}
}

func OptDbName(s string) Option {
	return func(cfg *Config) {
		cfg.DbDatabase = s
	}
}

func OptFormat(f gnfmt.Format) Option {
	return func(cfg *Config) {
		cfg.Format = f
	}
}

func OptWithRebuild(b bool) Option {
	return func(cfg *Config) {
		cfg.WithRebuild = b
	}
}

func OptJobsNum(i int) Option {
	return func(cfg *Config) {
		cfg.JobsNum = i
	}
}

func OptSortDesc(b bool) Option {
	return func(cfg *Config) {
		cfg.SortDesc = b
	}
}

func OptShort(b bool) Option {
	return func(cfg *Config) {
		cfg.WithShortenedOutput = b
	}
}

func OptWithSynonyms(b bool) Option {
	return func(cfg *Config) {
		cfg.WithSynonyms = b
	}
}

func OptPortREST(i int) Option {
	return func(cfg *Config) {
		if i > 0 {
			cfg.PortREST = i
		}
	}
}

func OptWithShortenedOutput(b bool) Option {
	return func(cfg *Config) {
		cfg.WithShortenedOutput = b
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
		BHLDumpURL:          "http://opendata.globalnames.org/dumps/bhl-data.zip",
		BHLNamesURL:         "http://opendata.globalnames.org/dumps/bhl-col.zip",
		InputDir:            InputDir(),
		Delimiter:           ',',
		DbHost:              "0.0.0.0",
		DbUser:              "postgres",
		DbPass:              "postgres",
		DbDatabase:          "bhlnames",
		JobsNum:             4,
		PortREST:            8888,
		Format:              gnfmt.CSV,
		WithSynonyms:        true,
		WithRebuild:         false,
		SortDesc:            false,
		WithShortenedOutput: false,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	cfg.DownloadBHLFile = filepath.Join(cfg.InputDir, "bhl-data.zip")
	cfg.DownloadNamesFile = filepath.Join(cfg.InputDir, "bhl-names.zip")
	cfg.DownloadDir = filepath.Join(cfg.InputDir, "Data")
	cfg.PageDir = filepath.Join(cfg.InputDir, "page")
	cfg.PageFileDir = filepath.Join(cfg.InputDir, "page-file")
	cfg.PartDir = filepath.Join(cfg.InputDir, "part")
	cfg.AhoCorasickDir = filepath.Join(cfg.InputDir, "ac")
	cfg.AhoCorKeyValDir = filepath.Join(cfg.InputDir, "ackv")
	return cfg
}

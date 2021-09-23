package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gnames/gnfmt"
	"github.com/gnames/gnsys"
)

type Config struct {
	BHLDumpURL          string
	BHLIndexHost        string
	InputDir            string
	DbHost              string
	DbUser              string
	DbPass              string
	DbName              string
	JobsNum             int
	PortREST            int
	Format              gnfmt.Format
	WithSynonyms        bool
	WithRebuild         bool
	SortDesc            bool
	WithShortenedOutput bool
	DownloadFile        string
	DownloadDir         string
	KeyValDir           string
	PartDir             string
}

// Option type for changing GNfinder settings.
type Option func(*Config)

func OptBHLDumpURL(s string) Option {
	return func(cnf *Config) {
		cnf.BHLDumpURL = s
	}
}

func OptBHLIndexHost(s string) Option {
	return func(cnf *Config) {
		cnf.BHLIndexHost = s
	}
}

func OptInputDir(s string) Option {
	return func(cnf *Config) {
		var err error
		s, err = gnsys.ConvertTilda(s)
		if err != nil {
			log.Fatal(err)
		}
		cnf.InputDir = s
	}
}

func OptDbHost(s string) Option {
	return func(cnf *Config) {
		cnf.DbHost = s
	}
}

func OptDbUser(s string) Option {
	return func(cnf *Config) {
		cnf.DbUser = s
	}
}

func OptDbPass(s string) Option {
	return func(cnf *Config) {
		cnf.DbPass = s
	}
}

func OptDbName(s string) Option {
	return func(cnf *Config) {
		cnf.DbName = s
	}
}

func OptFormat(f gnfmt.Format) Option {
	return func(cnf *Config) {
		cnf.Format = f
	}
}

func OptWithRebuild(b bool) Option {
	return func(cnf *Config) {
		cnf.WithRebuild = b
	}
}

func OptJobsNum(i int) Option {
	return func(cnf *Config) {
		cnf.JobsNum = i
	}
}

func OptSortDesc(b bool) Option {
	return func(cnf *Config) {
		cnf.SortDesc = b
	}
}

func OptShort(b bool) Option {
	return func(cnf *Config) {
		cnf.WithShortenedOutput = b
	}
}

func OptWithSynonyms(b bool) Option {
	return func(cnf *Config) {
		cnf.WithSynonyms = b
	}
}

func OptPortREST(i int) Option {
	return func(cnf *Config) {
		if i > 0 {
			cnf.PortREST = i
		}
	}
}

func OptWithShortenedOutput(b bool) Option {
	return func(cnf *Config) {
		cnf.WithShortenedOutput = b
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
		BHLDumpURL:          "https://www.biodiversitylibrary.org/data/data.zip",
		BHLIndexHost:        "bhlrpc.globalnames.org:80",
		InputDir:            InputDir(),
		DbHost:              "localhost",
		DbUser:              "postgres",
		DbPass:              "",
		DbName:              "bhlnames",
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

	cfg.DownloadFile = filepath.Join(cfg.InputDir, "data.zip")
	cfg.DownloadDir = filepath.Join(cfg.InputDir, "Data")
	cfg.KeyValDir = filepath.Join(cfg.InputDir, "keyval")
	cfg.PartDir = filepath.Join(cfg.InputDir, "part")
	return cfg
}

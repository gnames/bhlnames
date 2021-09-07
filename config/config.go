package config

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/gnames/gnfmt"
	homedir "github.com/mitchellh/go-homedir"
)

type Config struct {
	Output
	Search
	DB
	BHL
	FileSystem
	Performance
	REST
}

type Performance struct {
	JobsNum int
}

type REST struct {
	Port int
}

type Search struct {
	NoSynonyms bool `json:"noSynonyms"`
}

type Output struct {
	Format       gnfmt.Format `json:"-"`
	FormatString string       `json:"format"`
	SortDesc     bool         `json:"sortDescending"`
	Short        bool         `json:"shortOutput"`
}

type DB struct {
	Host string
	User string
	Pass string
	Name string
}

type FileSystem struct {
	InputDir     string
	DownloadFile string
	DownloadDir  string
	KeyValDir    string
	PartDir      string
}

type BHL struct {
	DumpURL      string
	BHLindexHost string
	Rebuild      bool
}

// Option type for changing GNfinder settings.
type Option func(*Config)

func OptDumpURL(d string) Option {
	return func(cnf *Config) {
		cnf.DumpURL = d
	}
}

func OptBHLindexHost(bh string) Option {
	return func(cnf *Config) {
		cnf.BHLindexHost = bh
	}
}

func OptInputDir(s string) Option {
	return func(cnf *Config) {
		if strings.HasPrefix(s, "~/") || strings.HasPrefix(s, "~\\") {
			home, err := homedir.Dir()
			if err != nil {
				log.Fatal(err)
			}
			s = filepath.Join(home, s[2:])
		}
		cnf.InputDir = s
	}
}

func OptDbHost(h string) Option {
	return func(cnf *Config) {
		cnf.Host = h
	}
}

func OptDbUser(u string) Option {
	return func(cnf *Config) {
		cnf.User = u
	}
}

func OptDbPass(p string) Option {
	return func(cnf *Config) {
		cnf.Pass = p
	}
}

func OptDbName(n string) Option {
	return func(cnf *Config) {
		cnf.Name = n
	}
}

func OptFormat(s string) Option {
	return func(cnf *Config) {
		f, err := gnfmt.NewFormat(s)
		if err != nil {
			log.Println(err)
			f = gnfmt.CSV
		}
		cnf.Format = f
		cnf.FormatString = f.String()
	}
}

func OptRebuild(r bool) Option {
	return func(cnf *Config) {
		cnf.Rebuild = r
	}
}

func OptJobsNum(j int) Option {
	return func(cnf *Config) {
		cnf.JobsNum = j
	}
}

func OptSortDesc(d bool) Option {
	return func(cnf *Config) {
		cnf.SortDesc = d
	}
}

func OptShort(s bool) Option {
	return func(cnf *Config) {
		cnf.Short = s
	}
}

func OptNoSynonyms(n bool) Option {
	return func(cnf *Config) {
		cnf.NoSynonyms = n
	}
}

func OptPortREST(i int) Option {
	return func(cnf *Config) {
		if i > 0 {
			cnf.Port = i
		}
	}
}

func NewConfig(opts ...Option) Config {
	cfg := Config{REST: REST{Port: 8888}}
	for _, opt := range opts {
		opt(&cfg)
	}
	cfg.DownloadFile = filepath.Join(cfg.InputDir, "data.zip")
	cfg.DownloadDir = filepath.Join(cfg.InputDir, "Data")
	cfg.KeyValDir = filepath.Join(cfg.InputDir, "keyval")
	cfg.PartDir = filepath.Join(cfg.InputDir, "part")
	return cfg
}

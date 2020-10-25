package config

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/gnames/gnlib/format"
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
	Format       format.Format `json:"-"`
	FormatString string        `json:"format"`
	SortDesc     bool          `json:"sortDescending"`
	Short        bool          `json:"shortOutput"`
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
		cnf.BHL.DumpURL = d
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
		cnf.FileSystem.InputDir = s
	}
}

func OptDbHost(h string) Option {
	return func(cnf *Config) {
		cnf.DB.Host = h
	}
}

func OptDbUser(u string) Option {
	return func(cnf *Config) {
		cnf.DB.User = u
	}
}

func OptDbPass(p string) Option {
	return func(cnf *Config) {
		cnf.DB.Pass = p
	}
}

func OptDbName(n string) Option {
	return func(cnf *Config) {
		cnf.DB.Name = n
	}
}

func OptFormat(s string) Option {
	return func(cnf *Config) {
		f, err := format.NewFormat(s)
		if err != nil {
			log.Println(err)
			f = format.CSV
		}
		cnf.Format = f
		cnf.FormatString = f.String()
	}
}

func OptRebuild(r bool) Option {
	return func(cnf *Config) {
		cnf.BHL.Rebuild = r
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
			cnf.REST.Port = i
		}
	}
}

func NewConfig(opts ...Option) Config {
	cfg := Config{REST: REST{Port: 8888}}
	for _, opt := range opts {
		opt(&cfg)
	}
	cfg.FileSystem.DownloadFile = filepath.Join(cfg.InputDir, "data.zip")
	cfg.FileSystem.DownloadDir = filepath.Join(cfg.InputDir, "Data")
	cfg.FileSystem.KeyValDir = filepath.Join(cfg.InputDir, "keyval")
	cfg.FileSystem.PartDir = filepath.Join(cfg.InputDir, "part")
	return cfg
}

/*
Copyright © 2020-2021 Dmitry Mozzherin <dmozzherin@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/package cmd

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/gnsys"
	"github.com/spf13/viper"
)

//go:embed bhlnames.yaml
var configText string

var (
	opts []config.Option
)

// fConfig purpose is to achieve automatic import of data from the
// configuration file, if it exists.
type fConfig struct {
	BHLDumpURL  string
	BHLNamesURL string
	CoLDataURL  string
	DbDatabase  string
	DbHost      string
	DbUser      string
	DbPass      string
	JobsNum     int
	PortREST    int
	RootDir     string
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bhlnames",
	Short: "BHLnames finds publications for scientific names in BHL",
	Long: `BHLnames uses Biodiversity Heritage Library (BHL) and its Name Index
data  to build a local database. It uses this database to return all
usages found at BHL for a scientific name.`,
	Run: func(cmd *cobra.Command, args []string) {
		versionFlag(cmd)

		if len(args) == 0 {
			_ = cmd.Help()
			os.Exit(0)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("Bootstrap command failed.", "error", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().BoolP("version", "V", false, "Returns version and build date")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	configFile := "bhlnames"
	configDir, err := os.UserConfigDir()
	if err != nil {
		slog.Error("Cannot find user's config directory.", "error", err)
		os.Exit(1)
	}
	viper.AddConfigPath(configDir)
	viper.SetConfigName(configFile)

	viper.BindEnv("BHLDumpURL", "BHL_NAMES_DUMP_URL")
	viper.BindEnv("BHLNamesURL", "BHL_NAMES_URL")
	viper.BindEnv("ColDataURL", "BHL_NAMES_COL_DATA_URL")
	viper.BindEnv("DbDatabase", "BHL_NAMES_DB_DATABASE")
	viper.BindEnv("DbHost", "BHL_NAMES_DB_HOST")
	viper.BindEnv("DbUser", "BHL_NAMES_DB_USER")
	viper.BindEnv("DbPass", "BHL_NAMES_DB_PASS")
	viper.BindEnv("JobsNum", "BHL_NAMES_JOBS_NUM")
	viper.BindEnv("PortREST", "BHL_NAMES_PORT_REST")
	viper.BindEnv("RootDir", "BHL_NAMES_ROOT_DIR")
	viper.AutomaticEnv()

	configPath := filepath.Join(configDir, fmt.Sprintf("%s.yaml", configFile))
	touchConfigFile(configPath)

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		slog.Info("Using config file.", "file", viper.ConfigFileUsed())
	}

	opts = getOpts()
}

// getOpts imports data from the configuration file. These settings can be
// overriden by command line flags.
func getOpts() []config.Option {
	var opts []config.Option
	cfg := &fConfig{}
	err := viper.Unmarshal(cfg)
	if err != nil {
		slog.Error("Cannot unmarshal config.", "error", err)
		os.Exit(1)
	}

	if cfg.BHLDumpURL != "" {
		opts = append(opts, config.OptBHLDumpURL(cfg.BHLDumpURL))
	}
	if cfg.BHLNamesURL != "" {
		opts = append(opts, config.OptBHLNamesURL(cfg.BHLNamesURL))
	}
	if cfg.CoLDataURL != "" {
		opts = append(opts, config.OptCoLDataURL(cfg.CoLDataURL))
	}
	if cfg.DbDatabase != "" {
		opts = append(opts, config.OptDbDatabase(cfg.DbDatabase))
	}
	if cfg.DbHost != "" {
		opts = append(opts, config.OptDbHost(cfg.DbHost))
	}
	if cfg.DbUser != "" {
		opts = append(opts, config.OptDbUser(cfg.DbUser))
	}
	if cfg.DbPass != "" {
		opts = append(opts, config.OptDbPass(cfg.DbPass))
	}
	if cfg.JobsNum != 0 {
		opts = append(opts, config.OptJobsNum(cfg.JobsNum))
	}
	if cfg.PortREST != 0 {
		opts = append(opts, config.OptPortREST(cfg.PortREST))
	}
	if cfg.RootDir != "" {
		opts = append(opts, config.OptRootDir(cfg.RootDir))
	}
	return opts
}

// touchConfigFile checks if config file exists, and if not, it gets created.
func touchConfigFile(configPath string) {
	exists, _ := gnsys.FileExists(configPath)
	if exists {
		return
	}

	slog.Info("Creating config file.", "file", configPath)
	createConfig(configPath)
}

// createConfig creates config file.
func createConfig(path string) {
	err := gnsys.MakeDir(filepath.Dir(path))
	if err != nil {
		slog.Error("Cannot create config dir.", "error", err)
		os.Exit(1)
	}

	err = os.WriteFile(path, []byte(configText), 0644)
	if err != nil {
		slog.Error("Cannot create config file.", "error", err)
		os.Exit(1)
	}
}

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
*/
package cmd

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/gnames/bhlnames"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/gnsys"
	"github.com/spf13/viper"
)

//go:embed bhlnames.yaml
var configText string

var (
	cfgFile string
	opts    []config.Option
)

// fConfig purpose is to achieve automatic import of data from the
// configuration file, if it exists.
type fConfig struct {
	BHLDumpURL  string
	BHLIndexURL string
	InputDir    string
	DbHost      string
	DbUser      string
	DbPass      string
	DbName      string
	JobsNum     int
	PortREST    int
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bhlnames",
	Short: "Finds publications for scientific names in BHL",
	Long: `Uses bhlindex service and Biodiversity Heritage Library (BHL)
metadata to build a local database. It Uses this database to return all
usages found at BHL for a scientific name.`,
	Run: func(cmd *cobra.Command, args []string) {
		version, err := cmd.Flags().GetBool("version")
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		if version {
			fmt.Printf("\nversion: %s\nbuild: %s\n\n", bhlnames.Version, bhlnames.Build)
			os.Exit(0)
		}

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
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/bhlnames.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("version", "V", false, "Returns version and build date")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	configFile := "bhlnames"
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Cannot find home directory: %s.", err)
	}
	viper.AddConfigPath(configDir)
	viper.SetConfigName(configFile)

	viper.BindEnv("BHLDumpURL", "BHL_NAMES_DUMP_URL")
	viper.BindEnv("BHLindexHost", "BHL_NAMES_INDEX_HOST")
	viper.BindEnv("InputDir", "BHL_NAMES_INPUT_DIR")
	viper.BindEnv("DbHost", "BHL_NAMES_DB_HOST")
	viper.BindEnv("DbPort", "BHL_NAMES_DB_PORT")
	viper.BindEnv("DbUser", "BHL_NAMES_DB_USER")
	viper.BindEnv("DbPass", "BHL_NAMES_DB_PASS")
	viper.BindEnv("DbName", "BHL_NAMES_DATABASE")
	viper.BindEnv("JobsNum", "BHL_NAMES_JOBS_NUM")
	viper.BindEnv("PortREST", "BHL_NAMES_PORT_REST")
	viper.AutomaticEnv() // read in environment variables that match

	configPath := filepath.Join(configDir, fmt.Sprintf("%s.yaml", configFile))
	touchConfigFile(configPath, configFile)

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Printf("Using config file: %s.", viper.ConfigFileUsed())
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
		log.Fatal(err)
	}

	if cfg.BHLDumpURL != "" {
		opts = append(opts, config.OptBHLDumpURL(cfg.BHLDumpURL))
	}
	if cfg.BHLIndexURL != "" {
		opts = append(opts, config.OptBHLIndexURL(cfg.BHLIndexURL))
	}
	if cfg.InputDir != "" {
		opts = append(opts, config.OptInputDir(cfg.InputDir))
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
	if cfg.DbName != "" {
		opts = append(opts, config.OptDbName(cfg.DbName))
	}
	if cfg.JobsNum != 0 {
		opts = append(opts, config.OptJobsNum(cfg.JobsNum))
	}
	if cfg.PortREST != 0 {
		opts = append(opts, config.OptPortREST(cfg.PortREST))
	}
	return opts
}

// touchConfigFile checks if config file exists, and if not, it gets created.
func touchConfigFile(configPath string, configFile string) {
	exists, _ := gnsys.FileExists(configPath)
	if exists {
		return
	}

	log.Printf("Creating config file: %s.", configPath)
	createConfig(configPath, configFile)
}

// createConfig creates config file.
func createConfig(path string, file string) {
	err := gnsys.MakeDir(filepath.Dir(path))
	if err != nil {
		log.Fatalf("Cannot create dir %s: %s.", path, err)
	}

	err = ioutil.WriteFile(path, []byte(configText), 0644)
	if err != nil {
		log.Fatalf("Cannot write to file %s: %s.", path, err)
	}
}

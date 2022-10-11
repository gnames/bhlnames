/*
Copyright © 2022 Dmitry Mozzherin <dmozzherin@gmail.com>

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
	"fmt"
	"os"

	"github.com/gnames/bhlnames"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/io/bayesio"
	"github.com/gnames/bhlnames/io/colbuildio"
	"github.com/gnames/bhlnames/io/reffinderio"
	"github.com/gnames/bhlnames/io/titlemio"
	"github.com/gnames/gnparser"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// initcolCmd represents the initcol command
var initcolCmd = &cobra.Command{
	Use:   "initcol",
	Short: "finds nomenclatural references for names from the Catalogue Of Life.",
	Long: `The subcommand uses imported from the Catalogue of Life
names and references strings. It calculates putative links to the references
in the Biodiversity Heritage Library and assigns quality, score to the finds
and saves them to database.

This command runs for several hours and is a part of initialization.`,
	Run: func(cmd *cobra.Command, _ []string) {
		rebuild, err := cmd.Flags().GetBool("restart")
		if err != nil {
			err = fmt.Errorf("initCmd: %#w", err)
			log.Fatal().Err(err).Msg("")
		}

		recalc, err := cmd.Flags().GetBool("recalc")
		if err != nil {
			err = fmt.Errorf("initCmd: %#w", err)
			log.Fatal().Err(err).Msg("")
		}

		opts = append(
			opts,
			config.OptWithRebuild(rebuild),
			config.OptWithCoLRecalc(recalc),
		)

		cfg := config.New(opts...)
		cn := colbuildio.New(cfg)
		rf := reffinderio.New(cfg)
		tm := titlemio.New(cfg)

		gnp := gnparser.New(gnparser.NewConfig())

		bnOpts := []bhlnames.Option{
			bhlnames.OptRefFinder(rf),
			bhlnames.OptParser(gnp),
			bhlnames.OptTitleMatcher(tm),
			bhlnames.OptNLP(bayesio.New()),
			bhlnames.OptColBuild(cn),
		}

		bn := bhlnames.New(cfg, bnOpts...)
		defer bn.Close()

		showWarning(bn)
		err = bn.InitializeCol()
		if err != nil {
			err = fmt.Errorf("InitializeCol: %w", err)
			log.Fatal().Err(err).Msg("")
		}
	},
}

func init() {
	rootCmd.AddCommand(initcolCmd)
	initcolCmd.Flags().BoolP("restart", "R", false, "Delete and rebuild files and tables for CoL data.")
	initcolCmd.Flags().BoolP("recalc", "r", false, "Keep downloads, rebuild and reimport tables")

}

func showWarning(bn bhlnames.BHLnames) {
	if bn.Config().WithRebuild || bn.Config().WithCoLRecalc {
		fmt.Println("Previously generated CoL data will be lost.")
		fmt.Println("All other data will not be affected.")
		fmt.Println("Do you want to proceed? (y/N)")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" {
			os.Exit(0)
		}
	}
}

/*
Copyright © 2024 Dmitry Mozzherin <dmozzherin@gmail.com>

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
	"log/slog"
	"os"

	"github.com/gnames/bhlnames/internal/io/bayesio"
	"github.com/gnames/bhlnames/internal/io/colio"
	"github.com/gnames/bhlnames/internal/io/reffndio"
	"github.com/gnames/bhlnames/internal/io/ttlmchio"
	bhlnames "github.com/gnames/bhlnames/pkg"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/spf13/cobra"
)

// colCmd represents the col command
var colCmd = &cobra.Command{
	Use:   "col",
	Short: "Discovers putative nomenclatural events for Catalogue of Life.",
	Long: `Imports names and their references from the Catalogue of Life.
Then it finds putative locations of the names' nomenclatural references
in the Biodiversity Heritage Library and saves results to the database.

This command runs for several hours and is a part of initialization. It needs
to be run only once, unless you want to update the data.`,

	Run: func(cmd *cobra.Command, args []string) {
		for _, flag := range []flagFunc{
			rebuildFlag, trimFlag,
		} {
			flag(cmd)
		}

		cfg := config.New(opts...)
		rf, err := reffndio.New(cfg)
		if err != nil {
			slog.Error("Cannot create a Reference Finder instance.", "error", err)
			os.Exit(1)
		}

		tm, err := ttlmchio.New(cfg)
		if err != nil {
			slog.Error("Cannot create a Title Matcher instance.", "error", err)
			os.Exit(1)
		}

		bnOpts := []bhlnames.Option{
			bhlnames.OptRefFinder(rf),
			bhlnames.OptTitleMatcher(tm),
			bhlnames.OptNLP(bayesio.New()),
		}

		bn := bhlnames.New(cfg, bnOpts...)
		defer bn.Close()

		showDeleteCoLDataWarning(bn)

		cn, err := colio.New(cfg)
		if err != nil {
			slog.Error("Cannot create CoL Builder.", "error", err)
			os.Exit(1)
		}

		err = bn.InitCoLNomenEvents(cn)
		if err != nil {
			slog.Error("Could not get nomen events from CoL.", "error", err)
			os.Exit(1)
		}
	},
}

func init() {
	initCmd.AddCommand(colCmd)

	colCmd.Flags().BoolP(
		"trim", "t", false,
		"trim data from CoL tables and rebuild them")
}

func showDeleteCoLDataWarning(bn bhlnames.BHLnames) {
	cfg := bn.Config()
	if cfg.WithRebuild || cfg.WithCoLDataTrim {
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

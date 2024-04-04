/*
Copyright © 2020-2024 Dmitry Mozzherin <dmozzherin@gmail.com>

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
	"log/slog"
	"os"

	"github.com/gnames/bhlnames/internal/io/builderio"
	bhlnames "github.com/gnames/bhlnames/pkg"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Creates database for bhlnames",
	Long: `Downloads database data from BHL, BHLIndex dump and uses them to build
additional data about names and their connection to references.`,
	Run: func(cmd *cobra.Command, _ []string) {
		// add rebuild option. If true, all data will be deleted and redownloaded.
		rebuildFlag(cmd)

		cfg := config.New(opts...)

		// builder is used exclusively for initialization
		builder, err := builderio.New(cfg)
		if err != nil {
			slog.Error("Cannot create builder.", "error", err)
			os.Exit(1)
		}
		defer builder.Close()

		bn := bhlnames.New(cfg)
		defer bn.Close()

		err = bn.Initialize(builder)
		if err != nil {
			slog.Error("Initialize failed.", "error", err)
			os.Exit(1)
		}
		slog.Info("Import of BHL data and names is done.")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolP("rebuild", "r", false, "Delete data and rebuild")
}

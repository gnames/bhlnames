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
	"fmt"
	"log/slog"
	"os"

	"github.com/gnames/bhlnames/internal/ent/input"
	"github.com/gnames/bhlnames/internal/io/bayesio"
	"github.com/gnames/bhlnames/internal/io/reffndio"
	"github.com/gnames/bhlnames/internal/io/ttlmchio"
	bhlnames "github.com/gnames/bhlnames/pkg"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/gnfmt"
	"github.com/spf13/cobra"
)

var inpOpts []input.Option

// namerefCmd represents the nameref command
var namerefCmd = &cobra.Command{
	Use:   "nameref",
	Short: "nameref finds references for a name-string in BHL.",
	Long: `The nameref command finds points in BHL where a name-string was
	mentioned. The list of results can be filtered by providing a information
	about a reference that should be found.

	There are additional options that allow to return not only results for a
	particular name-string, but for all synonyms of a taxon. It is also possible
	to find nomenclatural events.
	`,

	Run: func(cmd *cobra.Command, args []string) {
		for _, flag := range []flagFunc{
			jobsFlag, descFlag, shortFlag, taxonFlag, refsLimitFlag,
			nomenFlag,
		} {
			flag(cmd)
		}

		cfg := config.New(opts...)

		rf, err := reffndio.New(cfg)
		if err != nil {
			slog.Error("Cannot create reference finder", "error", err)
			os.Exit(1)
		}

		tm, err := ttlmchio.New(cfg)
		if err != nil {
			slog.Error("Cannot create title matcher", "error", err)
			os.Exit(1)
		}

		bnOpts := []bhlnames.Option{
			bhlnames.OptRefFinder(rf),
			bhlnames.OptTitleMatcher(tm),
			bhlnames.OptNLP(bayesio.New()),
		}

		bn := bhlnames.New(cfg, bnOpts...)
		defer bn.Close()

		argData := readArgs(cmd, args)
		name(bn, argData)
	},
}

func init() {
	rootCmd.AddCommand(namerefCmd)

	namerefCmd.Flags().StringP("format", "f", "compact",
		"JSON output format can be 'compact' or 'pretty.")

	namerefCmd.Flags().IntP("jobs", "j", 0,
		"Number of parallel jobs to get references.")

	namerefCmd.Flags().BoolP("sort_desc", "d", false,
		"Sort references by year in descending order.")

	namerefCmd.Flags().BoolP("short_output", "s", false,
		"Return only summary (no references data).")

	namerefCmd.Flags().BoolP("taxon", "t", false,
		"Find references for taxon (with name synonyms).")

	namerefCmd.Flags().BoolP("nomen_event", "n", false,
		"Find nomenclatural events.")

	namerefCmd.Flags().IntP("refs_limit", "l", 0,
		"Limit number of returned references")

	namerefCmd.Flags().StringP("delimiter", "D", ",",
		"Delimiter for reading CSV files, default is comma.")
}

type data struct {
	name, ref string
}

func readArgs(cmd *cobra.Command, args []string) data {
	var res data
	switch len(args) {
	case 1:
		res.name = args[0]
	case 2:
		res.name = args[0]
		res.ref = args[1]

	default:
		_ = cmd.Help()
		os.Exit(0)
	}
	return res
}

func name(bn bhlnames.BHLnames, args data) {
	inpOpts = append(inpOpts, input.OptNameString(args.name))
	if args.ref != "" {
		inpOpts = append(inpOpts, input.OptRefString(args.ref))
	}

	inp := input.New(bn.ParserPool(), inpOpts...)
	enc := gnfmt.GNjson{}
	res, err := bn.NameRefs(inp)
	if err != nil {
		slog.Error("Cannot get names with references", "error", err)
		os.Exit(1)
	}
	out := enc.Output(res, gnfmt.CompactJSON)
	fmt.Println(out)
}

// Copyright © 2020 Dmitry Mozzherin <dmozzherin@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlnames/internal/ent/input"
	"github.com/gnames/bhlnames/internal/ent/namerefs"
	"github.com/gnames/bhlnames/internal/io/reffinderio"
	"github.com/gnames/bhlnames/internal/io/titlemio"
	bhlnames "github.com/gnames/bhlnames/pkg"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/gnfmt"
	"github.com/gnames/gnparser"
	"github.com/spf13/cobra"
)

// nameCmd represents the name command
var nameCmd = &cobra.Command{
	Use:   "name",
	Short: "Finds references in BHL for name/s",
	Long: `Takes one name string or a file with scientific names and creates
a list of usages/references for the names in Biodiversity Heritage Library.`,
	Run: func(cmd *cobra.Command, args []string) {
		f := formatFlag(cmd)
		d := descFlag(cmd)
		s := shortFlag(cmd)
		n := noSynonymsFlag(cmd)
		delim := delimiterFlag(cmd)
		opts = append(opts,
			config.OptFormat(f), config.OptSortDesc(d),
			config.OptShort(s), config.OptWithSynonyms(!n),
			config.OptDelimiter(delim),
		)
		j := jobsFlag(cmd)
		if j > 0 {
			opts = append(opts, config.OptJobsNum(j))
		}
		cfg := config.New(opts...)

		rf, err := reffinderio.New(cfg)
		if err != nil {
			slog.Error("Cannot create reffinder", "error", err)
			os.Exit(1)
		}

		tm, err := titlemio.New(cfg)
		if err != nil {
			slog.Error("Cannot create title matcher", "error", err)
			os.Exit(1)
		}

		gnp := gnparser.New(gnparser.NewConfig())

		bnOpts := []bhlnames.Option{
			bhlnames.OptRefFinder(rf),
			bhlnames.OptParser(gnp),
			bhlnames.OptTitleMatcher(tm),
		}

		bn := bhlnames.New(cfg, bnOpts...)
		defer bn.Close()

		if len(args) == 0 {
			processStdin(cmd, bn)
			os.Exit(0)
		}
		data := getInput(cmd, args)
		name(bn, data)
	},
}

func init() {
	rootCmd.AddCommand(nameCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// nameCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	nameCmd.Flags().StringP("format", "f", "compact",
		"JSON output format can be 'compact' or 'pretty.")

	nameCmd.Flags().IntP("jobs", "j", 0,
		"Number of parallel jobs to get references.")

	nameCmd.Flags().BoolP("sort_desc", "D", false,
		"Sort references by year in descending order.")

	nameCmd.Flags().BoolP("short_output", "s", false,
		"Return only summary (no references data).")

	nameCmd.Flags().BoolP("no_synonyms", "n", false,
		"Do not expand name to synonyms.")

	nameCmd.Flags().StringP("delimiter", "d", ",",
		"Delimiter for reading CSV files, default is comma.")
}

func formatFlag(cmd *cobra.Command) gnfmt.Format {
	format := gnfmt.CSV
	s, _ := cmd.Flags().GetString("format")

	if s == "" {
		return format
	}
	if s != "csv" {
		fmt, _ := gnfmt.NewFormat(s)
		if fmt == gnfmt.FormatNone {
			slog.Info(
				"Cannot set format from string, setting it to csv",
				"format-string", s,
			)
			return format
		}

		format = fmt
	}
	return format
}

func jobsFlag(cmd *cobra.Command) int {
	j, err := cmd.Flags().GetInt("jobs")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return j
}

func descFlag(cmd *cobra.Command) bool {
	b, err := cmd.Flags().GetBool("sort_desc")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return b
}

func shortFlag(cmd *cobra.Command) bool {
	s, err := cmd.Flags().GetBool("short_output")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return s
}

func noSynonymsFlag(cmd *cobra.Command) bool {
	n, err := cmd.Flags().GetBool("no_synonyms")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return n
}

func processStdin(cmd *cobra.Command, bn bhlnames.BHLnames) {
	if !checkStdin() {
		_ = cmd.Help()
		return
	}
	nameFile(bn, os.Stdin)
}

func checkStdin() bool {
	stdInFile := os.Stdin
	stat, err := stdInFile.Stat()
	if err != nil {
		err = fmt.Errorf("checkStdin: %#w", err)
		slog.Error("Cannot get Stdin stat", "error", err)
		os.Exit(1)
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func getInput(cmd *cobra.Command, args []string) string {
	var data string
	switch len(args) {
	case 1:
		data = args[0]
	default:
		_ = cmd.Help()
		os.Exit(0)
	}
	return data
}

func name(bn bhlnames.BHLnames, data string) {
	path := string(data)
	if fileExists(path) {
		f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
		if err != nil {
			slog.Error("Cannot open file", "error", err, "path", path)
		}
		nameFile(bn, f)
		f.Close()
	} else {
		nameString(bn, data)
	}
}

func fileExists(path string) bool {
	if fi, err := os.Stat(path); err == nil {
		if fi.Mode().IsRegular() {
			return true
		}
	}
	return false
}

func nameFile(bn bhlnames.BHLnames, f io.Reader) {
	in := make(chan input.Input)
	out := make(chan *namerefs.NameRefs)

	var wg sync.WaitGroup
	wg.Add(1)

	go bn.NameRefsStream(in, out)
	go processResults(bn.Config().Format, out, &wg)

	r := csv.NewReader(f)
	r.Comma = bn.Config().Delimiter

	// read header
	header := make(map[string]int)
	hdr, err := r.Read()
	if err != nil {
		slog.Error("Cannot read CSV file", "error", err)
		os.Exit(1)
	}
	for i, v := range hdr {
		header[v] = i
	}
	csvVal := func(row []string, key string) string {
		if val, ok := header[key]; ok {
			return row[val]
		}
		return ""
	}

	count := 0
	slog.Info("Finding references.")
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Error("Cannot read CSV row", "error", err)
		}

		count++
		if count%1000 == 0 {
			slog.Info("Processing lines", "line", humanize.Comma(int64(count)))
		}
		opts := []input.Option{
			input.OptID(csvVal(row, "Id")),
			input.OptNameString(csvVal(row, "NameString")),
			input.OptRefString(csvVal(row, "RefString")),
		}
		input := input.New(bn.Parser(), opts...)
		in <- input
	}
	close(in)
	wg.Wait()
	slog.Info("Finished finding references.")
}

func processResults(f gnfmt.Format, chOut <-chan *namerefs.NameRefs,
	wg *sync.WaitGroup) {
	enc := gnfmt.GNjson{Pretty: true}
	defer wg.Done()
	var dump []*namerefs.NameRefs
	for nameRef := range chOut {
		dump = append(dump, nameRef)
		fmt.Println(enc.Output(nameRef, f))
	}
	encDump, err := enc.Encode(dump)
	if err != nil {
		slog.Error("Cannot encode results", "error", err)
		os.Exit(1)
	}
	err = os.WriteFile("testdata/stubs_namerefs.json", encDump, 0644)
	if err != nil {
		slog.Error("Cannot write results", "error", err)
		os.Exit(1)
	}
}

func nameString(bn bhlnames.BHLnames, name string) {
	opts := []input.Option{input.OptNameString(name)}
	data := input.New(bn.Parser(), opts...)
	enc := gnfmt.GNjson{}
	res, err := bn.NameRefs(data)
	if err != nil {
		slog.Error("Cannot get names with references", "error", err)
		os.Exit(1)
	}
	fmt.Println(enc.Output(res, bn.Config().Format))
}

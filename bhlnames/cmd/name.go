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
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlnames"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/ent/input"
	"github.com/gnames/bhlnames/ent/namerefs"
	"github.com/gnames/bhlnames/ent/reffinder"
	"github.com/gnames/bhlnames/io/reffinderio"
	"github.com/gnames/gnfmt"
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
		opts = append(opts,
			config.OptFormat(f), config.OptSortDesc(d),
			config.OptShort(s), config.OptWithSynonyms(!n),
		)
		j := jobsFlag(cmd)
		if j > 0 {
			opts = append(opts, config.OptJobsNum(j))
		}
		cfg := config.New(opts...)
		bhln := bhlnames.New(cfg)
		rf := reffinderio.New(cfg)
		defer rf.Close()
		if len(args) == 0 {
			processStdin(cmd, rf, bhln)
			os.Exit(0)
		}
		data := getInput(cmd, args)
		name(rf, bhln, data)
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

	nameCmd.Flags().BoolP("sort_desc", "d", false,
		"Sort references by year in descending order.")

	nameCmd.Flags().BoolP("short_output", "s", false,
		"Return only summary (no references data).")

	nameCmd.Flags().BoolP("no_synonyms", "n", false,
		"Do not expand name to synonyms.")
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
			log.Printf(
				"Cannot set format from '%s', setting format to csv",
				s,
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

func processStdin(cmd *cobra.Command, rf reffinder.RefFinder, bhln bhlnames.BHLnames) {
	if !checkStdin() {
		_ = cmd.Help()
		return
	}
	nameFile(rf, bhln, os.Stdin)
}

func checkStdin() bool {
	stdInFile := os.Stdin
	stat, err := stdInFile.Stat()
	if err != nil {
		log.Panic(err)
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

func name(rf reffinder.RefFinder, bhln bhlnames.BHLnames, data string) {
	path := string(data)
	if fileExists(path) {
		f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		nameFile(rf, bhln, f)
		f.Close()
	} else {
		nameString(rf, bhln, data)
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

func nameFile(rf reffinder.RefFinder, bhln bhlnames.BHLnames, f io.Reader) {
	in := make(chan input.Input)
	out := make(chan *namerefs.NameRefs)
	var wg sync.WaitGroup
	wg.Add(1)

	go bhln.NameRefsStream(rf, in, out)
	go processResults(bhln.Format(), out, &wg)
	sc := bufio.NewScanner(f)
	count := 0
	log.Println("Finding references")
	for sc.Scan() {
		count++
		if count%1000 == 0 {
			log.Printf("Processing %s-th line\n", humanize.Comma(int64(count)))
		}

		name := input.Name{NameString: sc.Text()}
		in <- input.Input{Name: name}
	}
	close(in)
	wg.Wait()
	log.Println("Finish finding references")
}

func processResults(f gnfmt.Format, chOut <-chan *namerefs.NameRefs,
	wg *sync.WaitGroup) {
	enc := gnfmt.GNjson{}
	defer wg.Done()
	for nameRef := range chOut {
		enc.Output(nameRef, f)
	}
}

func nameString(rf reffinder.RefFinder, bhln bhlnames.BHLnames, name string) {
	data := input.Input{Name: input.Name{NameString: name}}
	enc := gnfmt.GNjson{}
	res, err := bhln.NameRefs(rf, data)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	fmt.Println(enc.Output(res, bhln.Format()))
}

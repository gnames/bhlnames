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

	"github.com/gnames/bhlnames"
	rfs "github.com/gnames/bhlnames/refs"
	"github.com/spf13/cobra"
)

// refsCmd represents the refs command
var refsCmd = &cobra.Command{
	Use:   "refs",
	Short: "Finds references in BHL for name/s",
	Long: `Takes one name string or a file with scientific names and creates
a list of usages/references for the names in Biodiversity Heritage Library.`,
	Run: func(cmd *cobra.Command, args []string) {
		f := formatFlag(cmd)
		d := descFlag(cmd)
		opts = append(opts, bhlnames.OptFormat(f), bhlnames.OptSortDesc(d))
		j := jobsFlag(cmd)
		if j > 0 {
			opts = append(opts, bhlnames.OptJobsNum(j))
		}
		bhln := bhlnames.NewBHLnames(opts...)
		if len(args) == 0 {
			processStdin(cmd, bhln)
			os.Exit(0)
		}
		data := getInput(cmd, args)
		refs(bhln, data)
	},
}

func init() {
	rootCmd.AddCommand(refsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// refsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	refsCmd.Flags().StringP("format", "f", "compact",
		"JSON output format can be 'compact' or 'pretty.")

	refsCmd.Flags().IntP("jobs", "j", 0,
		"Number of parallel jobs to get references.")

	refsCmd.Flags().BoolP("sort_desc", "d", false,
		"Sort references by year in descendent order.")
}

func formatFlag(cmd *cobra.Command) string {
	str, err := cmd.Flags().GetString("format")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return str
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

func processStdin(cmd *cobra.Command, bhln bhlnames.BHLnames) {
	if !checkStdin() {
		_ = cmd.Help()
		return
	}
	refsFile(bhln, os.Stdin)
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

func refs(bhln bhlnames.BHLnames, data string) {
	path := string(data)
	if fileExists(path) {
		f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		refsFile(bhln, f)
		f.Close()
	} else {
		refsString(bhln, data)
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

func refsFile(bhln bhlnames.BHLnames, f io.Reader) {
	in := make(chan string)
	out := make(chan *rfs.RefsResult)
	var wg sync.WaitGroup
	wg.Add(1)

	go bhlnames.RefsStream(bhln, in, out)
	go processResults(bhln.Format, out, &wg)
	sc := bufio.NewScanner(f)
	count := 0
	log.Println("Finding references")
	for sc.Scan() {
		count++
		if count%1000 == 0 {
			log.Printf("Processing %d-th line\n", count)
		}
		name := sc.Text()
		in <- name
	}
	close(in)
	wg.Wait()
	log.Println("Finish finding references")
}

func processResults(format string, out <-chan *rfs.RefsResult,
	wg *sync.WaitGroup) {
	defer wg.Done()
	for r := range out {
		if r.Error != nil {
			log.Println(r.Error)
		}
		fmt.Println(bhlnames.FormatOutput(r.Output, format))
	}
}

func refsString(bhln bhlnames.BHLnames, name string) {
	res, err := bhln.Refs(name)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	fmt.Println(bhlnames.FormatOutput(res, bhln.Format))
}

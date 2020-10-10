/*
Copyright © 2020 Dmitry Mozzherin <dmozzherin@gmail.com>

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
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gdower/bhlinker"
	linkent "github.com/gdower/bhlinker/domain/entity"
	"github.com/gnames/bhlnames"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/gnames/lib/encode"
	"github.com/gnames/gnames/lib/format"
	"github.com/gnames/gnames/lib/sys"
	"github.com/spf13/cobra"
)

// nomenCmd represents the nomen command
var nomenCmd = &cobra.Command{
	Use:   "nomen",
	Short: "returns a possible BHL link to a given nomenclatural event.",
	Long: `Nomenclatural event is a first validly published appearance of
a scientific name. In a simple form nomenclatural event can be described by
a name-string and the corresponding publication.

This command takes a nomenclatural event input and, if found, returns back
a putative link in BHL to the event.
`,
	Run: func(cmd *cobra.Command, args []string) {
		f := formatFlag(cmd)
		j := jobsFlag(cmd)
		y := yearFlag(cmd)
		opts = append(opts,
			config.OptFormat(f),
		)
		if j > 0 {
			opts = append(opts, config.OptJobsNum(j))
		}
		cnf := config.NewConfig(opts...)
		bhln := bhlnames.NewBHLnames(cnf)
		defer bhln.Librarian.Close()
		if len(args) == 0 {
			processStdin(cmd, bhln)
			os.Exit(0)
		}
		data := getInput(cmd, args)
		lnkr := bhlinker.NewBHLinker(bhln, bhln.JobsNum)
		nomen(lnkr, data, y)
	},
}

func init() {
	rootCmd.AddCommand(nomenCmd)

	nomenCmd.Flags().StringP("format", "f", "compact",
		"JSON output format can be 'compact' or 'pretty.")

	nomenCmd.Flags().IntP("jobs", "j", 0,
		"Number of parallel jobs to get references.")

	nomenCmd.Flags().IntP("year", "y", 0,
		"A year when a name was published.")
}

func nomen(lnkr bhlinker.BHLinker, data, year string) {
	path := string(data)
	if sys.FileExists(path) {
		f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		nomensFromFile(lnkr, f)
		f.Close()
	} else {
		nomenFromString(lnkr, data, year)
	}
}

func nomensFromFile(lnkr bhlinker.BHLinker, f io.Reader) {
	chIn := make(chan linkent.Input)
	chOut := make(chan linkent.Output)
	var wg sync.WaitGroup
	wg.Add(1)

	go lnkr.GetLinks(chIn, chOut)
	go processNomenResults(format.CompactJSON, chOut, &wg)

	r := csv.NewReader(f)

	// read header
	header := make(map[string]int)
	hdr, err := r.Read()
	if err != nil {
		log.Fatalf("Cannot read tab-separated file: %s", err)
	}
	for i, v := range hdr {
		header[v] = i
	}
	count := 0
	log.Println("Finding references")
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Cannot read tab-separated row: %s", err)
		}

		count++
		if count%1000 == 0 {
			log.Printf("Processing %s-th line\n", humanize.Comma(int64(count)))
		}
		input := linkent.Input{
			ID: row[header["Id"]],
			Name: linkent.Name{
				NameString: row[header["NameString"]],
				Canonical:  row[header["NameCanonical"]],
				Authorship: row[header["NameAuthorship"]],
				Year:       row[header["NameYear"]],
			},
			Reference: linkent.Reference{
				RefString: row[header["RefString"]],
				Year:      row[header["RefYear"]],
			},
		}
		chIn <- input
	}
	close(chIn)
	wg.Wait()
	log.Println("Finish finding references")
}

func processNomenResults(f format.Format, out <-chan linkent.Output,
	wg *sync.WaitGroup) {
	defer wg.Done()
	enc := encode.GNjson{}
	for r := range out {
		if r.Error != nil {
			log.Println(r.Error)
		}
		fmt.Println(enc.Output(r, f))
	}
}

func nomenFromString(lnkr bhlinker.BHLinker, name string, year string) {
	enc := encode.GNjson{}
	inp := linkent.Input{
		Name:      linkent.Name{NameString: name},
		Reference: linkent.Reference{Year: year},
	}
	res, err := lnkr.GetLink(inp)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	out, _ := enc.Encode(res)
	fmt.Println(string(out))
}

func yearFlag(cmd *cobra.Command) string {
	now := time.Now()
	maxYear := now.Year() + 2
	y, err := cmd.Flags().GetInt("year")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if y < 1750 || y > maxYear {
		return ""
	}
	return strconv.Itoa(y)
}

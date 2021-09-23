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
	"github.com/gnames/bhlnames"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/ent/input"
	"github.com/gnames/bhlnames/ent/namerefs"
	"github.com/gnames/bhlnames/ent/reffinder"
	"github.com/gnames/bhlnames/io/reffinderio"
	"github.com/gnames/gnfmt"
	"github.com/gnames/gnsys"
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
		opts = append(opts, config.OptWithSynonyms(false))
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
		nomen(rf, bhln, data, y)
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

func nomen(rf reffinder.RefFinder, bhln bhlnames.BHLnames, data, year string) {
	path := string(data)
	exists, _ := gnsys.FileExists(path)
	if exists {
		f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		nomensFromFile(rf, bhln, f)
		f.Close()
	} else {
		nomenFromString(rf, bhln, data, year)
	}
}

func nomensFromFile(rf reffinder.RefFinder, bhln bhlnames.BHLnames, f io.Reader) {
	chIn := make(chan input.Input)
	chOut := make(chan *namerefs.NameRefs)
	var wg sync.WaitGroup
	wg.Add(1)

	go bhln.NomenRefsStream(rf, chIn, chOut)
	go processNomenResults(gnfmt.CompactJSON, chOut, &wg)

	r := csv.NewReader(f)

	// read header
	header := make(map[string]int)
	hdr, err := r.Read()
	if err != nil {
		log.Fatalf("Cannot read CSV file: %s", err)
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
	log.Println("Finding references")
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Cannot read CSV row: %s", err)
		}

		count++
		if count%1000 == 0 {
			log.Printf("Processing %s-th line\n", humanize.Comma(int64(count)))
		}
		input := input.Input{
			ID: csvVal(row, "Id"),
			Name: input.Name{
				NameString: csvVal(row, "NameString"),
				NameYear:   csvVal(row, "NameYear"),
				Canonical:  csvVal(row, "NameCanonical"),
				Authorship: csvVal(row, "NameAuthorship"),
			},
			Reference: input.Reference{
				RefString: csvVal(row, "RefString"),
				RefYear:   csvVal(row, "RefYear"),
			},
		}
		chIn <- input
	}
	close(chIn)
	wg.Wait()
	log.Println("Finish finding references")
}

func processNomenResults(f gnfmt.Format, out <-chan *namerefs.NameRefs,
	wg *sync.WaitGroup) {
	defer wg.Done()
	enc := gnfmt.GNjson{}
	for r := range out {
		if r.Error != nil {
			log.Println(r.Error)
		}
		fmt.Println(enc.Output(r, f))
	}
}

func nomenFromString(rf reffinder.RefFinder, bhln bhlnames.BHLnames, name string, year string) {
	enc := gnfmt.GNjson{}
	data := input.Input{
		Name:      input.Name{NameString: name},
		Reference: input.Reference{RefYear: year},
	}
	res, err := bhln.NomenRefs(rf, data)
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

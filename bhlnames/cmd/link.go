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

	"github.com/gdower/bhlinker"
	linkent "github.com/gdower/bhlinker/domain/entity"
	"github.com/gnames/bhlnames"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/data/librarian_pg"
	"github.com/gnames/gnames/lib/encode"
	"github.com/gnames/gnames/lib/format"
	"github.com/gnames/gnames/lib/sys"
	"github.com/spf13/cobra"
)

// linkCmd represents the link command
var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "returns an apparent link to the first description of a name-string",
	Long: `The command "link" tries to find the first nomenclatural reference to
a scientific name-string. It uses name-string itself as well as supplied
information about first publication of a name-string. Ideally the returned
reference from BHL should be the same as this first publication. However it is
not always the case because the publication might be missing in BHL, or the
meta-information about references could not be matched correctly.

The command uses BHL date from the nomenclatura, not taxonomical point of view,
therefore it does not try to resolve a name-string to its currently accepted
name of to find synonyms. Use "refs" command for such purposes.`,
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
		l := librarian_pg.NewLibrarianPG(cnf)
		bhln := bhlnames.NewBHLnames(cnf, l)
		defer l.Close()
		if len(args) == 0 {
			processStdin(cmd, bhln)
			os.Exit(0)
		}
		data := getInput(cmd, args)
		lnkr := bhlinker.NewBHLinker(bhln, bhln.JobsNum)
		link(lnkr, data, y)
	},
}

func link(lnkr bhlinker.BHLinker, data, year string) {
	path := string(data)
	if sys.FileExists(path) {
		f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		linksFromFile(lnkr, f)
		f.Close()
	} else {
		linkFromString(lnkr, data, year)
	}
}

func init() {
	rootCmd.AddCommand(linkCmd)

	linkCmd.Flags().StringP("format", "f", "compact",
		"JSON output format can be 'compact' or 'pretty.")

	linkCmd.Flags().IntP("jobs", "j", 0,
		"Number of parallel jobs to get references.")

	linkCmd.Flags().IntP("year", "y", 0,
		"A year when a name was published.")
}

func linksFromFile(lnkr bhlinker.BHLinker, f io.Reader) {
	chIn := make(chan linkent.Input)
	chOut := make(chan linkent.Output)
	var wg sync.WaitGroup
	wg.Add(1)

	go lnkr.GetLinks(chIn, chOut)
	go processLinkResults(format.CompactJSON, chOut, &wg)

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
			log.Printf("Processing %d-th line\n", count)
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

func processLinkResults(f format.Format, out <-chan linkent.Output,
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

func linkFromString(lnkr bhlinker.BHLinker, name string, year string) {
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

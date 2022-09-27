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
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlnames"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/ent/input"
	"github.com/gnames/bhlnames/ent/namerefs"
	"github.com/gnames/bhlnames/io/bayesio"
	"github.com/gnames/bhlnames/io/reffinderio"
	"github.com/gnames/bhlnames/io/titlemio"
	"github.com/gnames/gnfmt"
	"github.com/gnames/gnparser"
	"github.com/gnames/gnsys"
	"github.com/rs/zerolog/log"
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
		yr := yearFlag(cmd)
		delim := delimiterFlag(cmd)
		curate := curationFlag(cmd)
		var output string
		if curate {
			output = outputFlag(cmd)
		}
		opts = append(opts,
			config.OptFormat(f),
		)
		opts = append(opts, config.OptDelimiter(delim))
		opts = append(opts, config.OptWithSynonyms(false))
		if j > 0 {
			opts = append(opts, config.OptJobsNum(j))
		}
		cfg := config.New(opts...)

		rf := reffinderio.New(cfg)
		defer rf.Close()

		tm := titlemio.New(cfg)
		defer tm.Close()

		gnp := gnparser.New(gnparser.NewConfig())

		bnOpts := []bhlnames.Option{
			bhlnames.OptRefFinder(rf),
			bhlnames.OptParser(gnp),
			bhlnames.OptTitleMatcher(tm),
			bhlnames.OptNLP(bayesio.New()),
		}

		bhln := bhlnames.New(cfg, bnOpts...)

		if len(args) == 0 {
			processStdin(cmd, bhln)
			os.Exit(0)
		}
		data := getInput(cmd, args)
		nomen(bhln, data, yr, curate, output)
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

	nomenCmd.Flags().StringP("delimiter", "d", ",",
		"Delimiter for reading CSV files, default is comma.")

	nomenCmd.Flags().StringP("output", "o", "",
		"output curation results to this file")

	nomenCmd.Flags().BoolP("curation", "c", false,
		"Curate data received from nomen finding.")

}

func nomen(bn bhlnames.BHLnames, data string, year int, curate bool, output string) {
	path := string(data)
	exists, _ := gnsys.FileExists(path)
	if exists {
		f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
		if err != nil {
			err = fmt.Errorf("nomen: %#w", err)
			log.Fatal().Err(err).Msg("nomen")
		}
		nomensFromFile(bn, f, curate, output)
		f.Close()
	} else {
		nomenFromString(bn, data, year)
	}
}

func nomensFromFile(bn bhlnames.BHLnames, f io.Reader, curate bool, output string) {
	chIn := make(chan input.Input)
	chOut := make(chan *namerefs.NameRefs)
	var wg sync.WaitGroup
	wg.Add(1)

	go bn.NomenRefsStream(chIn, chOut)
	go processNomenResults(gnfmt.CompactJSON, chOut, curate, output, &wg)

	r := csv.NewReader(f)
	r.Comma = bn.Config().Delimiter

	// read header
	header := make(map[string]int)
	hdr, err := r.Read()
	if err != nil {
		err = fmt.Errorf("nomensFromFile: %#w", err)
		log.Fatal().Err(err).Msg("Cannot read CSV file")
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
	log.Info().Msg("Finding references.")
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			err = fmt.Errorf("nomensFromFile: %#w", err)
			log.Fatal().Err(err).Msg("Cannot read CSV row.")
		}

		count++
		if count%1000 == 0 {
			log.Info().Msgf("Processing %s-th line.\n", humanize.Comma(int64(count)))
		}
		opts := []input.Option{
			input.OptID(csvVal(row, "Id")),
			input.OptNameString(csvVal(row, "NameString")),
			input.OptRefString(csvVal(row, "RefString")),
		}
		input := input.New(bn, opts...)
		chIn <- input
	}
	close(chIn)
	wg.Wait()
	log.Info().Msg("Finish finding references")
}

func processNomenResults(f gnfmt.Format, out <-chan *namerefs.NameRefs,
	curate bool, output string, wg *sync.WaitGroup) {
	defer wg.Done()

	if curate {
		curateData(out, output)
	} else {
		enc := gnfmt.GNjson{}
		for r := range out {
			if r.Error != nil {
				log.Warn().Err(r.Error)
			}
			fmt.Println(enc.Output(r, f))
		}
	}
}

func nomenFromString(bn bhlnames.BHLnames, name string, year int) {
	enc := gnfmt.GNjson{}
	gnpCfg := gnparser.NewConfig()
	gnp := gnparser.New(gnpCfg)
	opts := []input.Option{
		input.OptNameString(name),
		input.OptNameYear(year),
	}
	data := input.New(gnp, opts...)
	res, err := bn.NomenRefs(data)
	if err != nil {
		err = fmt.Errorf("nomenFromString: %#w", err)
		log.Fatal().Err(err).Msg("nomenFromString")
	}
	out, _ := enc.Encode(res)
	fmt.Println(string(out))
}

func yearFlag(cmd *cobra.Command) int {
	now := time.Now()
	maxYear := now.Year() + 2
	y, err := cmd.Flags().GetInt("year")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if y < 1750 || y > maxYear {
		return 0
	}
	return y
}

func delimiterFlag(cmd *cobra.Command) rune {
	delim, err := cmd.Flags().GetString("delimiter")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	switch delim {
	case "":
		log.Info().Msg("empty delimiter option")
		log.Info().Msg("keeping the default delimiter \",\"")
		return ','
	case "\\t":
		log.Info().Msg("Setting delimiter to \"\\t\"")
		return '\t'
	case ",":
		log.Info().Msg("Setting delimiter to \",\"")
		return ','
	default:
		log.Info().Msg("supported delimiters are \",\" and \"\t\"")
		log.Info().Msg("keeping the default delimiter \",\"")
		return ','
	}
}

func curationFlag(cmd *cobra.Command) bool {
	cur, err := cmd.Flags().GetBool("curation")
	if err != nil {
		err = fmt.Errorf("curationFlag: %#w", err)
		log.Fatal().Err(err).Msg("curationFlag")
	}

	return cur
}

func outputFlag(cmd *cobra.Command) string {
	output, err := cmd.Flags().GetString("output")
	if output == "" {
		err = errors.New("output path for curated results should be set")
	}
	if err != nil {
		err = fmt.Errorf("outputFlag: %#w", err)
		log.Fatal().Err(err).Msg("outputFlag")
	}
	return output
}

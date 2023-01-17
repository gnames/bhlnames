/*
Copyright © 2020-2022 Dmitry Mozzherin <dmozzherin@gmail.com>

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
	"github.com/gnames/bhlnames/internal/io/bayesio"
	"github.com/gnames/bhlnames/internal/io/reffinderio"
	"github.com/gnames/bhlnames/internal/io/restio"
	"github.com/gnames/bhlnames/internal/io/titlemio"
	bhlnames "github.com/gnames/bhlnames/pkg"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/gnparser"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// restCmd represents the rest command
var restCmd = &cobra.Command{
	Use:   "rest",
	Short: "REST service for bhlnames",
	Run: func(cmd *cobra.Command, _ []string) {
		p := portFlag(cmd)
		if p > 0 {
			opts = append(opts, config.OptPortREST(p))
		}
		cfg := config.New(opts...)
		rf := reffinderio.New(cfg)

		tm := titlemio.New(cfg)

		gnp := gnparser.New(gnparser.NewConfig())

		bnOpts := []bhlnames.Option{
			bhlnames.OptRefFinder(rf),
			bhlnames.OptParser(gnp),
			bhlnames.OptTitleMatcher(tm),
			bhlnames.OptNLP(bayesio.New()),
		}

		bn := bhlnames.New(cfg, bnOpts...)
		defer bn.Close()
		api := restio.New(bn)
		api.Run()
	},
}

func init() {
	rootCmd.AddCommand(restCmd)

	restCmd.Flags().IntP("port", "p", 8888, "Port to use for the REST service.")
}

func portFlag(cmd *cobra.Command) int {
	p, err := cmd.Flags().GetInt("port")
	if err != nil {
		log.Fatal().Err(err).Msg("portFlag")
	}
	return p
}

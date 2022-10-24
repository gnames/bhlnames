/*
Copyright © 2022 Dmitry Mozzherin <dmozzherin@gmail.com>

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
	"github.com/gnames/bhlnames/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// taxonCmd represents the taxon command
var taxonCmd = &cobra.Command{
	Use:   "taxon",
	Short: "Finds BHL items dedicated mostly to a particular taxon.",
	Long: `Takes a name of a higher taxon (genus and above) registered in
the Catalogue of Life and finds Items from the Biodiversity Heritage Library
where it is the most prevalent taxon.

Optionally it can also return nomenclatural events on a species level for
the taxon's children.`,
	Run: func(cmd *cobra.Command, args []string) {
		f := formatFlag(cmd)
		nomen := nomenFlag(cmd)
		opts = append(opts,
			config.OptFormat(f),
		)

		cfg := config.New(opts...)

		if len(args) != 1 {
			log.Fatal().Msg("a name of a taxon higher than a genus is needed")
		}

		if nomen {
			nomenEvents(cfg, args[0])
		} else {
			items(cfg, args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(taxonCmd)

	taxonCmd.Flags().BoolP("nomen", "n", false, "provide nomenclatural events for the taxon")
}

func nomenEvents(cfg config.Config, name string) {
	_ = cfg
	log.Info().Msgf("Finding nomenclatural events for taxon '%s'", name)
}

func items(cfg config.Config, name string) {
	_ = cfg
	log.Info().Msgf("Finding BHL items for '%s'", name)
}

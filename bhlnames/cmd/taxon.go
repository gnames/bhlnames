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
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		f := formatFlag(cmd)
		nomen := nomenFlag(cmd)
		opts = append(opts,
			config.OptFormat(f),
		)

		cfg := config.New(opts...)

		if len(args) != 1 {
			log.Fatal().Msg("a name of a higher taxon is needed")
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
	log.Info().Msgf("Finding nomenclatural events for '%s'", name)
}

func items(cfg config.Config, name string) {
	_ = cfg
	log.Info().Msgf("Finding BHL items for '%s'", name)
}

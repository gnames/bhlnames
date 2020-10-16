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
	"log"

	"github.com/gnames/bhlnames"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/data/librarian_pg"
	"github.com/gnames/bhlnames/rest"
	"github.com/spf13/cobra"
)

// restCmd represents the rest command
var restCmd = &cobra.Command{
	Use:   "rest",
	Short: "REST service for bhlnames",
	Run: func(cmd *cobra.Command, args []string) {
		p := portFlag(cmd)
		if p > 0 {
			opts = append(opts, config.OptPortREST(p))
		}
		cfg := config.NewConfig(opts...)
		bhln := bhlnames.NewBHLnames(cfg)
		bhln.Librarian = librarian_pg.NewLibrarianPG(cfg)
		defer bhln.Librarian.Close()
		api := rest.NewAPI(bhln)
		rest.Run(api)
	},
}

func init() {
	rootCmd.AddCommand(restCmd)

	restCmd.Flags().IntP("port", "p", 0, "Port to use for the REST service.")
}

func portFlag(cmd *cobra.Command) int {
	p, err := cmd.Flags().GetInt("port")
	if err != nil {
		log.Fatal(err)
	}
	return p
}

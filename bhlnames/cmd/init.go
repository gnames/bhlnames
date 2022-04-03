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
	"fmt"

	"github.com/gnames/bhlnames"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/io/builderio"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Creates database for bhlnames",
	Long: `Downloads BHL metadata and uses it to create local BHL database. Then it
uses bhlindex grpc service to build additional data about names. When the
process is finished, the program can be used for generating list of
publications for names.

To separate these two processes use "bhlnames bhl" and "bhlnames names" one
after enother. The resul will be identical to "bhlnames init".`,
	Run: func(cmd *cobra.Command, _ []string) {
		rebuild, err := cmd.Flags().GetBool("rebuild")
		if err != nil {
			err = fmt.Errorf("initCmd: %#w", err)
			log.Fatal().Err(err)
		}
		opts = append(opts, config.OptWithRebuild(rebuild))
		cfg := config.New(opts...)

		builder := builderio.New(cfg)
		bn := bhlnames.New(cfg, bhlnames.OptBuilder(builder))

		err = bn.Initialize()
		if err != nil {
			err = fmt.Errorf("initCmd: %#w", err)
			log.Fatal().Err(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	initCmd.Flags().BoolP("rebuild", "r", false, "Delete data and rebuild")
}

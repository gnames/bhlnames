package cmd

import (
	"fmt"
	"os"

	"github.com/gnames/gnfmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func formatFlag(cmd *cobra.Command) gnfmt.Format {
	format := gnfmt.CSV
	s, _ := cmd.Flags().GetString("format")

	if s == "" {
		return format
	}
	if s != "csv" {
		fmt, _ := gnfmt.NewFormat(s)
		if fmt == gnfmt.FormatNone {
			log.Info().Msgf(
				"Cannot set format from '%s', setting format to csv",
				s,
			)
			return format
		}

		format = fmt
	}
	return format
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

func nomenFlag(cmd *cobra.Command) bool {
	n, err := cmd.Flags().GetBool("nomen")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return n
}

func shortFlag(cmd *cobra.Command) bool {
	s, err := cmd.Flags().GetBool("short_output")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return s
}

func noSynonymsFlag(cmd *cobra.Command) bool {
	n, err := cmd.Flags().GetBool("no_synonyms")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return n
}

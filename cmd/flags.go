package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	bhlnames "github.com/gnames/bhlnames/pkg"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/gnfmt"
	"github.com/spf13/cobra"
)

type flagFunc func(*cobra.Command)

// versionFlag checks if version flag is set and prints version information.
func versionFlag(cmd *cobra.Command) {
	b, _ := cmd.Flags().GetBool("version")
	if b {
		version := bhlnames.GetVersion()
		fmt.Printf(
			"\nVersion: %s\nBuild:   %s\n\n",
			version.Version,
			version.Build,
		)
		os.Exit(0)
	}
}

func formatFlag(cmd *cobra.Command) {
	res := gnfmt.CSV
	s, _ := cmd.Flags().GetString("format")

	if s != "" && s != "csv" {
		f, _ := gnfmt.NewFormat(s)
		if f == gnfmt.FormatNone {
			slog.Info(
				"Cannot set format from string, setting it to csv",
				"format-string", s,
			)
		}
		res = f
	}
	opts = append(opts, config.OptFormat(res))
}

func rebuildFlag(cmd *cobra.Command) {
	b, _ := cmd.Flags().GetBool("rebuild")
	if b {
		opts = append(opts, config.OptWithRebuild(b))
	}
}

func jobsFlag(cmd *cobra.Command) {
	i, _ := cmd.Flags().GetInt("jobs")
	if i > 0 {
		opts = append(opts, config.OptJobsNum(i))
	}
}

func descFlag(cmd *cobra.Command) {
	b, _ := cmd.Flags().GetBool("sort_desc")
	if b {
		opts = append(opts, config.OptSortDesc(b))
	}
}

func shortFlag(cmd *cobra.Command) {
	b, _ := cmd.Flags().GetBool("short_output")
	if b {
		opts = append(opts, config.OptWithShortenedOutput(b))
	}
}

func noSynonymsFlag(cmd *cobra.Command) {
	b, _ := cmd.Flags().GetBool("no_synonyms")
	opts = append(opts, config.OptWithSynonyms(!b))
}

func yearFlag(cmd *cobra.Command) int {
	now := time.Now()
	maxYear := now.Year() + 2
	i, _ := cmd.Flags().GetInt("year")
	if i < 1750 || i > maxYear {
		slog.Warn("Year is out of range, ignoring", "year", i)
		return 0
	}
	return i
}

func delimiterFlag(cmd *cobra.Command) {
	s, _ := cmd.Flags().GetString("delimiter")

	var res rune
	switch s {
	case "":
		slog.Info("Empty delimiter option")
		slog.Info("Keeping the default delimiter \",\"")
		res = ','
	case "\\t":
		slog.Info("Setting delimiter to \"\\t\"")
		res = '\t'
	case ",":
		slog.Info("Setting delimiter to \",\"")
		res = ','
	default:
		slog.Info("Supported delimiters are \",\" and \"\t\"")
		slog.Info("Keeping the default delimiter \",\"")
		res = ','
	}
	opts = append(opts, config.OptDelimiter(res))
}

func curationFlag(cmd *cobra.Command) bool {
	b, _ := cmd.Flags().GetBool("curation")
	return b
}

// func outputFlag(cmd *cobra.Command) string {
// 	output, err := cmd.Flags().GetString("output")
// 	if output == "" {
// 		err = errors.New("output path for curated results should be set")
// 	}
// 	if err != nil {
// 		slog.Error("Flag output failed", "error", err)
// 		os.Exit(1)
// 	}
// 	return output
// }

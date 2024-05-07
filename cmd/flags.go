package cmd

import (
	"fmt"
	"os"

	"github.com/gnames/bhlnames/internal/ent/input"
	bhlnames "github.com/gnames/bhlnames/pkg"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/spf13/cobra"
)

// flagFunc sets 'interface' for flags.
type flagFunc func(*cobra.Command)

func curationFlag(cmd *cobra.Command) bool {
	b, _ := cmd.Flags().GetBool("curation")
	return b
}

func descFlag(cmd *cobra.Command) {
	b, _ := cmd.Flags().GetBool("sort_desc")
	if b {
		inpOpts = append(inpOpts, input.OptSortDesc(b))
	}
}

func jobsFlag(cmd *cobra.Command) {
	i, _ := cmd.Flags().GetInt("jobs")
	if i > 0 {
		opts = append(opts, config.OptJobsNum(i))
	}
}

func nomenFlag(cmd *cobra.Command) {
	b, _ := cmd.Flags().GetBool("nomen_event")
	if b {
		inpOpts = append(inpOpts, input.OptWithNomenEvent(b))
	}
}

func portFlag(cmd *cobra.Command) {
	p, _ := cmd.Flags().GetInt("port")
	opts = append(opts, config.OptPortREST(p))
}

func rebuildFlag(cmd *cobra.Command) {
	b, _ := cmd.Flags().GetBool("rebuild")
	if b {
		opts = append(opts, config.OptWithRebuild(b))
	}
}

func refsLimitFlag(cmd *cobra.Command) {
	i, _ := cmd.Flags().GetInt("refs_limit")
	if i > 0 {
		inpOpts = append(inpOpts, input.OptRefsLimit(i))
	}
}

func shortFlag(cmd *cobra.Command) {
	b, _ := cmd.Flags().GetBool("short_output")
	if b {
		inpOpts = append(inpOpts, input.OptWithShortenedOutput(b))
	}
}

func taxonFlag(cmd *cobra.Command) {
	b, _ := cmd.Flags().GetBool("taxon")
	if b {
		inpOpts = append(inpOpts, input.OptWithTaxon(b))
	}
}

func trimFlag(cmd *cobra.Command) {
	b, _ := cmd.Flags().GetBool("trim")
	if b {
		opts = append(opts, config.OptWithCoLDataTrim(b))
	}
}

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

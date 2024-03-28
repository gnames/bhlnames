package cmd

import (
	"fmt"
	"os"

	bhlnames "github.com/gnames/bhlnames/pkg"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/spf13/cobra"
)

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

func rebuildFlag(cmd *cobra.Command) {
	b, _ := cmd.Flags().GetBool("rebuild")
	if b {
		opts = append(opts, config.OptWithRebuild(b))
	}
}

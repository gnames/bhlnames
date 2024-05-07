//go:build tools
// +build tools

package main

import (
	_ "github.com/spf13/cobra"
	_ "github.com/spf13/cobra-cli"
	_ "github.com/swaggo/swag/cmd/swag"
)

package bhlnames

import "github.com/gnames/gnlib/ent/gnvers"

var Version = "v0.2.3"
var Build = "n/a"

// GetVersion returns BHLnames version and build information.
func GetVersion() gnvers.Version {
	return gnvers.Version{Version: Version, Build: Build}
}

package labee

import "github.com/urfave/cli/v2"

var (
	quiet     = false
	flagQuiet = &cli.BoolFlag{
		Name:        "quiet",
		Aliases:     []string{"q"},
		Usage:       "Mute output",
		Destination: &quiet,
	}
)

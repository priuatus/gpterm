package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

var (
	// Version contains the application version number. It's set via ldflags
	// when building.
	Version = ""

	// CommitSHA contains the SHA of the commit that this application was built
	// against. It's set via ldflags when building.
	CommitSHA = ""
)

func main() {

	cli := &CLI{}
	ctx := kong.Parse(
		cli,
		kong.Description("Query OpenAI's GPT models."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: false,
		}),
	)

	if err := ctx.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v", os.Args[0], err)
		os.Exit(1)
	}
}

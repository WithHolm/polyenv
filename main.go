package main

import (
	_ "embed"

	"github.com/withholm/polyenv/cmd"
)

//go:embed CONTRIBUTORS
var Contributors string

func main() {
	cmd.SetContributors(Contributors)
	cmd.Execute()
}

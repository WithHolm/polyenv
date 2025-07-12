package cmd

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

var AuthorsCmd = &cobra.Command{
	Use:   "credits",
	Short: "list the authors and contributors of the project",
	Run:   authors,
}

// set from main.go
var contributors string

func SetContributors(s string) {
	contributors = s
}

func init() {
	rootCmd.AddCommand(AuthorsCmd)
}

func authors(cmd *cobra.Command, args []string) {
	fmt.Println(contributors)
}

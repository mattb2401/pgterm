package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var (
	rootCmd = &cobra.Command{
		Use:   "pgterm",
		Short: "A PostgresDB client for the terminal",
		Long: `pgterm is a CLI Postgres DB client that empowers users with 
easy sql commands that makes managing, querying and 
monitoring your database easier.`,
	}
)

func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

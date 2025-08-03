package cmd

import (
	"fmt"
	"os"

	"github.com/mattb2401/pgterm/internal/pgterm"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	username         string
	host             string
	port             int
	requiresPassword bool
	password         string
	database         string

	connectCmd = &cobra.Command{
		Use:   "connect -h host(optional) -P port(optional) -u username -p <requires password>",
		Short: "Connects to the Postgres database",
		Long: `connect command connects to the database, -h takes a host address, if empty reverts back to localhost, 
    -p is the port of the database instance on which it is running on. 
    -u is required as the username and -p if username password is set`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(username) <= 0 {
				fmt.Println("Username is required, use the -u flag")
				return
			}
			if len(database) <= 0 {
				fmt.Println("Database is required, use the -d flag")
				return
			}
			if cmd.Flags().Changed("requiresPassword") {
				fmt.Print("Enter password: ")
				bPassword, err := term.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					fmt.Println("Error picking password from the stdin")
					return
				}
				password = string(bPassword)
			}
			db, err := pgterm.InitiateConnection(&pgterm.Connection{
				Host:     host,
				Port:     port,
				Username: username,
				Password: password,
				Database: database,
				SSLConfig: pgterm.SSLConfig{
					SSLMode: "disable",
				},
			})
			if err != nil {
				fmt.Println("\nConnection Error: ", err.Error())
				return
			}
			prompt := pgterm.Prompt{
				DB: db,
			}
			prompt.New()
		},
	}
)

func init() {
	rootCmd.AddCommand(connectCmd)
	connectCmd.Flags().BoolP("help", "", false, "")
	connectCmd.Flags().StringVarP(&host, "host", "h", "localhost", "Server address")
	connectCmd.Flags().IntVarP(&port, "port", "P", 5432, "Server Port")
	connectCmd.Flags().StringVarP(&username, "username", "u", "", "")
	connectCmd.Flags().StringVarP(&database, "database", "d", "", "")
	connectCmd.Flags().BoolVarP(&requiresPassword, "requiresPassword", "p", true, "")
}

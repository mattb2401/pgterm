package pgterm

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/c-bata/go-prompt"
)

type Prompt struct {
	DB *sql.DB
}

var currentPrompt *prompt.Prompt
var version, currentUser, currentDatabase string
var buffer []string

func (p *Prompt) New() {
	p.DB.QueryRow("SELECT current_user, current_database(), version()").Scan(&currentUser, &currentDatabase, &version)
	fmt.Print("\n")
	fmt.Println(fmt.Sprintf(`
Welcome to the PgTerm PostgresSQL CLI client v1.0.0.  Commands end with ;.
Your PostgreSQL user ID is %s
Server version: PostgreSQL %s

Copyright (c) 2025 Matt Sebuuma

This software includes open source components licensed under their respective terms.

Licensed under the MIT License.

Type 'help;' or '\h' for help.`, currentUser, extractPostgresVersion(version)))
	session.SetDatabase(currentDatabase)
	fmt.Println("\n")

	currentPrompt = prompt.New(p.executor, p.completer, prompt.OptionPrefix(fmt.Sprintf("pgterm [%s.%s]> ", session.GetDatabase(), session.GetSchema())),
		prompt.OptionPrefixTextColor(prompt.Green),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
		prompt.OptionLivePrefix(func() (string, bool) {
			// Show different prefix if collecting multiline input
			if len(buffer) > 0 {
				return "... ", true
			}
			return fmt.Sprintf("pgterm [%s.%s]> ", session.GetDatabase(), session.GetSchema()), true
		}))
	currentPrompt.Run()
}

func (p *Prompt) completer(in prompt.Document) []prompt.Suggest {
	if len(buffer) > 0 {
		return nil
	}
	//s := []prompt.Suggest{
	//	{Text: "use schema <schema name>;", Description: "Selects a specific schema"},
	//	{Text: "show schemas;", Description: "Show all schemas in this database"},
	//	{Text: "show tables;", Description: "Show all the tables in the selected schema"},
	//	{Text: "describe <table name>;", Description: "Show table structure or ddl"},
	//	{Text: "exit;", Description: "Exit the console and return to obivilion"},
	//	{Text: "quit;", Description: "Quit the console and hide"},
	//}
	s := []prompt.Suggest{}
	return prompt.FilterHasPrefix(s, in.GetWordBeforeCursor(), true)
}

func (p *Prompt) executor(input string) {
	input = strings.TrimSpace(input)
	// ensure that all input has a termination at the end
	if len(input) > 0 {
		termination := input[len(input)-1:]
		if termination != ";" {
			fmt.Println("commands need to be terminated")
		} else {
			switch input {
			case "exit;", "quit;":
				p.DB.Close()
				fmt.Println("Goodbye!")
				os.Exit(0)
			case "help;":
				fmt.Println(helpString())
			case "":
				return
			default:
				trimmed := strings.TrimSpace(input)
				buffer = append(buffer, trimmed)
				if strings.HasSuffix(trimmed, ";") {
					full := strings.Join(buffer, " ")
					buffer = nil
					executor := Executor{
						DB: p.DB,
					}
					resp, promptResetRequired, err := executor.Execute(full)
					if err != nil {
						fmt.Println(err.Error())
					}
					fmt.Println(resp)
					if promptResetRequired {
						p.restartPrompt()
					}
				} else {
					return
				}
			}
		}
	}
}

func (p *Prompt) restartPrompt() {
	// Stop old prompt
	currentPrompt = prompt.New(p.executor, p.completer, prompt.OptionPrefix(fmt.Sprintf("pgterm [%s.%s]> ", session.GetDatabase(), session.GetSchema())),
		prompt.OptionPrefixTextColor(prompt.Green),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
		prompt.OptionLivePrefix(func() (string, bool) {
			// Show different prefix if collecting multiline input
			if len(buffer) > 0 {
				return "... ", true
			}
			return fmt.Sprintf("pgterm [%s.%s]> ", session.GetDatabase(), session.GetSchema()), true
		}))
	currentPrompt.Run()
}

func extractPostgresVersion(input string) string {
	re := regexp.MustCompile(`PostgreSQL\s+(\d+\.\d+)`)
	match := re.FindStringSubmatch(input)
	if len(match) >= 2 {
		return match[1] // e.g., "16.3"
	}
	return ""
}

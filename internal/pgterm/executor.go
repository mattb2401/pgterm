package pgterm

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	pg_query "github.com/pganalyze/pg_query_go/v4"
)

// Executor is responsible for parsing and executing user input SQL/commands against the database.
type Executor struct {
	DB *sql.DB // Active database connection
}

// Execute parses user input, rewrites SQL with schema (if needed), and executes it.
// It handles both SQL commands and internal pseudo-commands like SHOW, USE, etc.
func (e *Executor) Execute(input string) (string, bool, error) {
	tokens := strings.Fields(input)
	if len(tokens) <= 0 {
		return "", false, fmt.Errorf("unsupported command")
	}

	// Interpret the input command to determine its SQL equivalent and metadata.
	sql, executable, sanitize, promptResetRequired, err := e.intepretCommand(input)
	if err != nil {
		return "", promptResetRequired, err
	}

	// Non-executable commands (e.g., "USE SCHEMA x") may change session state only.
	if !executable {
		return sql, promptResetRequired, err
	}

	// If sanitization is needed, add schema names to table references.
	if sanitize {
		sql, err = e.rewriteSQLWithSchema(sql)
		if err != nil {
			return "", promptResetRequired, err
		}
	}

	// Execute the SQL query.
	sqlFields := strings.Fields(sql)
	if sqlFields[0] == "SELECT" {
		now := time.Now()
		rows, err := e.DB.Query(sql)
		if err != nil {
			return "", promptResetRequired, err
		}
		timeDif := time.Since(now).Seconds()
		defer rows.Close()
		// Prepare the table writer for rendering output as markdown.
		columns, _ := rows.Columns()
		table := tablewriter.NewTable(os.Stdout, tablewriter.WithRenderer(renderer.NewMarkdown(
			tw.Rendition{
				Settings: tw.Settings{Separators: tw.Separators{BetweenRows: tw.On}},
				Borders:  tw.Border{Top: tw.On, Bottom: tw.On},
			},
		)), tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignLeft}, // Left-align row data
			},
			Header: tw.CellConfig{
				Formatting: tw.CellFormatting{AutoFormat: tw.On},
				Alignment:  tw.CellAlignment{Global: tw.AlignLeft},
			},
			Footer: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignRight},
			},
		}))

		// Setup containers for scanning row values.
		values := make([]interface{}, len(columns))
		pointers := make([]interface{}, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}

		table.Header(columns)
		rowCount := 0
		// Read and format each row
		for rows.Next() {
			if err := rows.Scan(pointers...); err != nil {
				return "", promptResetRequired, err
			}
			row := make([]string, len(columns))
			for i, val := range values {
				if val == nil {
					row[i] = "NULL"
				} else {
					switch v := val.(type) {
					case []byte:
						row[i] = string(v)
					default:
						row[i] = fmt.Sprintf("%v", v)
					}
				}
			}
			rowCount++
			table.Append(row)
		}
		table.Render()
		return fmt.Sprintf("\n%d rows returned in set (%.3f Sec)", rowCount, timeDif), promptResetRequired, nil
	} else {
		res, err := e.DB.Exec(sql)
		if err != nil {
			return "", promptResetRequired, err
		}
		affected, _ := res.RowsAffected()
		return fmt.Sprintf("%d rows affected", affected), promptResetRequired, nil
	}
}

// intepretCommand parses pseudo-SQL commands like SHOW, USE, DESCRIBE, etc.,
// and returns the equivalent SQL, execution flags, and session state instructions.
func (e *Executor) intepretCommand(cmd string) (string, bool, bool, bool, error) {
	tokens := strings.Fields(cmd)
	if len(tokens) == 0 {
		return "", false, false, false, fmt.Errorf("error: missing command")
	}
	mainCmd := strings.ToUpper(tokens[0])

	switch mainCmd {
	case "SHOW":
		if len(tokens) < 2 {
			return "", false, false, false, fmt.Errorf("missing argument for SHOW")
		}
		subCmd := strings.TrimSuffix(strings.ToUpper(tokens[1]), ";")
		switch subCmd {
		case "SCHEMAS", "schemas":
			return "SELECT schema_name FROM information_schema.schemata;", true, false, false, nil
		case "TABLES", "tables":
			return fmt.Sprintf("SELECT tablename FROM pg_tables WHERE schemaname = '%s';", session.ActiveSchema), true, false, false, nil
		case "DATABASES", "databases":
			return "SELECT datname FROM pg_database WHERE datistemplate = false;", true, false, false, nil
		case "CREATE", "create":
			if len(tokens) < 4 {
				return "", false, false, false, fmt.Errorf("missing argument for SHOW CREATE TABLE")
			}
			table := strings.TrimSuffix(tokens[3], ";")
			return fmt.Sprintf(`
        SELECT 'CREATE TABLE ' || relname || E'\n(\n' ||
        string_agg(
          '  ' || column_name || ' ' || data_type ||
          CASE WHEN character_maximum_length IS NOT NULL
               THEN '(' || character_maximum_length || ')'
          ELSE ''
          END ||
          CASE WHEN is_nullable = 'NO' THEN ' NOT NULL' ELSE '' END,
          E',\n'
        )
        || E'\n);' as create_table
        FROM information_schema.columns c
        JOIN pg_class t ON c.table_name = t.relname
        WHERE table_schema = '%s'
          AND table_name = '%s'
        GROUP BY relname;
    `, session.ActiveSchema, table), true, false, false, nil
		}

	case "DESCRIBE", "DESC":
		if len(tokens) < 2 {
			return "", false, false, false, fmt.Errorf("DESCRIBE needs a table name")
		}
		table := strings.TrimSuffix(tokens[1], ";")
		return fmt.Sprintf(`
            SELECT column_name, data_type, is_nullable
            FROM information_schema.columns
            WHERE table_schema = '%s' AND table_name = '%s';`, session.GetSchema(), table), true, false, false, nil
	case "USE", "use":
		if len(tokens) < 3 {
			return "", false, false, false, fmt.Errorf("missing argument for USE")
		}
		subCmd := strings.ToUpper(tokens[1])
		name := strings.TrimSuffix(strings.ToUpper(tokens[2]), ";")
		switch subCmd {
		case "SCHEMA":
			session.SetSchema(strings.ToLower(name))
			return fmt.Sprintf("Schema changed to %s", session.ActiveSchema), false, false, true, nil
		default:
			return "", false, false, false, fmt.Errorf("Missing argument for USE")
		}
	case "GRANT":
		return cmd, true, false, false, nil // pass through to database without adding schema
	case "CREATE":
		if len(tokens) < 3 {
			return "", false, false, false, fmt.Errorf("missing argument for CREATE")
		}
		subCmd := strings.ToUpper(tokens[1])
		switch subCmd {
		case "TABLE":
			return cmd, true, true, false, nil // sanitize if its a create table to ensure that a
			//table is created in th active schema
		default:
			return cmd, true, false, false, nil
		}
	case "ALTER":
		if len(tokens) < 3 {
			return "", false, false, false, fmt.Errorf("missing argument for ALTER")
		}
		subCmd := strings.ToUpper(tokens[1])
		switch subCmd {
		case "TABLE":
			return cmd, true, true, false, nil // sanitize if its a table table to ensure that a
			//table is created in th active schema
		default:
			return cmd, true, false, false, nil
		}
	default:
		// Treat unknown input as executable SQL that may need schema rewriting
		return cmd, true, true, false, nil
	}
	return "", false, true, false, fmt.Errorf("unsupported command: %s", cmd)
}

// rewriteSQLWithSchema rewrites a SELECT/INSERT/UPDATE/DELETE query
// to prefix table names with the current schema if they are unqualified.
func (e *Executor) rewriteSQLWithSchema(rawSQL string) (string, error) {
	// Parse to Protobuf AST
	tree, err := pg_query.Parse(rawSQL)
	if err != nil {
		return "", fmt.Errorf("invalid SQL: %v", err)
	}

	schema := session.GetSchema()
	rewritten := false

	// Walk all statements in the parsed tree
	for _, stmt := range tree.Stmts {
		node := stmt.Stmt

		switch n := node.Node.(type) {

		case *pg_query.Node_CreateStmt:
			// Add schema to table name
			relation := n.CreateStmt.Relation
			if relation.Schemaname == "" {
				relation.Schemaname = schema
				rewritten = true
			}

		case *pg_query.Node_SelectStmt:
			for _, fromItem := range n.SelectStmt.FromClause {
				if r, ok := fromItem.Node.(*pg_query.Node_RangeVar); ok {
					if r.RangeVar.Schemaname == "" {
						r.RangeVar.Schemaname = schema
						rewritten = true
					}
				}
			}

		case *pg_query.Node_InsertStmt:
			r := n.InsertStmt.Relation
			if r.Schemaname == "" {
				r.Schemaname = schema
				rewritten = true
			}

		case *pg_query.Node_UpdateStmt:
			r := n.UpdateStmt.Relation
			if r.Schemaname == "" {
				r.Schemaname = schema
				rewritten = true
			}

		case *pg_query.Node_DeleteStmt:
			r := n.DeleteStmt.Relation
			if r.Schemaname == "" {
				r.Schemaname = schema
				rewritten = true
			}
		}
	}
	if !rewritten {
		return rawSQL, nil
	}
	// Deparse back to SQL
	modifiedSQL, err := pg_query.Deparse(tree)
	if err != nil {
		return "", fmt.Errorf("deparse error: %v", err)
	}
	return strings.TrimSpace(modifiedSQL), nil
}

package pgterm

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/xwb1989/sqlparser"
)

type Executor struct {
	DB *sql.DB
}

func (e *Executor) Execute(input string) (string, bool, error) {
	tokens := strings.Fields(input)
	if len(tokens) <= 0 {
		return "", false, fmt.Errorf("unsupported command")
	}
	sql, executable, sanitize, promptResetRequired, err := e.intepretCommand(input)
	if err != nil {
		return "", promptResetRequired, err
	}
	if !executable {
		if err != nil {
			return "", promptResetRequired, err
		}
		return sql, promptResetRequired, err
	}
	if sanitize {
		sql, err = e.rewriteSQLWithSchema(sql)
		if err != nil {
			return "", promptResetRequired, err
		}
	}
	rows, err := e.DB.Query(sql)
	if err != nil {
		return "", promptResetRequired, err
	}
	defer rows.Close()
	columns, _ := rows.Columns()
	table := tablewriter.NewTable(os.Stdout, tablewriter.WithRenderer(renderer.NewMarkdown(
		tw.Rendition{
			Settings: tw.Settings{Separators: tw.Separators{BetweenRows: tw.On}},
			Borders: tw.Border{
				Top:    tw.On,
				Bottom: tw.On,
			},
		},
	)), tablewriter.WithConfig(tablewriter.Config{
		Row: tw.CellConfig{
			Alignment: tw.CellAlignment{Global: tw.AlignLeft}, // Left-align rows
		},
		Header: tw.CellConfig{
			Formatting: tw.CellFormatting{AutoFormat: tw.On},
			Alignment:  tw.CellAlignment{Global: tw.AlignLeft},
		},
	}))
	values := make([]interface{}, len(columns))
	pointers := make([]interface{}, len(columns))
	for i := range values {
		pointers[i] = &values[i]
	}
	table.Header(columns)
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
		table.Append(row)
	}
	table.Render()
	return "", promptResetRequired, nil
}

func (e *Executor) intepretCommand(cmd string) (string, bool, bool, bool, error) {
	cmdSplit := strings.Split(cmd, " ")
	switch cmdSplit[0] {
	case "SHOW", "show":
		var cmdDesc string
		if len(cmdSplit) <= 2 {
			cmdDesc = cmdSplit[1][:len(cmdSplit[1])-1]
		} else {
			cmdDesc = cmdSplit[1]
		}
		if len(cmdDesc) == 0 {
			return "", false, false, false, fmt.Errorf("missing argument for SHOW")
		}
		switch cmdDesc {
		case "SCHEMAS", "schemas":
			return "SELECT schema_name FROM information_schema.schemata;", true, false, false, nil
		case "TABLES", "tables":
			return fmt.Sprintf("SELECT tablename FROM pg_tables WHERE schemaname = '%s';", session.ActiveSchema), true, false, false, nil
		case "DATABASES", "databases":
			return "SELECT datname FROM pg_database WHERE datistemplate = false;", true, false, false, nil
		case "CREATE", "create":
			if len(cmdSplit) < 3 {
				return "", false, false, false, fmt.Errorf("missing argument for SHOW CREATE")
			}
			table := cmdSplit[3][:len(cmdSplit[3])-1]
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
	case "DESCRIBE", "DESC", "describe", "desc":
		cmdDesc := cmdSplit[1][:len(cmdSplit[1])-1]
		if len(cmdDesc) == 0 {
			return "", false, false, false, fmt.Errorf("DESCRIBE needs a table name")
		}
		return fmt.Sprintf(`
            SELECT column_name, data_type, is_nullable
            FROM information_schema.columns
            WHERE table_schema = '%s' AND table_name = '%s';`, session.GetSchema(), cmdDesc), true, false, false, nil
	case "USE", "use":
		if len(cmdSplit[1]) == 0 {
			return "", false, false, false, fmt.Errorf("missing argument for USE")
		}
		if len(cmdSplit) < 3 {
			return "", false, false, false, fmt.Errorf("missing argument for SCHEMA")
		}
		cmdDesc := cmdSplit[2][:len(cmdSplit[2])-1]
		switch cmdSplit[1] {
		case "SCHEMA", "schema":
			session.SetSchema(cmdDesc)
			return fmt.Sprintf("Schema changed to %s", session.ActiveSchema), false, false, true, nil
		case "DATABASE", "database":
			session.SetDatabase(cmdDesc)
			return fmt.Sprintf(`\c %s`, cmdDesc), true, false, true, nil
		}
	case "CREATE", "create", "GRANT", "grant", "ALTER", "alter":
		return cmd, true, false, false, nil
	default:
		return cmd, true, true, false, nil
	}
	return "", false, true, false, fmt.Errorf("unsupported command: %s", cmd)
}

func (e *Executor) rewriteSQLWithSchema(rawSQL string) (string, error) {
	stmt, err := sqlparser.Parse(rawSQL)
	if err != nil {
		return "", fmt.Errorf("invalid SQL: %v", err)
	}

	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		for _, expr := range stmt.From {
			if aliased, ok := expr.(*sqlparser.AliasedTableExpr); ok {
				if tableName, ok := aliased.Expr.(sqlparser.TableName); ok {
					// only add schema if it's not already qualified
					if tableName.Qualifier.String() == "" {
						tableName.Qualifier = sqlparser.NewTableIdent(session.GetSchema())
						aliased.Expr = tableName
					}
				}
			}
		}
	case *sqlparser.Delete:
		if tableName, ok := stmt.TableExprs[0].(*sqlparser.AliasedTableExpr); ok {
			if tn, ok := tableName.Expr.(sqlparser.TableName); ok {
				if tn.Qualifier.String() == "" {
					tn.Qualifier = sqlparser.NewTableIdent(session.GetSchema())
					tableName.Expr = tn
				}
			}
		}
	case *sqlparser.Insert:
		tn := stmt.Table
		if tn.Qualifier.String() == "" {
			tn.Qualifier = sqlparser.NewTableIdent(session.GetSchema())
			stmt.Table = tn
		}
	case *sqlparser.Update:
		tn := stmt.TableExprs[0].(*sqlparser.AliasedTableExpr).Expr.(sqlparser.TableName)
		if tn.Qualifier.String() == "" {
			tn.Qualifier = sqlparser.NewTableIdent(session.GetSchema())
			stmt.TableExprs[0].(*sqlparser.AliasedTableExpr).Expr = tn
		}
	default:
		return rawSQL, nil // fallback for unsupported types
	}
	buf := sqlparser.NewTrackedBuffer(nil)
	stmt.Format(buf)
	return buf.String(), nil
}

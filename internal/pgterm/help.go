package pgterm

import "fmt"

func helpString() string {
	return fmt.Sprintf(`
Supported Commands:

SHOW SCHEMAS;
    → Lists all available schemas in the current database.

SHOW TABLES;
    → Lists all tables in the current schema (%s).

SHOW DATABASES;
    → Lists all available databases (excluding templates).

SHOW CREATE TABLE <table>;
    → Outputs a SQL CREATE TABLE statement for the specified table.

DESCRIBE <table>;
DESC <table>;
    → Shows column names, data types, and nullability for the specified table.

USE SCHEMA <schema>;
    → Sets the active schema (affects future queries).

USE DATABASE <dbname>;
    → Connects to a different database (\\c <dbname> equivalent).

CREATE ...;
GRANT ...;
ALTER ...;
    → These commands are passed directly to the database.

All other valid SQL statements (SELECT, INSERT, UPDATE, DELETE, etc.) are supported and passed directly to PostgreSQL.

Note:
    - Semicolons (;) are mandatory.
    - Commands are case-insensitive.
    - Current schema: %s
    - Current database: %s
`, session.GetSchema(), session.GetSchema(), session.GetDatabase())
}

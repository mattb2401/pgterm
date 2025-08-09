# pgterm

**A lightweight PostgreSQL command-line client written in Go.**

Offers MySQL-style commands like `SHOW TABLES`, `DESCRIBE`, `USE SCHEMA`, interactive SQL execution, and more.

---

## 🚀 Features

- MySQL-style commands (`SHOW TABLES`, `DESCRIBE`, `USE SCHEMA`, etc.)
- Auto-translates to PostgreSQL queries
- Interactive or single-command mode
- Dynamic schema switching (`USE SCHEMA audit;`)
- Formatted output with table writer (Markdown or ASCII)
- Built with Go using Cobra and pgx / lib/pq

---

## ⚙️ Installation

```bash
git clone https://github.com/mattb2401/pgterm.git
cd pgterm
go build -o pgterm
# or
go install github.com/mattb2401/pgterm@latest

# brew 
brew tap mattb2401/pgterm
brew install pgterm

# To upgrade 
brew upgrade pgterm
````

---

## 🔧 Usage
By default: host is localhost and port is 5432
-p flag is used if your database requires a password for authentication
```bash
# Start interactive REPL for remote connections
pgterm connect -h x.x.x.x -p 543200 -u myuser -d mydb -p 
```
```bash
# Start interactive REPL for local connections
pgterm connect -u myuser -d mydb -p 
```

### Supported Commands

| Commands Style             |
| -------------------------- 
| `SHOW SCHEMAS;`            
| `SHOW TABLES;`             
| `SHOW CREATE TABLE <tbl>;` 
| `DESCRIBE <tbl>;`          
| `USE SCHEMA <name>;`       
| Other SQL statements       

---

## 🧪 Examples

```
Welcome to the PgTerm PostgresSQL CLI client v1.0.0.  Commands end with ;.
Your PostgreSQL user ID is test
Server version: PostgreSQL 16.4
Type 'help;' for help.

> USE SCHEMA audit;
Schema changed to audit

> SHOW TABLES;
+--------------+
| tablename    |
+--------------+
| users        |
| permissions  |
+--------------+

> DESCRIBE users;
+-------------+-----------+----------+
| column_name | data_type | nullable |
+-------------+-----------+----------+
| id          | integer   | NO       |
| name        | text      | YES      |
+-------------+-----------+----------+

> SHOW CREATE TABLE users;
... DDL of table ...
```

---

## Safety Prompt for UPDATE and DELETE Statements

Starting from this v1.2.5 version, **pgterm** detects `UPDATE` and `DELETE` SQL commands that **do not include a WHERE clause** and 
prompts you for confirmation before executing them. This safety feature helps prevent accidental data loss by making sure 
you explicitly approve potentially dangerous commands.

### How it works

When you run an `UPDATE` or `DELETE` without a `WHERE` clause, pgterm will show a warning and ask:
If you respond with **no** (or anything other than yes/y), the command will be cancelled, and your data remains safe.

### Example session

```sql
pgterm [postgres.test_schema]> update cars set brand = 'Toyota';
WARNING: Your UPDATE statement has NO WHERE clause.
Are you sure you want to continue? (yes/no): no
Safe choice. Query cancelled
````

## 📦 Developer Notes

* Built with Go, Cobra CLI library, and `pgx` or `lib/pq` PostgreSQL driver.
* Query output formatted via `olekukonko/tablewriter`.
* Schema-specific rewriting uses `xwb1989/sqlparser` for robust SQL transformation.

---

## 🧩 Contributing

Contributions are welcome! Feel free to:

* Add more MySQL-style command mappings (e.g. `SHOW INDEXES`, `USER` commands)
* Improve REPL UX (history, auto-completion)
* Extend support for custom SQL parsing & CLI features

---

## ⚖️ License

Licensed under the [MIT License](LICENSE)
© 2025 Matt Sebuuma


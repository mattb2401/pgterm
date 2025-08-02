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
````

---

## 🔧 Usage

```bash
# Start interactive REPL
pgterm connect -h localhost -p 5432 -u myuser -d mydb -p 
```

### Supported Commands

| MySQL Style                | PostgreSQL Equivalent                                       |
| -------------------------- | ----------------------------------------------------------- |
| `SHOW SCHEMAS;`            | `SELECT schema_name FROM information_schema.schemata;`      |
| `SHOW TABLES;`             | `SELECT tablename FROM pg_tables WHERE schemaname = '...';` |
| `SHOW CREATE TABLE <tbl>;` | Derived DDL via SQL on `information_schema`                 |
| `DESCRIBE <tbl>;`          | Column metadata from `information_schema`                   |
| `USE SCHEMA <name>;`       | Sets active schema context for future queries               |
| Other SQL statements       | Passed directly to `pgx` or `database/sql`                  |

---

## 🧪 Examples

```
Welcome to pgterm monitor. Commands end with `;` or `\g`.
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


package pgterm

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Connection struct {
	// Host of the intended database, default localhost
	Host string
	// Port of the intended database, default is 5432
	Port int
	// Username of the database
	Username string
	// Password of the database
	Password string
	// Database name
	Database string
	// SSLConfig defines TLS Credentials if required
	SSLConfig SSLConfig
}

type SSLConfig struct {
	// Postgres ssl modes
	// verify-ca gets to validate the server, does not validate the hostname but requires CA
	// verify-full validates sever, hostname and requires CA
	// Production recommends verify-full
	SSLMode string
	// SSL public certificate. This will be used to prove identity to the Postgres Server
	SSLCert string
	// SSL Certificate Key contains the key that matches the SSLCert
	SSLKey string
	// Trusted root certificate used the client to verify the postgres server certificate.
	SSLRootCert string
}

// InitiateConnection initiates a new postgres connection with the postgres server.
func InitiateConnection(conn *Connection) (*sql.DB, error) {
	if len(conn.SSLConfig.SSLMode) <= 0 {
		return conn.Connect()
	} else {
		if conn.SSLConfig.SSLMode == "disable" {
			return conn.Connect()
		} else {
			return conn.ConnectWithSSL()
		}
	}
}

// Connect initiates a database connection
func (c *Connection) Connect() (*sql.DB, error) {
	psqlDsn := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.Username, c.Password, c.Database)
	db, err := sql.Open("postgres", psqlDsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func (c *Connection) ConnectWithSSL() (*sql.DB, error) {
	psqlDsn := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=%s sslcert=%s sslkey=%s sslrootca=%s",
		c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLConfig.SSLMode, c.SSLConfig.SSLCert, c.SSLConfig.SSLKey, c.SSLConfig.SSLRootCert)
	db, err := sql.Open("postgres", psqlDsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

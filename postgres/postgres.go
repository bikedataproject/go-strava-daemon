package postgres

import "fmt"

/*
import (
	// Import the Posgres driver for the database/sql package
	_ "github.com/lib/pq"
)
*/

// Database : The configuration for Postgres
type Database struct {
	PostgresHost       string `required:"true"`
	PostgresUser       string `required:"true"`
	PostgresPassword   string `required:"true"`
	PostgresPort       int    `default:"5432"`
	PostgresDb         string `required:"true"`
	PostgresRequireSSL string `default:"require"`
}

func (db Database) GetConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%v", db.PostgresHost, db.PostgresPort, db.PostgresUser, db.PostgresPassword, db.PostgresDb, db.PostgresRequireSSL)
}

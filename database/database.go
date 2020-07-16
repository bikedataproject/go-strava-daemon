package database

import (
	"database/sql"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// Database : Struct to hold the database connection
type Database struct {
	PostgresHost       string
	PostgresUser       string
	PostgresPassword   string
	PostgresPort       int
	PostgresDb         string
	PostgresRequireSSL string

	Connection *sql.DB
}

// getDBConnectionString : Generate
func (db Database) getDBConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%v", db.PostgresHost, db.PostgresPort, db.PostgresUser, db.PostgresPassword, db.PostgresDb, db.PostgresRequireSSL)
}

// ping : Check if the database can be reached
func (db Database) ping() {
	if err := db.Connection.Ping(); err == nil {
		log.Infof("Database connection says Pong")
	} else {
		log.Fatalf("Could not reach database")
	}
}

// Connect : Connect to Postgres
func (db Database) Connect() (err error) {
	if db.Connection, err = sql.Open("postgres", db.getDBConnectionString()); err != nil {
		return
	}
	db.ping()
	return
}

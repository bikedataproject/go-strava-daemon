package database

import (
	"database/sql"
	"fmt"

	// Import postgres backend for database/sql module
	_ "github.com/lib/pq"
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
	Connection         *sql.DB
}

// User : Struct to respresent a user object from the database
type User struct {
	ID                int
	Provider          string
	ProviderUser      string
	AccessToken       string
	RefreshToken      string
	TokenCreationDate string
	ExpiresAt         int
	ExpiresIn         int
}

// getDBConnectionString : Generate
func (db Database) getDBConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%v", db.PostgresHost, db.PostgresPort, db.PostgresUser, db.PostgresPassword, db.PostgresDb, db.PostgresRequireSSL)
}

// checkConnection : Check if the database can be reached
func (db Database) checkConnection() bool {
	if err := db.Connection.Ping(); err == nil {
		return true
	}
	return false
}

// Connect : Connect to Postgres
func (db Database) Connect() (err error) {
	if db.Connection, err = sql.Open("postgres", db.getDBConnectionString()); err != nil {
		return
	}
	if db.checkConnection() {
		log.Info("Database is reachable")
	} else {
		log.Fatal("Database is unreachable")
	}
	return
}

// GetUserData : Request a user token for the ID
func (db Database) GetUserData(userID string) (usr User, err error) {
	connection, err := sql.Open("postgres", db.getDBConnectionString())
	if err != nil {
		return
	}
	err = connection.QueryRow("SELECT \"AccessToken\" FROM public.\"Users\" where \"ProviderUser\"=$1", userID).Scan(&usr.AccessToken)
	return
}

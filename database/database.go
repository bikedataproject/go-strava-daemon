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
	ID                string
	UserIdentifier    string
	Provider          string
	ProviderUser      string
	AccessToken       string
	RefreshToken      string
	TokenCreationDate string
	ExpiresAt         int
	ExpiresIn         int
}

// Contribution : Struct to respresent a contribution object from the database
type Contribution struct {
	ContributionID string
	UserAgent      string
	Distance       int
	TimeStampStart string
	TimeStampStop  string
	Duration       int
	PointsGeom     []byte
	PointsTime     []byte
}

// UserContribution : Struct to respresent a UserContribution object from the database
type UserContribution struct {
	UserContributionID string
	UserID             string
	ContributionID     string
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
	err = connection.QueryRow("SELECT \"UserId\", \"AccessToken\" FROM \"Users\" where \"ProviderUser\"=$1", userID).Scan(&usr.ID, &usr.AccessToken)
	return
}

// GetNewContributionID : Get the count of current contributions
func (db Database) GetNewContributionID() (id string, err error) {
	connection, err := sql.Open("postgres", db.getDBConnectionString())
	if err != nil {
		return
	}
	err = connection.QueryRow("SELECT Count(1) FROM \"Contributions\";").Scan(&id)
	return
}

// GetNewUserContributionID : Get the count of current contributions
func (db Database) GetNewUserContributionID() (id string, err error) {
	connection, err := sql.Open("postgres", db.getDBConnectionString())
	if err != nil {
		return
	}
	err = connection.QueryRow("SELECT Count(1) FROM \"UserContributions\";").Scan(&id)
	return
}

// AddContribution : Create new user contribution
func (db Database) AddContribution(contribution Contribution, user User) (err error) {
	// Generate IDs
	newUserContribID, err := db.GetNewUserContributionID()
	if err != nil {
		return
	}

	// Create contributions
	userContrib := UserContribution{
		UserID:             user.ID,
		ContributionID:     contribution.ContributionID,
		UserContributionID: newUserContribID,
	}

	// Connect to database
	connection, err := sql.Open("postgres", db.getDBConnectionString())
	if err != nil {
		return
	}

	// Write Contribution
	query := `
	INSERT INTO "Contributions"
	("ContributionId", "UserAgent", "Distance", "TimeStampStart", "TimeStampStop", "Duration", "PointsGeom", "PointsTime")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	if _, err = connection.Query(query, &contribution.ContributionID, &contribution.UserAgent, &contribution.Distance, &contribution.TimeStampStart, &contribution.TimeStampStop, &contribution.Duration, &contribution.PointsGeom, &contribution.PointsTime); err != nil {
		log.Warnf("Could not insert value into contributions: %v", err)
	}

	// Write WriteUserContribution
	query = `
	INSERT INTO "UserContributions"
	("UserContributionId", "UserId", "ContributionId")
	VALUES ($1, $2, $3)
	`
	if _, err = connection.Query(query, &userContrib.UserContributionID, &userContrib.UserID, &userContrib.ContributionID); err != nil {
		log.Warnf("Could not insert value into contributions: %v", err)
	}

	return
}

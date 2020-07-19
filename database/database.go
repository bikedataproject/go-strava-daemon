package database

import (
	"database/sql"
	"fmt"
	"time"

	// Import postgres backend for database/sql module
	_ "github.com/lib/pq"
	geo "github.com/paulmach/go.geo"
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
	TokenCreationDate time.Time
	ExpiresAt         int
	ExpiresIn         int
}

// Contribution : Struct to respresent a contribution object from the database
type Contribution struct {
	ContributionID string
	UserAgent      string
	Distance       float32
	TimeStampStart time.Time
	TimeStampStop  time.Time
	Duration       int
	PointsGeom     *geo.Path
	PointsTime     []time.Time
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
	err := db.Connection.Ping()
	return err == nil
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
	err = connection.QueryRow(`
	SELECT "Id", "AccessToken"
	FROM "Users"
	WHERE "ProviderUser"=$1;
	`, userID).Scan(&usr.ID, &usr.AccessToken)
	return
}

// GetNewContributionID : Get the count of current contributions
func (db Database) GetNewContributionID() (id string, err error) {
	connection, err := sql.Open("postgres", db.getDBConnectionString())
	if err != nil {
		return
	}
	err = connection.QueryRow(`
	SELECT Count(1)
	FROM "Contributions";
	`).Scan(&id)
	return
}

// GetNewUserContributionID : Get the count of current contributions
func (db Database) GetNewUserContributionID() (id string, err error) {
	connection, err := sql.Open("postgres", db.getDBConnectionString())
	if err != nil {
		return
	}
	err = connection.QueryRow(`
	SELECT Count(1)
	FROM "UserContributions";
	`).Scan(&id)
	return
}

// AddContribution : Create new user contribution
func (db Database) AddContribution(contribution *Contribution, user *User) (err error) {
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
	("ContributionId", "UserAgent", "Distance", "TimeStampStart", "TimeStampStop", "Duration", "PointsGeom")
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	if _, err = connection.Exec(query, contribution.ContributionID, contribution.UserAgent, contribution.Distance, contribution.TimeStampStart, contribution.TimeStampStop, contribution.Duration, contribution.PointsGeom.ToWKT()); err != nil {
		return fmt.Errorf("Could not insert value into contributions: %s", err)
	}

	// Write WriteUserContribution
	query = `
	INSERT INTO "UserContributions"
	("UserContributionId", "UserId", "ContributionId")
	VALUES ($1, $2, $3)
	`
	if _, err = connection.Exec(query, &userContrib.UserContributionID, &userContrib.UserID, &userContrib.ContributionID); err != nil {
		return fmt.Errorf("Could not insert value into contributions: %s", err)
	}
	return
}

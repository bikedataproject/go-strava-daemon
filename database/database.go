package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/bikedataproject/go-bike-data-lib/dbmodel"
	"github.com/lib/pq"

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
func (db Database) GetUserData(userID string) (usr dbmodel.User, err error) {
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
func (db Database) AddContribution(contribution *dbmodel.Contribution, user *dbmodel.User) (err error) {
	// Generate IDs
	newUserContribID, err := db.GetNewUserContributionID()
	if err != nil {
		return
	}

	// Create contributions
	userContrib := dbmodel.UserContribution{
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
	if _, err = connection.Exec(query, contribution.ContributionID, contribution.UserAgent, contribution.Distance, contribution.TimeStampStart, contribution.TimeStampStop, contribution.Duration, contribution.PointsGeom.ToWKT(), pq.Array(contribution.PointsTime)); err != nil {
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

// GetExpiringUsers : Get users which are expiring within half an hour
func (db Database) GetExpiringUsers() (users []dbmodel.User, err error) {
	// Connect to database
	connection, err := sql.Open("postgres", db.getDBConnectionString())
	if err != nil {
		return
	}

	// Fetch expiring users
	response, err := connection.Query(`
	SELECT "Id", "RefreshToken", "UserIdentifier" FROM "Users"
	WHERE "ExpiresAt" <= $1 and "Provider" = 'app/strava';
	`, time.Now().Add(30*time.Minute).Unix())
	if err != nil {
		return
	}

	// Convert sql.Rows into User objects
	for response.Next() {
		var user dbmodel.User
		err = response.Scan(&user.ID, &user.RefreshToken, &user.UserIdentifier)
		if err != nil {
			log.Warnf("Could not add expiring user to result: %v", err)
		}
		users = append(users, user)
	}

	return
}

// UpdateUser : Update an existing user
func (db Database) UpdateUser(user *dbmodel.User) (err error) {
	// Connect to database
	connection, err := sql.Open("postgres", db.getDBConnectionString())
	if err != nil {
		return
	}

	// Update user in database
	_, err = connection.Exec(`
	UPDATE "Users"
	SET "ExpiresAt" = $1,
		"ExpiresIn" = $2,
		"AccessToken" = $3,
		"RefreshToken" = $4
	WHERE "UserIdentifier" = $5;
	`, &user.ExpiresAt, &user.ExpiresIn, &user.AccessToken, &user.RefreshToken, &user.UserIdentifier)
	return
}

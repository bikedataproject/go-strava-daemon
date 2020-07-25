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
	PostgresPort       int64
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
		defer connection.Close()
		return
	}
	err = connection.QueryRow(`
	SELECT "Id", "UserIdentifier", "Provider", "ProviderUser", "AccessToken", "RefreshToken", "TokenCreationDate", "ExpiresAt", "ExpiresIn", "IsHistoryFetched"
	FROM "Users"
	WHERE "ProviderUser"=$1;
	`, userID).Scan(&usr.ID, &usr.UserIdentifier, &usr.Provider, &usr.ProviderUser, &usr.AccessToken, &usr.RefreshToken, &usr.TokenCreationDate, &usr.ExpiresAt, &usr.ExpiresIn, &usr.IsHistoryFetched)
	log.Info(usr)
	defer connection.Close()
	return
}

// AddContribution : Create new user contribution
func (db Database) AddContribution(contribution *dbmodel.Contribution, user *dbmodel.User) (err error) {
	// Connect to database
	connection, err := sql.Open("postgres", db.getDBConnectionString())
	if err != nil {
		defer connection.Close()
		return fmt.Errorf("Could not create database connection: %v", err)
	}

	// Write Contribution
	query := `
	INSERT INTO "Contributions"
	("UserAgent", "Distance", "TimeStampStart", "TimeStampStop", "Duration", "PointsGeom", "PointsTime")
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING "ContributionId";
	`
	response := connection.QueryRow(query, contribution.UserAgent, contribution.Distance, contribution.TimeStampStart, contribution.TimeStampStop, contribution.Duration, contribution.PointsGeom.ToWKT(), pq.Array(contribution.PointsTime))
	defer connection.Close()

	// Create contributions
	userContrib := dbmodel.UserContribution{
		UserID: user.ID,
	}
	response.Scan(&userContrib.ContributionID)

	// Write WriteUserContribution
	query = `
	INSERT INTO "UserContributions"
	("UserId", "ContributionId")
	VALUES ($1, $2);
	`
	if _, err = connection.Exec(query, userContrib.UserID, &userContrib.ContributionID); err != nil {
		defer connection.Close()
		return fmt.Errorf("Could not insert value into contributions: %s", err)
	}

	defer connection.Close()
	return
}

// GetExpiringUsers : Get users which are expiring within half an hour
func (db Database) GetExpiringUsers() (users []dbmodel.User, err error) {
	// Connect to database
	connection, err := sql.Open("postgres", db.getDBConnectionString())
	if err != nil {
		defer connection.Close()
		return
	}

	// Fetch expiring users
	response, err := connection.Query(`
	SELECT "Id", "RefreshToken", "UserIdentifier" FROM "Users"
	WHERE "ExpiresAt" <= $1 and "Provider" = 'web/Strava';
	`, time.Now().Add(30*time.Minute).Unix())
	if err != nil {
		defer connection.Close()
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

	defer connection.Close()
	return
}

// UpdateUser : Update an existing user
func (db Database) UpdateUser(user *dbmodel.User) (err error) {
	// Connect to database

	log.Info(user)
	connection, err := sql.Open("postgres", db.getDBConnectionString())
	if err != nil {
		defer connection.Close()
		return
	}

	// Update user in database
	_, err = connection.Exec(`
	UPDATE "Users"
	SET "ExpiresAt" = $1,
		"ExpiresIn" = $2,
		"AccessToken" = $3,
		"RefreshToken" = $4,
		"IsHistoryFetched" = $5
	WHERE "UserIdentifier" = $6;
	`, &user.ExpiresAt, &user.ExpiresIn, &user.AccessToken, &user.RefreshToken, &user.IsHistoryFetched, &user.UserIdentifier)

	defer connection.Close()
	return
}

// FetchNewUsers : Fetch an array of new users that have not yet fetched their old data
func (db Database) FetchNewUsers() (users []dbmodel.User, err error) {
	// Connect to database
	connection, err := sql.Open("postgres", db.getDBConnectionString())
	if err != nil {
		defer connection.Close()
		return
	}

	// Fetch new users
	response, err := connection.Query(`
	SELECT "Id", "UserIdentifier", "Provider", "ProviderUser", "AccessToken", "RefreshToken", "TokenCreationDate", "ExpiresAt", "ExpiresIn", "IsHistoryFetched"
	FROM "Users"
	WHERE "Provider" = 'web/Strava'
	AND NOT "IsHistoryFetched";
	`)
	if err != nil {
		defer connection.Close()
		return
	}

	// Convert sql.Rows into User objects
	for response.Next() {
		var user dbmodel.User
		err = response.Scan(&user.ID, &user.UserIdentifier, &user.Provider, &user.ProviderUser, &user.AccessToken, &user.RefreshToken, &user.TokenCreationDate, &user.ExpiresAt, &user.ExpiresIn, &user.IsHistoryFetched)
		if err != nil {
			log.Warnf("Could not add expiring user to result: %v", err)
		}
		users = append(users, user)
	}

	defer connection.Close()
	return
}

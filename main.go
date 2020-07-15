package main

import (
	"database/sql"
	"go-strava-daemon/config"
	"go-strava-daemon/inboundhandler"
	"go-strava-daemon/postgres"
	"net/http"

	"github.com/koding/multiconfig"

	log "github.com/sirupsen/logrus"

	// Import the Posgres driver for the database/sql package
	_ "github.com/lib/pq"
)

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

func main() {
	// Load & verify the configuration
	conf := &config.Config{}
	multiconfig.MustLoad(conf)

	dbConf := postgres.Database{
		PostgresHost:       conf.PostgresEndpoint,
		PostgresUser:       conf.PostgresUser,
		PostgresPassword:   conf.PostgresPassword,
		PostgresDb:         conf.PostgresDb,
		PostgresPort:       conf.PostgresPort,
		PostgresRequireSSL: conf.PostgresRequireSSL,
	}
	conStr := dbConf.GetConnectionString()
	log.Info(conStr)
	db, err := sql.Open("postgres", conStr)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Info("pong")
	}

	response, err := db.Query("select * from public.\"Users\"")
	if err == nil {
		for response.Next() {
			var user User
			if err := response.Scan(&user.ID, &user.Provider, &user.ProviderUser, &user.AccessToken, &user.RefreshToken, &user.TokenCreationDate, &user.ExpiresAt, &user.ExpiresIn); err != nil {
				log.Fatal(err)
			} else {
				log.Info(user)
			}
		}

	} else {
		log.Fatal(err)
	}

	// Launch API server
	log.Info("Launching API server")
	http.HandleFunc("/webhook/strava", inboundhandler.HandleStravaMessage)
	log.Fatal(http.ListenAndServe(":5000", nil))
}

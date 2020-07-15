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

	// Launch API server
	log.Info("Launching API server")
	http.HandleFunc("/webhook/strava", inboundhandler.HandleStravaMessage)
	log.Fatal(http.ListenAndServe(":5000", nil))
}

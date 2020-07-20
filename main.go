package main

import (
	// Import the Posgres driver for the database/sql package

	"net/http"
	"time"

	"github.com/koding/multiconfig"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	"go-strava-daemon/config"
	"go-strava-daemon/database"
	"go-strava-daemon/outboundhandler"
)

var db database.Database

func main() {
	// Load configuration values
	conf := &config.Config{}
	multiconfig.MustLoad(conf)

	// Subscribe to Strava
	out := outboundhandler.StravaHandler{
		ClientID:     conf.StravaClientID,
		ClientSecret: conf.StravaClientSecret,
		CallbackURL:  conf.CallbackURL,
		// Generate a new token on restarting
		VerifyToken: string(time.Now().UnixNano()),
		EndPoint:    conf.StravaWebhookURL,
	}

	db = database.Database{
		PostgresHost:       conf.PostgresHost,
		PostgresUser:       conf.PostgresUser,
		PostgresPassword:   conf.PostgresPassword,
		PostgresPort:       conf.PostgresPort,
		PostgresDb:         conf.PostgresDb,
		PostgresRequireSSL: conf.PostgresRequireSSL,
	}
	db.Connect()

	// Unsubscribe from previous connections
	out.UnsubscribeFromStrava()

	// Subscribe in a thread so that the API can go online
	go out.SubscribeToStrava()

	// Launch the API
	log.Info("Launching HTTP API")
	// Handle endpoints - add below if required
	http.HandleFunc("/webhook/strava", HandleStravaWebhook)

	// Run the server untill a Fatal error occurs
	if err := http.ListenAndServe(":5000", nil); err != nil {
		log.Fatalf("Webserver crashed: %v", err)
	}
}

package main

import (
	// Import the Posgres driver for the database/sql package
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/koding/multiconfig"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	"go-strava-daemon/config"
	"go-strava-daemon/database"
	"go-strava-daemon/outboundhandler"
)

// Global variables
var (
	db  database.Database
	out outboundhandler.StravaHandler
)

func main() {
	// Set logging to file
	logfile, err := os.OpenFile(fmt.Sprintf("log/%v.log", time.Now().Unix()), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Could not create logfile: %v", err)
	}
	log.SetOutput(logfile)

	// Load configuration values
	conf := &config.Config{}
	multiconfig.MustLoad(conf)

	// Check configuration type
	if conf.DeploymentType == "production" {
		port, err := strconv.ParseInt(config.ReadSecret("db_port"), 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		conf = &config.Config{
			PostgresHost:       config.ReadSecret("db_host"),
			PostgresUser:       config.ReadSecret("db_user"),
			PostgresPassword:   config.ReadSecret("db_password"),
			PostgresPort:       port,
			PostgresDb:         config.ReadSecret("db_name"),
			StravaClientID:     config.ReadSecret("strava_client_id"),
			StravaClientSecret: config.ReadSecret("strava_client_secret"),
		}
	} else {
		if conf.CallbackURL == "" || conf.PostgresDb == "" || conf.PostgresHost == "" || conf.PostgresPassword == "" || conf.PostgresPort == 0 || conf.PostgresRequireSSL == "" || conf.PostgresUser == "" || conf.StravaClientID == "" || conf.StravaClientSecret == "" || conf.StravaWebhookURL == "" {
			log.Fatal("Configuration not completed")
		}
	}

	// Subscribe to Strava
	out = outboundhandler.StravaHandler{
		ClientID:     conf.StravaClientID,
		ClientSecret: conf.StravaClientSecret,
		CallbackURL:  conf.CallbackURL,
		// Generate a new token on restarting
		VerifyToken: strconv.FormatInt(time.Now().Unix(), 10),
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

	// Handle expiring users from Strava
	go HandleExpiringUsers()

	// Handle fetching data from new Strava users
	go HandleNewUsers()

	// Launch the API
	log.Info("Launching HTTP API")
	// Handle endpoints - add below if required
	http.HandleFunc("/webhook/strava", HandleStravaWebhook)

	// Run the server untill a Fatal error occurs
	if err := http.ListenAndServe(":5000", nil); err != nil {
		log.Fatalf("Webserver crashed: %v", err)
	}
}

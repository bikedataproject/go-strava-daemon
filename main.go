package main

import (
	// Import the Posgres driver for the database/sql package
	"net/http"

	log "github.com/sirupsen/logrus"

	_ "github.com/lib/pq"

	"go-strava-daemon/config"
	"go-strava-daemon/inboundhandler"
	"go-strava-daemon/outboundhandler"

	"github.com/koding/multiconfig"
)

func main() {
	// Load configuration values
	conf := &config.Config{}
	multiconfig.MustLoad(conf)

	// Subscribe to Strava
	out := outboundhandler.StravaHandler{
		ClientID:     conf.StravaClientID,
		ClientSecret: conf.StravaClientSecret,
		CallbackURL:  conf.CallbackURL,
		VerifyToken:  "JustSomeToken",
		EndPoint:     conf.StravaWebhookURL,
	}

	// Subscribe in a thread
	go out.SubscribeToStrava()

	log.Info("Launching HTTP API")
	http.HandleFunc("/webhook/strava", inboundhandler.HandleStravaWebhook)
	log.Fatal(http.ListenAndServe(":5000", nil))
}

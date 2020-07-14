package main

import (
	"go-strava-daemon/config"
	"go-strava-daemon/inboundhandler"
	"go-strava-daemon/outboundhandler"
	"net/http"

	"github.com/koding/multiconfig"

	log "github.com/sirupsen/logrus"
)

func main() {
	// Load & verify the configuration
	conf := &config.Configuration{}
	multiconfig.MustLoad(conf)

	// Set webhook subscriptions
	out := outboundhandler.StravaHandler{
		clientID:     conf.StravaClientID,
		clientSecret: conf.StravaClientSecret,
		callbackURL:  conf.StravaCallbackURL,
		verifyToken:  "",
	}

	// Launch API server
	log.Info("Launching API server")
	http.HandleFunc("/webhook/strava", inboundhandler.HandleStravaMessage)
	log.Fatal(http.ListenAndServe(":5000", nil))
}

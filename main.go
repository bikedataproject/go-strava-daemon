package main

import (
	"go-strava-daemon/config"
	"go-strava-daemon/inboundhandler"
	"go-strava-daemon/outboundhandler"
	"net/http"

	"github.com/koding/multiconfig"

	postgres "github.com/bikedataproject/go-bike-data-lib/postgres"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Load & verify the configuration
	conf := &config.Config{}
	multiconfig.MustLoad(conf)

	// Set webhook subscriptions
	out := outboundhandler.StravaHandler{
		ClientID:     conf.StravaClientID,
		ClientSecret: conf.StravaClientSecret,
		CallbackURL:  conf.StravaCallbackURL,
		VerifyToken:  "",
	}
	log.Info(out)

	tmp := postgres.Database{
		PostgresHost:       conf.PostgresEndpoint,
		PostgresUser:       conf.PostgresUser,
		PostgresPassword:   conf.PostgresPassword,
		PostgresDb:         conf.PostgresDb,
		PostgresPort:       conf.PostgresPort,
		PostgresRequireSSL: conf.PostgresRequireSSL,
	}
	log.Info(tmp)

	// Launch API server
	log.Info("Launching API server")
	http.HandleFunc("/webhook/strava", inboundhandler.HandleStravaMessage)
	log.Fatal(http.ListenAndServe(":5000", nil))
}

package main

import (
	// Import the Posgres driver for the database/sql package
	_ "github.com/lib/pq"

	"go-strava-daemon/config"
	"go-strava-daemon/outboundhandler"

	"github.com/koding/multiconfig"
)

func main() {
	// Load configuration values
	conf := &config.Config{}
	multiconfig.MustLoad(conf)

	// Subscribe to Strava
	tmp := outboundhandler.StravaHandler{
		ClientID:     conf.StravaClientID,
		ClientSecret: conf.StravaClientSecret,
		CallbackURL:  conf.CallbackURL,
		VerifyToken:  "JustSomeToken",
		EndPoint:     conf.StravaWebhookURL,
	}
	tmp.SubscribeToStrava()
}

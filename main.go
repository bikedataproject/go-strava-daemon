package main

import (
	"go-strava-daemon/inboundhandler"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("Launching API server")
	http.HandleFunc("/webhook/strava", inboundhandler.HandleStravaMessage)
	log.Fatal(http.ListenAndServe(":5000", nil))
}

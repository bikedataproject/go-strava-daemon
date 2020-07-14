package inboundhandler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go-strava-daemon/response"

	"github.com/bikedataproject/go-bike-data-lib/strava"

	"github.com/prometheus/common/log"
)

// HandleStravaMessage : Handle incoming webhook messages
func HandleStravaMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var msg strava.StravaWebhookMessage
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			log.Errorf("Could not decode message as StravaMessageSinglepdate: %v", err)
		}
		log.Infof("Received message: %v", msg)
		returnJSON(w, msg)
	} else {
		returnJSON(w, response.RequestResponse{Message: "Call this endpoint using HTTP POST"})
		log.Warnf("Endpoint called with HTTP %s instead of HTTP POST", r.Method)
	}
}

// returnJSON : Return JSON to the requester or log fatally and exit
func returnJSON(target http.ResponseWriter, response interface{}) {
	if msg, err := json.Marshal(response); err != nil {
		log.Fatalf("Could not marshall response message: %v", err)
	} else {
		fmt.Fprintf(target, string(msg))
	}
}

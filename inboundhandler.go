package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bikedataproject/go-bike-data-lib/strava"

	log "github.com/sirupsen/logrus"
)

// ResponseMessage : General response to send on requests
type ResponseMessage struct {
	Message string `json:"message"`
}

// SendJSONResponse : Send a struct as JSON response
func SendJSONResponse(w http.ResponseWriter, obj interface{}) {
	response, err := json.Marshal(&obj)
	if err != nil {
		log.Fatalf("Could not parse response: %v", err)
	} else {
		if _, err := w.Write([]byte(response)); err != nil {
			log.Errorf("Could not send response: %v", err)
		}
	}
}

// HandleStravaWebhook : Handle incoming requests from Strava
func HandleStravaWebhook(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		defer r.Body.Close()
		// Decode the JSON body as struct
		var msg StravaWebhookMessage
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			SendJSONResponse(w, ResponseMessage{
				Message: "Could not decode JSON body",
			})
		} else {
			// Get activity data
			if err := msg.WriteToDatabase(); err != nil {
				log.Errorf("Could not get activity data: %v", err)
			} else {
				SendJSONResponse(w, ResponseMessage{
					Message: "Ok",
				})
			}
		}
	case "GET":
		// Try to fetch a message from Strava
		challenge, err := getURLParam("hub.challenge", r)
		if err != nil {
			log.Warn("Could not get hub challenge from URL params")
		} else {
			log.Info("Received valid Strava verification request")
			msg := strava.WebhookValidationRequest{
				HubChallenge: challenge,
			}
			SendJSONResponse(w, msg)
		}
	default:
		log.Warnf("Received a HTTP %s request instead of GET or POST on webhook handler", r.Method)
		SendJSONResponse(w, ResponseMessage{
			Message: fmt.Sprintf("Use HTTP POST or HTTP GET instead of %v", r.Method),
		})
	}
}

// getURLParam : Request a parameter from the URL
func getURLParam(paramName string, request *http.Request) (result string, err error) {
	result = request.URL.Query()[paramName][0]
	if result == "" {
		err = fmt.Errorf("Param not found in URL")
	}
	return
}

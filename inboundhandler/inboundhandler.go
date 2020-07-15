package inboundhandler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/schema"
	"github.com/prometheus/common/log"
	log "github.com/sirupsen/logrus"
)

// ResponseMessage : General response to send on requests
type ResponseMessage struct {
	Message string `json:"message"`
}

// ValidResponse : Response to send to Strava on a valid request
type ValidResponse struct {
	HubChallenge string `json:"hub.challenge"`
}

// StravaWebhookValidationRequest : Body of the incoming GET request to verify the endpoint
type StravaWebhookValidationRequest struct {
	HubMode        string `json:"hub.mode"`
	HubChallenge   string `json:"hub.challenge"`
	HubVerifyToken string `json:"hub.verify_token"`
}

// StravaWebhookMessage : Body of incoming webhook messages
type StravaWebhookMessage struct {
	ObjectType     string      `json:"object_type"`
	ObjectID       int         `json:"object_id"`
	AspectType     string      `json:"aspect_type"`
	OwnerID        int         `json:"owner_id"`
	SubscriptionID int         `json:"subscription_id"`
	EventTime      int         `json:"event_time"`
	Updates        interface{} `json:"updates"`
}

// SendJsonResponse : Send a struct as JSON response
func SendJsonResponse(w http.ResponseWriter, obj interface{}) {
	response, err := json.Marshal(&obj)
	if err != nil {
		log.Fatalf("Could not parse response: %v", err)
	} else {
		fmt.Fprintf(w, string(response))
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
			SendJsonResponse(w, ResponseMessage{
				Message: "Could not decode JSON body",
			})
		} else {
			SendJsonResponse(w, ResponseMessage{
				Message: "Ok",
			})
		}
		break
	case "GET":
		// Try to fetch a message from Strava
		var decoder = schema.NewDecoder()
		msg := StravaWebhookValidationRequest{}
		if err := decoder.Decode(&msg, r.URL.Query()); err != nil {
			log.Warnf("Could not decode URL parameters into validation request: %v", err)
			SendJsonResponse(w, ResponseMessage{
				Message: "Recieved values were invalid!",
			})
		}
		break
	default:
		log.Warnf("Received a HTTP %s request instead of GET or POST...", r.Method)
		SendJsonResponse(w, ResponseMessage{
			Message: fmt.Sprintf("Use HTTP POST or HTTP GET instead of %v", r.Method),
		})
		break
	}
}

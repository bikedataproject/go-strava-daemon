package inboundhandler

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// ResponseMessage : General response to send on requests
type ResponseMessage struct {
	Message string `json:"message"`
}

// StravaWebhookValidationRequest : Body of the incoming GET request to verify the endpoint
type StravaWebhookValidationRequest struct {
	HubChallenge string `json:"hub.challenge"`
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

// SendJSONResponse : Send a struct as JSON response
func SendJSONResponse(w http.ResponseWriter, obj interface{}) {
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
			SendJSONResponse(w, ResponseMessage{
				Message: "Could not decode JSON body",
			})
		} else {
			log.Infof("Message type: %s, Object type: %s; Object ID: %v", msg.AspectType, msg.ObjectType, msg.ObjectID)
			SendJSONResponse(w, ResponseMessage{
				Message: "Ok",
			})
		}
		break
	case "GET":
		// Try to fetch a message from Strava
		challenge, err := getURLParam("hub.challenge", r)
		if err != nil {
			log.Warn("Could not get hub challenge from URL params")
		} else {
			msg := StravaWebhookValidationRequest{
				HubChallenge: challenge,
			}
			SendJSONResponse(w, msg)
		}

		break
	default:
		log.Warnf("Received a HTTP %s request instead of GET or POST...", r.Method)
		SendJSONResponse(w, ResponseMessage{
			Message: fmt.Sprintf("Use HTTP POST or HTTP GET instead of %v", r.Method),
		})
		break
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

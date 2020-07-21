package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bikedataproject/go-bike-data-lib/dbmodel"
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
			// Get activity data
			if err := msg.WriteToDatabase(); err != nil {
				log.Warnf("Could not get activity data: %v", err)
			} else {
			}
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
			log.Info("Received valid Strava verification request")
			msg := strava.WebhookValidationRequest{
				HubChallenge: challenge,
			}
			SendJSONResponse(w, msg)
		}

		break
	default:
		log.Warnf("Received a HTTP %s request instead of GET or POST on webhook handler", r.Method)
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

// HandleNewUser : Handle the registration of a new user
func HandleNewUser(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Attempt closing the body
		defer r.Body.Close()

		var user *dbmodel.User
		err := json.NewDecoder(r.Body).Decode(user)
		if err != nil && user.ProviderUser != "" {
			log.Errorf("Could not decode JSON body: %v", err)
			SendJSONResponse(w, ResponseMessage{
				Message: "Could not decode JSON body",
			})
		}

		// Fetch user data
		newUser, err := db.GetUserData(user.ProviderUser)
		if err != nil {
			log.Errorf("Could not fetch user data: %v", err)
			SendJSONResponse(w, ResponseMessage{
				Message: "Could not fetch user data",
			})
		}

		// Fetch activities for users in a goroutine
		if err := FetchNewUserActivities(&newUser); err != nil {
			log.Errorf("Could not load data for user %v: %v", newUser.ID, err)
			SendJSONResponse(w, ResponseMessage{
				Message: "Could not fetch user data",
			})
		} else {
			log.Infof("Successfully loaded data for new user %v", newUser.ID)
			SendJSONResponse(w, ResponseMessage{
				Message: "Ok",
			})
		}
		break
	default:
		log.Warnf("Received a HTTP %s request instead of GET or POST on new user handler", r.Method)
		SendJSONResponse(w, ResponseMessage{
			Message: fmt.Sprintf("Use HTTP POST instead of %v", r.Method),
		})
		break
	}
}

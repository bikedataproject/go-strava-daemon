package inboundhandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// StravaWebhookMessage : Body of incoming webhook messages
type StravaWebhookMessage struct {
	ObjectType     string      `json:"object_type"`
	ObjectID       int         `json:"object_id"`
	AspectType     string      `json:"aspect_type"`
	OwnerID        int         `json:"owner_id"`
	SubscriptionID int         `json:"subscription_id"`
	EventTime      int         `json:"event_time"`
	Updates        interface{} `json:"updates"`

	MessageHandler Handler
}

// StravaActivity : Struct representing an activity from Strava
type StravaActivity struct {
	Distance           float64   `json:"distance"`
	MovingTime         int       `json:"moving_time"`
	ElapsedTime        int       `json:"elapsed_time"`
	TotalElevationGain float64   `json:"total_elevation_gain"`
	Type               string    `json:"type"`
	WorkoutType        int       `json:"workout_type"`
	StartDateLocal     time.Time `json:"start_date_local"`
	StartLatlng        []float64 `json:"start_latlng"`
	EndLatlng          []float64 `json:"end_latlng"`
	Map                struct {
		ID              string `json:"id"`
		Polyline        string `json:"polyline"`
		ResourceState   int    `json:"resource_state"`
		SummaryPolyline string `json:"summary_polyline"`
	} `json:"map"`
	Commute bool `json:"commute"`
}

// GetActivityData : Get data for an activity
func (msg StravaWebhookMessage) GetActivityData() (result StravaActivity, err error) {
	if msg.ObjectType == "activity" {
		// Get owner information from database
		user, err := msg.MessageHandler.DatabaseConnection.GetUserData(string(msg.OwnerID))
		if err != nil {
			log.Fatalf("Could not get user information: %v", err)
		}

		// Fetch activity
		client := &http.Client{}
		req, err := http.NewRequest("GET", fmt.Sprintf("https://www.strava.com/api/v3/activities/%v", msg.ObjectID), nil)
		if err != nil {
			log.Fatalf("Could not create request: %v", err)
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", user.AccessToken))

		response, err := client.Do(req)
		if err != nil {
			log.Fatalf("Could not make request: %v", err)
		}
		defer response.Body.Close()

		if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
			log.Fatalf("Could not decode response body: %v", err)
		}
	}
	return
}

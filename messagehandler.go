package main

import (
	"encoding/json"
	"fmt"
	"go-strava-daemon/database"
	"net/http"
	"time"

	geo "github.com/paulmach/go.geo"
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
}

// StravaActivity : Struct representing an activity from Strava
type StravaActivity struct {
	Distance           float32   `json:"distance"`
	MovingTime         int       `json:"moving_time"`
	ElapsedTime        int       `json:"elapsed_time"`
	TotalElevationGain float64   `json:"total_elevation_gain"`
	Type               string    `json:"type"`
	WorkoutType        int       `json:"workout_type"`
	StartDateLocal     time.Time `json:"start_date_local"`
	EndDateLocal       time.Time
	PointsTime         []time.Time
	StartLatlng        []float64         `json:"start_latlng"`
	EndLatlng          []float64         `json:"end_latlng"`
	Map                StravaActivityMap `json:"map"`
	Commute            bool              `json:"commute"`
	LineString         *geo.Path
}

// StravaActivityMap : Struct representing the Map field in an activity message
type StravaActivityMap struct {
	ID              string `json:"id"`
	Polyline        string `json:"polyline"`
	ResourceState   int    `json:"resource_state"`
	SummaryPolyline string `json:"summary_polyline"`
}

// decodePolyline : Convert an encoded polyline into a decoded geo.Path object
func (activity StravaActivity) decodePolyline() {
	activity.LineString = geo.NewPathFromEncoding(activity.Map.Polyline)
}

// createTimeStampArray : Function to create a TimestampArray from the StartDateLocal and ElapsedTime
func (activity StravaActivity) createTimeStampArray() (err error) {
	start := activity.StartDateLocal
	activity.EndDateLocal = start.Add(time.Duration(activity.ElapsedTime))
	nbOfIntervals := 5
	intervalLength := activity.ElapsedTime / nbOfIntervals
	var timeStamps []time.Time
	for i := 0; i < nbOfIntervals; i++ {
		timeStamps = append(timeStamps, start.Add(time.Second*time.Duration((intervalLength*i))))
	}
	activity.PointsTime = timeStamps
	return
}

// ConvertToContribution : Convert a Strava activity to a database contribution
func (activity StravaActivity) ConvertToContribution() (contribution database.Contribution, err error) {
	if newID, err := db.GetNewContributionID(); err == nil {
		// Convert polyline to useable format
		activity.decodePolyline()
		// Generate timestamp per coordinate
		activity.createTimeStampArray()
		contribution = database.Contribution{
			ContributionID: newID,
			UserAgent:      "app/Strava",
			Distance:       activity.Distance,
			TimeStampStart: activity.StartDateLocal,
			TimeStampStop:  activity.EndDateLocal,
			Duration:       activity.ElapsedTime,
			PointsGeom:     activity.LineString,
			PointsTime:     activity.PointsTime,
		}
	}
	return
}

// WriteToDatabase : Write activity message to database
func (msg StravaWebhookMessage) WriteToDatabase() error {
	if msg.ObjectType == "activity" {
		var activity StravaActivity

		// Get owner information from database
		user, err := db.GetUserData(string(msg.OwnerID))
		if err != nil {
			return fmt.Errorf("Could not get user information: %v", err)
		}

		// Fetch activity
		client := &http.Client{}
		req, err := http.NewRequest("GET", fmt.Sprintf("https://www.strava.com/api/v3/activities/%v", msg.ObjectID), nil)
		if err != nil {
			return fmt.Errorf("Could not create request: %v", err)
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", user.AccessToken))

		response, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("Could not make request: %v", err)
		}
		defer response.Body.Close()

		if err := json.NewDecoder(response.Body).Decode(&activity); err != nil {
			return fmt.Errorf("Could not decode response body: %v", err)
		}

		// Check activity type: cycling
		if activity.Type == "Ride" && activity.WorkoutType == 10 {
			// Convert activity to contribution
			contrib, err := activity.ConvertToContribution()
			if err != nil {
				return fmt.Errorf("Could not convert activity to contribution: %v", err)
			}

			// Store in database
			if err = db.AddContribution(contrib, user); err != nil {
				err = fmt.Errorf("Could not save contribution: %v", err)
			} else {
				log.Info("Contribution written to database")
			}
		} else {
			return fmt.Errorf("The activity is not a cycling trip %s", "")
		}
	}
	return nil
}

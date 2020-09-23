package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bikedataproject/go-bike-data-lib/dbmodel"
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
func (activity *StravaActivity) decodePolyline() {
	// Handle empty polyline
	if activity.Map.Polyline != "" {
		activity.LineString = geo.NewPathFromEncoding(activity.Map.Polyline)
	} else {
		activity.LineString = geo.NewPathFromEncoding(activity.Map.SummaryPolyline)
	}
}

// createTimeStampArray : Function to create a TimestampArray from the StartDateLocal and ElapsedTime
func (activity *StravaActivity) createTimeStampArray() error {
	start := activity.StartDateLocal
	activity.EndDateLocal = start.Add(time.Duration(activity.ElapsedTime) * time.Second)
	nbOfIntervals := activity.LineString.PointSet.Length()
	if nbOfIntervals == 0 {
		return fmt.Errorf("There were 0 location points, could not create timestamp array")
	}
	intervalLength := activity.ElapsedTime / nbOfIntervals
	var timeStamps []time.Time
	for i := 0; i < nbOfIntervals; i++ {
		timeStamps = append(timeStamps, start.Add(time.Second*time.Duration((intervalLength*i))))
	}
	activity.PointsTime = timeStamps
	return nil
}

// ConvertToContribution : Convert a Strava activity to a database contribution
func (activity *StravaActivity) ConvertToContribution() (contribution dbmodel.Contribution, err error) {
	// Convert polyline to useable format
	activity.decodePolyline()
	// Generate timestamp per coordinate
	activity.createTimeStampArray()
	contribution = dbmodel.Contribution{
		UserAgent:      "app/Strava",
		Distance:       int(activity.Distance),
		TimeStampStart: activity.StartDateLocal,
		TimeStampStop:  activity.EndDateLocal,
		Duration:       activity.ElapsedTime,
		PointsGeom:     activity.LineString,
		PointsTime:     activity.PointsTime,
	}
	return
}

// writeToCache : Write a StravaWebhookMessage to cache & fetch it later
func (msg *StravaWebhookMessage) writeToCache() error {
	// Marshall message to bytes
	data, err := json.Marshal(&msg)
	if err != nil {
		return err
	}
	// Write to file
	err = ioutil.WriteFile(fmt.Sprintf("%v/%v.tmp", Cachedir, time.Now().Unix()), data, 0644)
	return err
}

// WriteToDatabase : Write activity message to database
func (msg *StravaWebhookMessage) WriteToDatabase() error {
	if msg.ObjectType == "activity" {
		var activity StravaActivity

		// Get owner information from database
		user, err := db.GetUserData(strconv.Itoa(msg.OwnerID))
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

		// Except strava request limit exceeded: write message to cache
		if response.StatusCode == 429 {
			if err := msg.writeToCache(); err != nil {
				return fmt.Errorf("Strava responded with HTTP 429 and could not write message to cache: %v", err)
			}
			return fmt.Errorf("Strava responded with HTTP 429: Too many requests when retrieving activity data (activity %v for user %v)", msg.ObjectID, msg.OwnerID)
		}

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
			if err = db.AddContribution(&contrib, &user); err != nil {
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

// GetFiles : Fetch files in a certain directory
func GetFiles(dir string, filetype string) (files []string, err error) {
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.Contains(path, filetype) {
			files = append(files, path)
		}
		return nil
	})
	return
}

// HandleCache : Check if any files are in cache and fetch their data
func HandleCache() {
	for {
		// Check if there are any cache files
		if files, err := GetFiles(Cachedir, "tmp"); err != nil {
			log.Errorf("Could not fetch cache files: %v", err)
		} else {
			if len(files) < 1 {
				log.Info("No new cache files found")
			} else {
				// Loop over files
				for _, file := range files {
					// Attempt to read cachefile
					if data, err := ioutil.ReadFile(file); err != nil {
						log.Errorf("Could not read cachefile: %v", err)
					} else {
						// Decode into stravawebhookmessage
						var msg *StravaWebhookMessage
						if err := json.Unmarshal(data, &msg); err != nil {
							log.Errorf("Could not decode cachefile into stravawebhookmessage: %v", err)
						} else {
							// Attempt deleting file
							if err := msg.WriteToDatabase(); err != nil {
								log.Errorf("Could not write cachefile content to database: %v", err)
							} else {
								log.Infof("Wrote cachefile (%v) data to database", file)
								// Delete file
								if err := os.Remove(file); err != nil {
									log.Errorf("Could not delete cachefile: %v", err)
								}
							}
						}
					}
				}
			}
		}

		// Sleep for an hour
		time.Sleep(1 * time.Hour)
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bikedataproject/go-bike-data-lib/dbmodel"
	log "github.com/sirupsen/logrus"
)

// HandleExpiringUsers : Handle users which are about to time out
func HandleExpiringUsers() {
	for {
		// Load expiring users
		users, err := db.GetExpiringUsers()
		if err != nil {
			log.Warn(err)
		}

		// Handle expiring users
		for _, user := range users {
			newUser, err := out.RefreshUserSubscription(&user)
			if err != nil {
				log.Warnf("Could not refresh user subscription: %v", err)
			}

			if err = db.UpdateUser(&newUser); err != nil {
				log.Warnf("Could not update user: %v", err)
			}
		}

		// Loop every 10 minutes
		time.Sleep(10 * time.Minute)
	}
}

// HandleNewUsers : Handle the registration of a new user
func HandleNewUsers() {
	for {
		if users, err := db.FetchNewUsers(); err != nil {
			log.Warnf("Could not fetch new users: %v", err)
		} else {
			// Check if there are any users to process
			if len(users) > 0 {
				log.Infof("Fetching Strava activities for %v new users", len(users))

				// Iterate over new users
				for _, user := range users {
					if err := FetchNewUserActivities(&user); err != nil {
						log.Errorf("Could not store new user activities: %v", err)
					} else {
						log.Infof("Fetching user activities for user %v was successfull", user.ID)
					}
					user.IsHistoryFetched = true
					if err := db.UpdateUser(&user); err != nil {
						log.Errorf("Something went wrong updating the user: %v", err)
					}
				}

			} else {
				log.Info("No new users to fetch data for")
			}
		}

		// Loop every 10 minutes
		time.Sleep(10 * time.Second)
	}
}

// FetchNewUserActivities : Handle storing "old" activities of a new user
func FetchNewUserActivities(user *dbmodel.User) error {
	// Fetch activities for user
	var activities []*StravaActivity
	client := &http.Client{}

	// Allow 20000 activities max
	for page := 1; page < 100; page++ {
		req, err := http.NewRequest("GET", fmt.Sprintf("https://www.strava.com/api/v3/athlete/activities?per_page=%v&page=%v", MaxActivities, page), nil)
		if err != nil {
			return fmt.Errorf("Could not get user activities: %v", err)
		}

		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", user.AccessToken))

		res, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("Could not make request: %v", err)
		}

		// Except HTTP 429: too many requests
		if res.StatusCode == 429 {
			log.Warnf("Strava request limit has been reached - storing in database what we have")
			break
		}
		defer res.Body.Close()

		// Attempt to decode response
		var tmpAct []*StravaActivity
		err = json.NewDecoder(res.Body).Decode(&tmpAct)
		if err != nil || len(tmpAct) < 1 {
			return fmt.Errorf("Could not fetch user activities: %v", err)
		}

		// Add to global array
		for _, act := range tmpAct {
			activities = append(activities, act)
		}

		// Check if there's no more - should be if the len is 200
		if len(tmpAct) < (200 * page) {
			break
		}
	}

	log.Infof("Fetching %v activities from strava user %v", len(activities), user.ProviderUser)

	// Write activities to database
	for _, act := range activities {
		// Convert activity
		contrib, err := act.ConvertToContribution()
		if err != nil {
			log.Warnf("Could not convert activity to contribution: %v", err)
		}

		// Get contribution in database
		err = db.AddContribution(&contrib, user)
		if err != nil {
			log.Warnf("Could not upload contribution to database: %v", err)
		} else {
			log.Infof("Added contribution to database")
		}
	}

	return nil
}

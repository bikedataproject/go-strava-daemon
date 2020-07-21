package main

import (
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

		// Loop every 5 minutes
		time.Sleep(10 * time.Minute)
	}
}

// HandleNewUserActivities : Handle storing "old" activities of a new user
func HandleNewUserActivities(user *dbmodel.User) {
	// Fetch activities for user
}

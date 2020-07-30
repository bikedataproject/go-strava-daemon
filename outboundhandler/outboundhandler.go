package outboundhandler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/bikedataproject/go-bike-data-lib/dbmodel"
	"github.com/bikedataproject/go-bike-data-lib/strava"
	log "github.com/sirupsen/logrus"
)

// StravaHandler : Object to handle outgoing Strava requests
type StravaHandler struct {
	ClientID     string
	ClientSecret string
	CallbackURL  string
	VerifyToken  string
	EndPoint     string
}

// makeRequest : Perform a HTTP request
func (conf StravaHandler) makeRequest(endpoint string, httpverb string, payload *bytes.Buffer) (response *http.Response, err error) {
	client := &http.Client{}
	request, err := http.NewRequest(httpverb, endpoint, payload)
	if err != nil {
		return
	}

	response, err = client.Do(request)
	if err != nil {
		return
	}

	if response.StatusCode == 429 {
		err = fmt.Errorf("Got HTTP 429 as response: too many requests")
	}
	return
}

// SubscribeToStrava : Subscribe to the strava webhooks service
func (conf *StravaHandler) SubscribeToStrava() (err error) {
	log.Info("10 seconds idle before subscription request")
	time.Sleep(10 * time.Second)
	log.Info("Subscribing to Strava")
	response, err := conf.makeRequest(fmt.Sprintf("%s?client_id=%s&client_secret=%s&callback_url=%s&verify_token=%s", conf.EndPoint, conf.ClientID, conf.ClientSecret, conf.CallbackURL, conf.VerifyToken), "POST", &bytes.Buffer{})
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	log.Infof("Strava subscription responded with: %v", string(body))
	return
}

// UnsubscribeFromStrava : Delete the current subscription from Strava
func (conf *StravaHandler) UnsubscribeFromStrava() {
	// Get current subscriptions
	response, err := conf.makeRequest(fmt.Sprintf("%v?client_id=%v&client_secret=%v", conf.EndPoint, conf.ClientID, conf.ClientSecret), "GET", &bytes.Buffer{})
	if err != nil {
		log.Fatalf("Could not get active subscriptions: %v", err)
	}
	defer response.Body.Close()

	decoder := json.NewDecoder(response.Body)
	var msg []strava.SubscriptionMessage
	if err := decoder.Decode(&msg); err != nil {
		log.Fatalf("Could not decode subscription message: %v", err)
	}

	for _, m := range msg {
		// Unsubscribe
		client := &http.Client{}
		payload := &bytes.Buffer{}
		writer := multipart.NewWriter(payload)
		_ = writer.WriteField("client_id", conf.ClientID)
		_ = writer.WriteField("client_secret", conf.ClientSecret)
		err := writer.Close()
		if err != nil {
			log.Fatalf("Could not close payload: %v", err)
		}

		request, err := http.NewRequest("DELETE", fmt.Sprintf("%v/%v", conf.EndPoint, m.ID), payload)
		if err != nil {
			log.Fatalf("Could not create HTTP request: %v", err)
		}

		request.Header.Set("Content-Type", writer.FormDataContentType())
		response, err = client.Do(request)
		if err != nil {
			log.Fatalf("Could not make unsubscribe request: %v", err)
		}

		// Handle responsecodes
		switch response.StatusCode {
		case 204:
			log.Infof("Unsubscribed successfully! (ID = %v)", m.ID)
			break
		case 429:
			log.Warnf("Received HTTP 429 when trying to unsubscribe from ID %v", m.ID)
			break
		default:
			break
		}
	}
}

// RefreshUserSubscription : Refresh the subscription from a user
func (conf StravaHandler) RefreshUserSubscription(user *dbmodel.User) (newUser dbmodel.User, err error) {
	// Create HTTPClient
	client := &http.Client{}
	// Initialise data
	payload := strings.NewReader(fmt.Sprintf("client_id=%s&client_secret=%s&grant_type=refresh_token&refresh_token=%s", conf.ClientID, conf.ClientSecret, user.RefreshToken))
	// Prepare request
	req, err := http.NewRequest("POST", "https://www.strava.com/api/v3/oauth/token", payload)
	if err != nil {
		return
	}
	// Set content-type
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// Make request
	response, err := client.Do(req)
	if err != nil {
		return
	}
	defer response.Body.Close()

	// Handle HTTP 429: Too many requests
	if response.StatusCode == 429 {
		err = fmt.Errorf("Strava responded with HTTP 429 (too many requests) trying to refresh user %v's access token", user.ID)
		return
	}

	// Decode body into RefreshMessage
	decoder := json.NewDecoder(response.Body)
	var msg strava.RefreshMessage
	if err := decoder.Decode(&msg); err != nil {
		err = fmt.Errorf("Could not decode subscription refresh message: %v", err)
	}

	if msg.AccessToken == "" && msg.RefreshToken == "" {
		err = fmt.Errorf("Failed to continue subscription refreshing due to HTTP response: %v", response)
		return
	}

	newUser = dbmodel.User{
		ID:             user.ID,
		UserIdentifier: user.UserIdentifier,
		AccessToken:    msg.AccessToken,
		RefreshToken:   msg.RefreshToken,
		ExpiresAt:      msg.ExpiresAt,
		ExpiresIn:      msg.ExpiresIn,
		// User does not expire before 5 hours in the database, data is fetched within 10seconds
		IsHistoryFetched: true,
	}

	return
}

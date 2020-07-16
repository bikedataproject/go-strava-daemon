package outboundhandler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"github.com/prometheus/common/log"
)

// StravaHandler : Object to handle outgoing Strava requests
type StravaHandler struct {
	ClientID     string
	ClientSecret string
	CallbackURL  string
	VerifyToken  string
	EndPoint     string
}

// StravaSubscriptionMessage : Struct that holds the ID of an individual webhook subscription
type StravaSubscriptionMessage struct {
	ID int `json:"id"`
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
	return
}

// SubscribeToStrava : Subscribe to the strava webhooks service
func (conf StravaHandler) SubscribeToStrava() (err error) {
	log.Info("Subscribing to Strava")
	response, err := conf.makeRequest(fmt.Sprintf("%v?client_id=%s&client_secret=%s&callback_url=%s&verify_token=%s", conf.EndPoint, conf.ClientID, conf.ClientSecret, conf.CallbackURL, conf.VerifyToken), "POST", &bytes.Buffer{})
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
func (conf StravaHandler) UnsubscribeFromStrava() {
	// Get current subscriptions
	response, err := conf.makeRequest(fmt.Sprintf("%v?client_id=%v&client_secret=%v", conf.EndPoint, conf.ClientID, conf.ClientSecret), "GET", &bytes.Buffer{})
	if err != nil {
		log.Fatalf("Could not get active subscriptions: %v", err)
	}
	defer response.Body.Close()

	decoder := json.NewDecoder(response.Body)
	var msg []StravaSubscriptionMessage
	if err := decoder.Decode(&msg); err != nil {
		log.Fatalf("Could not decode subscription messages: %v", err)
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
		if response.StatusCode == 204 {
			log.Info("Unsubscribed successfully!")
		}
	}
}

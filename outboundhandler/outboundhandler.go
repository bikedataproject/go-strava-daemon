package outboundhandler

import (
	"bytes"
	"fmt"
	"io/ioutil"
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
	response, err := conf.makeRequest(fmt.Sprintf("https://www.strava.com/api/v3/push_subscriptions?client_id=%s&client_secret=%s&callback_url=%s&verify_token=%s", conf.ClientID, conf.ClientSecret, conf.CallbackURL, conf.VerifyToken), "POST", &bytes.Buffer{})
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
func (conf StravaHandler) UnsubscribeFromStrava() (err error) {
	// Get current subscriptions
	response, err := conf.makeRequest("POST", fmt.Sprintf("https://www.strava.com/api/v3/push_subscriptions?client_id=%s&client_secret=%s&callback_url=%s&verify_token=%s", conf.ClientID, conf.ClientSecret, conf.CallbackURL, conf.VerifyToken), &bytes.Buffer{})
	if err != nil {
		return err
	}
	log.Info(response)

	// Unsubscribe them
	return
}

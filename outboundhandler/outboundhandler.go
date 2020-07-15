package outboundhandler

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

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

// SubscribeToStrava : Subscribe to the strava webhooks service
func (conf StravaHandler) SubscribeToStrava() (err error) {
	client := &http.Client{}
	request, err := http.NewRequest("POST", fmt.Sprintf("https://www.strava.com/api/v3/push_subscriptions?client_id=%s&client_secret=%s&callback_url=%s&verify_token=%s", conf.ClientID, conf.ClientSecret, conf.CallbackURL, conf.VerifyToken), &bytes.Buffer{})
	if err != nil {
		return err
	}

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	log.Infof("Strava subscription responded with: %v", string(body))
	return
}

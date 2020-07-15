package outboundhandler

import (
	"bytes"
	"mime/multipart"
	"net/http"

	"github.com/bikedataproject/go-bike-data-lib/strava"
	"github.com/prometheus/common/log"
)

// StravaHandler : Object to handle outgoing Strava requests
type StravaHandler struct {
	ClientID     string
	ClientSecret string
	CallbackURL  string
	VerifyToken  string
}

// StravaSubscribe : Function to subscribe to a Strava webhook
func (handler StravaHandler) StravaSubscribe(endpoint string, subscribeData strava.StravaSubscribeRequest) {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	_ = writer.WriteField("client_id", handler.ClientID)
	_ = writer.WriteField("client_secret", handler.ClientSecret)
	_ = writer.WriteField("callback_url", handler.CallbackURL)
	_ = writer.WriteField("verify_token", handler.VerifyToken)

	if err := writer.Close(); err != nil {
		log.Fatalf("Could not properly close multipart writer: %v", err)
	}

	client := &http.Client{}
	if req, err := http.NewRequest("POST", endpoint, payload); err != nil {
		log.Fatalf("Could not perform HTTP request: %v", err)
	} else {
		req.Header.Set("Content-Type", writer.FormDataContentType())
		if res, err := client.Do(req); err != nil {
			log.Fatalf("Could not set content-type: %v", err)
		} else {
			defer res.Body.Close()
		}
	}
}

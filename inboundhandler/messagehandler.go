package inboundhandler

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

func (msg StravaWebhookMessage) GetActivityData(userID string, activityID string) (result interface{}, err error) {
	return
}

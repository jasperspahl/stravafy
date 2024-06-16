package worker

type Callback struct {
	ObjectType     string            `json:"object_type"`
	ObjectId       int64             `json:"object_id"`
	AspectType     string            `json:"aspect_type"`
	Updates        map[string]string `json:"updates"`
	OwnerId        int64             `json:"owner_id"`
	SubscriptionId int64             `json:"subscription_id"`
	EventTime      int64             `json:"event_time"`
}


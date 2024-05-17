package messages

// Message Types
const (
	GET_COLLECTIONS = 1
	SELECT          = 2
	STOP            = 3
	GET_START_TIME  = 4
	SUBSCRIBE       = 5
	Synchronize     = 6
)

type StreamerRequest struct {
	MsgType int32       `json:"type"`
	Data    interface{} `json:"data"`
}

type StreamerResponse struct {
	MsgType int32       `json:"type"`
	Data    interface{} `json:"data"`
}

type SelectOptions struct {
	Database     string `json:"database"`
	Collection   string `json:"collection"`
	StartIsoDate string `json:"startIsoDate"`
	EndIsoDate   string `json:"endIsoDate"`
}

type SubscribeOptions struct {
	Database     string `json:"database"`
	Collection   string `json:"collection"`
	PropertyName string `json:"propertyName"`
	Enable       bool   `json:"enable"`
}

type PropertyUpdate struct {
	Time     int64       `json:"time"`
	Name     string      `json:"name"`
	Property interface{} `json:"property"`
}

type Filter struct {
}

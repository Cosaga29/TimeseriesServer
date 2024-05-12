package main

type StreamerRequest struct {
	MsgType int32       `json:"type"`
	Data    interface{} `json:"data"`
}

type StreamerResponse struct {
	MsgType int32       `json:"type"`
	Data    interface{} `json:"data"`
}

// Message Types
const (
	GET_COLLECTIONS = 1
	SELECT          = 2
	STOP            = 3
)

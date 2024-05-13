package main

type StreamerRequest struct {
	MsgType int32       `json:"type"`
	Data    interface{} `json:"data"`
}

type StreamerResponse struct {
	MsgType int32       `json:"type"`
	Data    interface{} `json:"data"`
}

type TsSelectOptions struct {
	Database     string `json:"database"`
	Collection   string `json:"collection"`
	StartIsoDate string `json:"startIsoDate"`
	EndIsoDate   string `json:"endIsoDate"`
}

type SelectOptions struct {
	Database   string      `json:"database"`
	Collection string      `json:"collection"`
	Filter     interface{} `json:"filter"`
}

type Filter struct {
}

// Message Types
const (
	GET_COLLECTIONS = 1
	SELECT          = 2
	STOP            = 3
	GET_START_TIME  = 4
)

package main

import (
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
)

type tsRequest struct {
	MsgType int32       `json:"type"`
	Data    interface{} `json:"data"`
}

type tsResponse struct {
	MsgType int32       `json:"type"`
	Data    interface{} `json:"data"`
}

type TsClient struct {
	Db        *mongo.Client
	Ws        *websocket.Conn
	requests  chan tsRequest
	responses chan tsResponse
}

func CreateTsClient(ws *websocket.Conn, db *mongo.Client) TsClient {
	return TsClient{
		Db:        db,
		Ws:        ws,
		requests:  make(chan tsRequest),
		responses: make(chan tsResponse),
	}
}

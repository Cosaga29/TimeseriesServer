package main

import (
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
)

type TsSignal struct {
	Code int64
}

type TsClient struct {
	Db        *mongo.Client
	Ws        *websocket.Conn
	requests  chan []byte
	responses chan StreamerResponse
	signals   chan TsSignal
}

func CreateTsClient(ws *websocket.Conn, db *mongo.Client) TsClient {
	return TsClient{
		Db:        db,
		Ws:        ws,
		requests:  make(chan []byte),
		responses: make(chan StreamerResponse),
		signals:   make(chan TsSignal),
	}
}

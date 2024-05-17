package client

import (
	"tsserver/m/v2/messages"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
)

type TsSignal struct {
	Code int64
}

type TsClient struct {
	Db            *mongo.Client
	Ws            *websocket.Conn
	requests      chan []byte
	responses     chan messages.StreamerResponse
	signals       chan TsSignal
	subscriptions map[string]chan messages.PropertyUpdate
}

func CreateTsClient(ws *websocket.Conn, db *mongo.Client) TsClient {
	return TsClient{
		Db:            db,
		Ws:            ws,
		requests:      make(chan []byte),
		responses:     make(chan messages.StreamerResponse),
		signals:       make(chan TsSignal),
		subscriptions: make(map[string]chan messages.PropertyUpdate),
	}
}

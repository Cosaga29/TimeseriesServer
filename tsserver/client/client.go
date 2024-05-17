package client

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
)

type TsSignal struct {
	Code int64
}

type TsClient struct {
	Db            *mongo.Client
	Ws            *websocket.Conn
	Clock         int64
	requests      chan []byte
	responses     chan StreamerResponse
	signals       chan TsSignal
	subscriptions map[string]chan PropertyUpdate
}

func CreateTsClient(ws *websocket.Conn, db *mongo.Client) TsClient {
	return TsClient{
		// TODO: Figure out potential sources of where timeseries data could be coming from
		// Note that data doesn't necessarily have to come from a database
		Db:            db,
		Ws:            ws,
		Clock:         0,
		requests:      make(chan []byte),
		responses:     make(chan StreamerResponse),
		signals:       make(chan TsSignal),
		subscriptions: make(map[string]chan PropertyUpdate),
	}
}

func (cli *TsClient) Start(ctx context.Context) {
	// Create and start the consumer
	go cli.handleRequests(ctx)

	// Start the producer
	go cli.handleResponses(ctx)
}

func (cli *TsClient) handleRequests(ctx context.Context) {
	fmt.Println("Client started request consumer")
	for {
		// Read the first byte to determine how the message should be parsed
		messageType, msg, err := cli.Ws.ReadMessage()

		// Handle bad message error
		if messageType == -1 {
			return
		}
		if err != nil {
			fmt.Printf("Error reading message type: %s\n", err.Error())
			return
		}

		// Decode the first byte as the message type
		clientMsgType := int(binary.BigEndian.Uint32(msg))

		switch clientMsgType {
		case GET_COLLECTIONS:
			go cli.GetCollections(ctx)
		case SELECT:
			var opts SelectOptions
			json.Unmarshal(msg[4:], &opts)
			go cli.Select(ctx, &opts)
		case STOP:
			// Signal for the producer to halt messages
			cli.signals <- TsSignal{Code: STOP}
		default:
			fmt.Printf("No handler for request type %d\n", clientMsgType)
		}
	}
}

func (cli *TsClient) handleResponses(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case res := <-cli.responses:
			cli.Ws.WriteJSON(res)
		}
	}
}

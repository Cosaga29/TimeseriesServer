package client

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"
	"tsserver/m/v2/messages"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Interface that should be supported by the client
type Selector interface {
	Start(ctx context.Context) error
	GetCollections(ctx context.Context) error
	Select(ctx context.Context, opts *messages.SelectOptions) error
}

// Public

func (cli *TsClient) Start(ctx context.Context) {
	// Create and start the consumer
	go cli.handleRequests(ctx)

	// Start the producer
	go cli.handleResponses(ctx)
}

func (cli *TsClient) GetCollections(ctx context.Context) {
	var collections = make([]string, 0)
	var result, _ = cli.Db.ListDatabaseNames(ctx, bson.D{})

	for i := range result {
		collectionNames, err := cli.Db.Database(result[i], nil).ListCollectionNames(ctx, bson.M{})
		if err == nil {
			collections = append(collections, collectionNames...)
		}
	}

	cli.responses <- messages.StreamerResponse{MsgType: 1, Data: collections}
}

func (cli *TsClient) Select(ctx context.Context, options *messages.SelectOptions) error {
	// Handle issues with parsing JSON
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("SelectTs Panic")
		}
	}()

	// Create a channel to move the query results into
	resultChan := make(chan interface{})

	start, startErr := time.Parse(time.RFC3339, options.StartIsoDate)
	if startErr != nil {
		return startErr
	}

	end, endErr := time.Parse(time.RFC3339, options.EndIsoDate)
	if endErr != nil {
		return endErr
	}

	// Create the time query
	query := bson.M{
		TS_FIELD: bson.M{
			"$gte": start,
			"$lte": end,
		},
	}

	cur, err := cli.Db.Database(options.Database).Collection(options.Collection).Find(ctx, query)
	if err != nil {
		return err
	}

	// Create a routine to decode the query results and serve them to the
	// results channel
	go cli.channelResults(ctx, cur, resultChan)

	// Continue serving results until the stop signal is given. Note that
	// this allows the query stream to be interrupted at any time by the
	// client
	for {
		select {
		case <-ctx.Done():
			return nil
		case signal := <-cli.signals:
			if signal.Code == messages.STOP {
				return nil
			}
		case result := <-resultChan:
			cli.responses <- messages.StreamerResponse{MsgType: messages.SELECT, Data: result}
		}
	}
}

// Private

func (cli *TsClient) channelResults(ctx context.Context, cur *mongo.Cursor, resultChan chan interface{}) error {
	var results []bson.M

	err := cur.All(ctx, &results)
	if err != nil {
		cli.signals <- TsSignal{Code: messages.STOP}
		return err
	}

	for i := range results {
		resultChan <- results[i]
	}

	cli.signals <- TsSignal{Code: messages.STOP}

	return nil
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
		case messages.GET_COLLECTIONS:
			go cli.GetCollections(ctx)
		case messages.SELECT:
			var opts messages.SelectOptions
			json.Unmarshal(msg[4:], &opts)
			go cli.Select(ctx, &opts)
		case messages.STOP:
			// Signal for the producer to halt messages
			cli.signals <- TsSignal{Code: messages.STOP}
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

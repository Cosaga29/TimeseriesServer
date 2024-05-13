package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

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

// Interface that should be supported by the client
type TsStreamer interface {
	Start(ctx context.Context) error
	GetCollections(ctx context.Context) error
	SelectTs(ctx context.Context, opts *TsSelectOptions) error
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

	// TODO decode out the response
	cli.responses <- StreamerResponse{MsgType: 1, Data: collections}
}

func (cli *TsClient) SelectTs(ctx context.Context, options *TsSelectOptions) {
	// Handle issues with parsing JSON
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Invalid format for select")
		}
	}()

	// Create a channel to move the query results into
	resultChan := make(chan interface{})

	start, startErr := time.Parse(time.RFC3339, options.StartIsoDate)
	end, EndErr := time.Parse(time.RFC3339, options.EndIsoDate)
	if startErr != nil || EndErr != nil {
		fmt.Println("Date parse error!")
		return
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
		fmt.Println(err.Error())
		panic("Error occured querying the database")
	}

	// Create a routine to decode the query results and serve them to the
	// results channel
	go func() {
		var results []bson.M
		err := cur.All(ctx, &results)

		if err == nil {
			for i := range results {
				resultChan <- results[i]
			}
		}

		cli.signals <- TsSignal{Code: STOP}
	}()

	// Continue serving results until the stop signal is given. Note that
	// this allows the query stream to be interupted at any time by the
	// client
	for {
		select {
		case <-ctx.Done():
			return
		case signal := <-cli.signals:
			if signal.Code == STOP {
				return
			}
		case result := <-resultChan:
			// TODO: Put the responses out
			cli.responses <- StreamerResponse{MsgType: 2, Data: result}
		}
	}
}

// Private

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
			var opts TsSelectOptions
			json.Unmarshal(msg[4:], &opts)
			go cli.SelectTs(ctx, &opts)
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

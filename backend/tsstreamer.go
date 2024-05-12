package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

type QueryOptions struct {
	Database   string
	Collection string
	Query      map[string]interface{}
}

// Interface that should be supported by the client
type TsStreamer interface {
	Start(ctx context.Context) error
	GetCollections(ctx context.Context) error
	Select(ctx context.Context, opts *QueryOptions) error
}

// Public

func (cli *TsClient) Start(ctx context.Context) {
	// Create and start the consumer
	go cli.handleRequests()

	// Start the query function
	go cli.handleQueries(ctx)

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

	cli.responses <- StreamerResponse{MsgType: 1, Data: &collections}
}

func (cli *TsClient) Select(ctx context.Context, req interface{}) {
	// Handle issues with parsing JSON
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Invalid format for select")
		}
	}()

	// Get the query information from the request
	opts := req.(map[string]interface{})
	qOpts := QueryOptions{
		Database:   opts["database"].(string),
		Collection: opts["collection"].(string),
		Query:      opts["query"].(map[string]interface{}),
	}

	// Create a channel to move the query results into
	resultChan := make(chan interface{})

	// TODO: Figure out how to pass query parameters to bson
	cur, err := cli.Db.Database(qOpts.Database).Collection(qOpts.Collection).Find(ctx, qOpts.Query)
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
			cli.responses <- StreamerResponse{MsgType: 2, Data: result}
		}
	}
}

// Private

func (cli *TsClient) handleRequests() {
	fmt.Println("Client started request consumer")
	for {
		var req StreamerRequest
		err := cli.Ws.ReadJSON(&req)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		cli.requests <- req
	}
}

func (cli *TsClient) handleQueries(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-cli.requests:
			switch req.MsgType {
			case GET_COLLECTIONS:
				go cli.GetCollections(ctx)
			case SELECT:
				go cli.Select(ctx, req.Data)
			case STOP:
				cli.signals <- TsSignal{Code: STOP}
			default:
				fmt.Printf("No handler for request type %d\n", req.MsgType)
			}
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

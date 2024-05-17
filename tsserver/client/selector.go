package client

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Interface that should be supported by the client
type Selector interface {
	GetCollections(ctx context.Context) error
	Select(ctx context.Context, opts *SelectOptions) error
}

// Public

func (cli *TsClient) GetCollections(ctx context.Context) {
	var collections = make([]string, 0)
	var result, _ = cli.Db.ListDatabaseNames(ctx, bson.D{})

	for i := range result {
		collectionNames, err := cli.Db.Database(result[i], nil).ListCollectionNames(ctx, bson.M{})
		if err == nil {
			collections = append(collections, collectionNames...)
		}
	}

	cli.responses <- StreamerResponse{MsgType: 1, Data: collections}
}

func (cli *TsClient) Select(ctx context.Context, options *SelectOptions) error {
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
			if signal.Code == STOP {
				return nil
			}
		case result := <-resultChan:
			cli.responses <- StreamerResponse{MsgType: SELECT, Data: result}
		}
	}
}

// Private

func (cli *TsClient) channelResults(ctx context.Context, cur *mongo.Cursor, resultChan chan interface{}) error {
	var results []bson.M

	err := cur.All(ctx, &results)
	if err != nil {
		cli.signals <- TsSignal{Code: STOP}
		return err
	}

	for i := range results {
		resultChan <- results[i]
	}

	cli.signals <- TsSignal{Code: STOP}

	return nil
}

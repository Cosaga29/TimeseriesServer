package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	GET_COLLECTIONS   = 1
	SELECT_COLLECTION = 2
	SOME_REQ          = 3
)

// Interface that should be supported by the client
type TsStreamer interface {
	Start(ctx context.Context) error
	GetCollections(ctx context.Context) error
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

	cli.responses <- tsResponse{MsgType: 1, Data: &collections}
}

func (cli *TsClient) Start(ctx context.Context) {
	// Create and start the consumer
	go func() {
		fmt.Println("Client started request consumer")
		for {
			var req tsRequest
			err := cli.Ws.ReadJSON(&req)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			cli.requests <- req
		}
	}()

	// Start the query function
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case req := <-cli.requests:
				switch req.MsgType {
				case GET_COLLECTIONS:
					cli.GetCollections(ctx)
				default:
				}
			}
		}
	}(ctx)

	// Start the producer
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case res := <-cli.responses:
				cli.Ws.WriteJSON(res)
			}
		}
	}(ctx)
}

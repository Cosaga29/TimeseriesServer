package client

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const TS_FIELD = "tstimestamp"

var META_FIELD = "metadata"

var GRANULARITY = "minutes"

func CreateTsCollection(ctx context.Context, db string, collection string, client *mongo.Client) {
	client.Database(db).CreateCollection(ctx, collection, &options.CreateCollectionOptions{
		TimeSeriesOptions: &options.TimeSeriesOptions{
			TimeField: TS_FIELD, MetaField: &META_FIELD, Granularity: &GRANULARITY,
		},
	})
}

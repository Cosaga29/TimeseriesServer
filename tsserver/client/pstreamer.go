package client

import "context"

// This interface allows a client to subscribe
// to property changes and have them live updated
type PStreamer interface {
	Subscribe(ctx context.Context) error
	Synchronize(clock int64, ctx context.Context) error
}

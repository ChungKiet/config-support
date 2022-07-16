package mpubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"google.golang.org/api/option"
)

var Client *pubsub.Client

func InitializeClient(ctx context.Context, projectID string, opts ...option.ClientOption) (*pubsub.Client, error) {
	var err error
	if Client == nil {
		Client, err = pubsub.NewClient(ctx, projectID, opts...)
		if err != nil {
			return nil, fmt.Errorf("pubsub.NewClient %v", err)
		}
	}
	return Client, nil
}

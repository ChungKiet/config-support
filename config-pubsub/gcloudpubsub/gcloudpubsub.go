package gcloudpubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	"example.com/sarang-apis/config-pubsub/publisher"
	"fmt"
	"google.golang.org/api/option"
)

var client *pubsub.Client

// Topics
const (
	TopicMongoConfig = "mongo-config-thetan-support"
)

// Subscriptions
const (
	SubscriptionMongoConfig = "mongo-config-thetan-support-sub"
)

var topicSubscriptions = map[string][]string{
	TopicMongoConfig: {
		SubscriptionMongoConfig,
	},
}

func Init(ctx context.Context, projectId string, opts ...option.ClientOption) (err error) {
	client, err = publisher.InitConfiguration(ctx, projectId, opts...)
	if err != nil {
		return err
	}

	initPublisher(ctx)

	fmt.Println("Google Cloud PubSub initialize successfully !")
	return nil
}

func initPublisher(ctx context.Context) {
	const op = "initPublisher"

	for topic, _ := range topicSubscriptions {
		if err := publisher.PullTopic(ctx, topic); err != nil {
			fmt.Println("Cannot pull topic")
		}
	}
}

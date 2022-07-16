package publisher

import (
	"context"
	"example.com/sarang-apis/config-pubsub/mpubsub"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

var clientMapper = make(map[string]*pubsub.Topic)

func InitConfiguration(ctx context.Context, projectID string, opts ...option.ClientOption) (*pubsub.Client, error) {
	return mpubsub.InitializeClient(ctx, projectID, opts...)
}

func Validate(ctx context.Context, subId string, topicId string) error {
	topic := mpubsub.Client.Topic(topicId)
	if exist, err := topic.Exists(ctx); err != nil {
		return err
	} else if !exist {
		_, err := mpubsub.Client.CreateTopic(ctx, topicId)
		if err != nil {
			return err
		}
	}

	sub := mpubsub.Client.Subscription(subId)
	if exist, err := sub.Exists(ctx); err != nil {
		return err
	} else if !exist {
		if _, err := mpubsub.Client.CreateSubscription(ctx, subId, pubsub.SubscriptionConfig{
			Topic: topic,
		}); err != nil {
			return err
		}
	}

	return nil
}

func PullTopic(ctx context.Context, topicId string) error {
	topic := mpubsub.Client.Topic(topicId)
	exist, err := topic.Exists(ctx)
	if exist {
		clientMapper[topicId] = mpubsub.Client.Topic(topicId)
		clientMapper[topicId].PublishSettings.NumGoroutines = 3
	} else {
		return fmt.Errorf("not found/ err topic with cfg = %s, %+v", topicId, err)
	}

	return nil
}

func PublishMessage(ctx context.Context, topicId string, rawMessage []byte) error {
	clientTopic, exist := clientMapper[topicId]
	if !exist {
		log.Printf("[PublishMessage] [ERROR] not found client for topic, clientMapper: %+v", clientMapper)
		return fmt.Errorf("not found client for topic %v", topicId)
	}

	message := pubsub.Message{
		Data: rawMessage,
	}
	result := clientTopic.Publish(ctx, &message)
	msgId, err := result.Get(ctx)
	if err != nil {
		log.Printf("Publish message successfully. MsgID = %v | Error = %v", msgId, err)
	}
	return err
}

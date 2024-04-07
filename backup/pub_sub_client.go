package backup

import (
	"context"

	"cloud.google.com/go/pubsub"
	"github.com/echenim/data-processor/config"
	"github.com/echenim/data-processor/logger"
)

type PubSubClient struct{}

func NewPubSubClient() *PubSubClient {
	return &PubSubClient{}
}

func StartSubscriber(ctx context.Context, cfg *config.Config, processMessageFunc func(context.Context, *pubsub.Message)) {
	client, err := pubsub.NewClient(ctx, "your_project_id")
	if err != nil {
		logger.Error("Failed to create Pub/Sub client:", err)
		return
	}

	sub := client.Subscription(cfg.PubSubSubscription)
	err = sub.Receive(ctx, processMessageFunc)
	if err != nil {
		logger.Error("Failed to receive messages:", err)
	}
}

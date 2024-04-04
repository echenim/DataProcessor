package clients

import (
	"context"

	"cloud.google.com/go/pubsub"
	"honnef.co/go/tools/config"
)

func StartSubscriber(ctx context.Context, cfg *config.Config, processMessageFunc func(context.Context, *pubsub.Message)) {
	client, err := pubsub.NewClient(ctx, "your_project_id")
	if err != nil {
		// logger.Error("Failed to create Pub/Sub client:", err)
		return
	}

	sub := client.Subscription(cfg.PubSubSubscription)
	err = sub.Receive(ctx, processMessageFunc)
	if err != nil {
		// logger.Error("Failed to receive messages:", err)
	}
}

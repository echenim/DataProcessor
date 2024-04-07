package clients

import (
	"context"

	"cloud.google.com/go/pubsub"
	"github.com/sirupsen/logrus"
)

type PubSubClient struct {
	subscription *pubsub.Subscription
}

// NewClient now accepts a *pubsub.Client and a subscription name,
// rather than just the configuration. This allows for more direct use
// of the pubsub library's Subscription object.
func NewPubSubClient(pubsubClient *pubsub.Client, subscriptionName string) *PubSubClient {
	sub := pubsubClient.Subscription(subscriptionName)
	return &PubSubClient{subscription: sub}
}

func (c *PubSubClient) ReceiveMessages(ctx context.Context, handleFunc func(ctx context.Context, msg *pubsub.Message)) error {
	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Setup the receive configuration as needed here. This example uses synchronous pulling.
	err := c.subscription.Receive(cctx, handleFunc)
	if err != nil {
		logrus.WithError(err).Error("Failed to receive messages")
		return err
	}

	return nil
}

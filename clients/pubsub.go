package clients

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
)

var GooglePubSubClient *pubsub.Client

type GooglePubSubOptions struct {
	Host      string
	ProjectID string
}

func InitGooglePubSub(opts *GooglePubSubOptions) error {
	if opts == nil {
		return nil
	}
	if err := os.Setenv("PUBSUB_EMULATOR_HOST", opts.Host); err != nil {
		return fmt.Errorf("failed to set env for pubsub emulator host: %w", err)
	}

	pubSubClient, err := pubsub.NewClient(
		context.Background(),
		opts.ProjectID,
	)
	if err != nil {
		return fmt.Errorf("failed to create pubsub client: %w", err)
	}
	GooglePubSubClient = pubSubClient
	return nil
}

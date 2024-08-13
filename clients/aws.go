package clients

import (
	"context"
	"fmt"
	"net/url"

	transport "github.com/aws/smithy-go/endpoints"
)

type AWSOptions struct {
	Host   string
	Region string
	Key    string
	Secret string
}

func endpoint[T any](s string) EndpointResolver[T] {
	uri, err := url.Parse(s)
	if err != nil {
		panic(fmt.Errorf("failed to parse sqs host: %w", err))
	}
	return EndpointResolver[T](func(ctx context.Context, params T) (transport.Endpoint, error) {
		return transport.Endpoint{
			URI: *uri,
		}, nil
	})
}

type EndpointResolver[T any] func(ctx context.Context, params T) (transport.Endpoint, error)

func (e EndpointResolver[T]) ResolveEndpoint(ctx context.Context, params T) (transport.Endpoint, error) {
	return e(ctx, params)
}

package clients

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

var SQSClient *sqs.Client

type SQSOptions struct {
	Host   string
	Key    string
	Secret string
}

func InitSQS(opts *SQSOptions) error {
	if opts == nil {
		return nil
	}
	staticCredentialProvider := aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
		return aws.Credentials{
			AccessKeyID:     opts.Key,
			SecretAccessKey: opts.Secret,
		}, nil
	})

	cfg, err := awsconfig.LoadDefaultConfig(
		context.TODO(),
		awsconfig.WithRegion("dev"),
		awsconfig.WithRetryMaxAttempts(3),
		awsconfig.WithCredentialsProvider(aws.CredentialsProvider(staticCredentialProvider)),
	)
	if err != nil {
		return fmt.Errorf("failed to init sqs config: %w", err)
	}

	SQSClient = sqs.NewFromConfig(
		cfg,
		sqs.WithEndpointResolverV2(endpoint[sqs.EndpointParameters](opts.Host)),
	)

	return nil
}

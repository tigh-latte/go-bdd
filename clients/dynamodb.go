package clients

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var DynamoDBClient *dynamodb.Client

type DynamoDBOptions struct {
	Host   string
	Key    string
	Secret string
}

func InitDynamoDB(opts *DynamoDBOptions) error {
	fmt.Printf("TIGH %#v\n", *opts)
	staticCredentialProvider := aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
		return aws.Credentials{
			AccessKeyID:     opts.Key,
			SecretAccessKey: opts.Secret,
		}, nil
	})

	cfg, err := awsconfig.LoadDefaultConfig(
		context.TODO(),
		awsconfig.WithRetryMaxAttempts(3),
		awsconfig.WithCredentialsProvider(aws.CredentialsProvider(staticCredentialProvider)),
	)
	if err != nil {
		return fmt.Errorf("failed to init sqs config: %w", err)
	}

	DynamoDBClient = dynamodb.NewFromConfig(
		cfg,
		dynamodb.WithEndpointResolverV2(endpoint[dynamodb.EndpointParameters](opts.Host)),
	)

	return nil
}

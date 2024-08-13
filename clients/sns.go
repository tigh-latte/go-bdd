package clients

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

var SNSClient *sns.Client

func InitSNS(opts *AWSOptions) error {
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
		awsconfig.WithRegion(opts.Region),
		awsconfig.WithRetryMaxAttempts(3),
		awsconfig.WithCredentialsProvider(aws.CredentialsProvider(staticCredentialProvider)),
	)
	if err != nil {
		return fmt.Errorf("failed to init sns config: %w", err)
	}

	SNSClient = sns.NewFromConfig(
		cfg,
		sns.WithEndpointResolverV2(endpoint[sns.EndpointParameters](opts.Host)),
	)

	return nil
}

package bdd

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"
)

var (
	s3Client *s3.Client
)

func initS3(opts *s3Options) error {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string,
		options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:       "aws",
			URL:               opts.Host,
			SigningRegion:     "dev",
			HostnameImmutable: true,
		}, nil
	})

	statusCredentialProvider := aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
		return aws.Credentials{
			AccessKeyID:     opts.Key,
			SecretAccessKey: opts.Secret,
		}, nil
	})

	cfg, err := awsconfig.LoadDefaultConfig(
		context.TODO(),
		awsconfig.WithRegion("dev"),
		awsconfig.WithEndpointResolverWithOptions(customResolver),
		awsconfig.WithCredentialsProvider(aws.CredentialsProvider(statusCredentialProvider)),
	)
	if err != nil {
		return err
	}

	if s3Client = s3.NewFromConfig(cfg); s3Client == nil {
		return errors.New("aws client nil")
	}

	return nil
}

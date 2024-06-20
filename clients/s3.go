package clients

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var S3Client *s3.Client

type S3Options struct {
	Host   string
	Key    string
	Secret string
}

func InitS3(opts *S3Options) error {
	// Nothing to init.
	if opts == nil {
		return nil
	}

	statusCredentialProvider := aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
		return aws.Credentials{
			AccessKeyID:     opts.Key,
			SecretAccessKey: opts.Secret,
		}, nil
	})

	cfg, err := awsconfig.LoadDefaultConfig(
		context.TODO(),
		awsconfig.WithRegion("dev"),
		// awsconfig.WithEndpointResolverWithOptions(customResolver),
		awsconfig.WithCredentialsProvider(aws.CredentialsProvider(statusCredentialProvider)),
	)
	if err != nil {
		return err
	}

	if S3Client = s3.NewFromConfig(
		cfg,
		s3.WithEndpointResolverV2(endpoint[s3.EndpointParameters](opts.Host)),
	); S3Client == nil {
		return errors.New("aws client nil")
	}

	return nil
}

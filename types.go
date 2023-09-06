package bdd

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/viper"
)

type TestSuiteHookFunc func()

type S3 interface {
	s3.ListObjectsV2APIClient
	s3.HeadObjectAPIClient

	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObjects(context.Context, *s3.DeleteObjectsInput, ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
}

type StepAdder interface {
	Step(expr interface{}, stepFunc interface{})
}

type TestCustomStepFunc func(sm StepAdder)

type TestCustomContextFunc func() any

type ViperConfigFunc func(v *viper.Viper)

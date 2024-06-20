package bdd

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/viper"
	"github.com/tigh-latte/go-bdd/bddcontext"
)

type TestSuiteHookFunc func() error

type S3 interface {
	s3.ListObjectsV2APIClient
	s3.HeadObjectAPIClient

	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObjects(context.Context, *s3.DeleteObjectsInput, ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
}

type StepAdder interface {
	Step(expr interface{}, stepFunc interface{})
	Given(expr interface{}, stepFunc interface{})
}

type TestCustomStepFunc func(sm StepAdder)

type TestCustomContextFunc func() any

type ViperConfigFunc func(v *viper.Viper)

type TemplateValue string

func (t TemplateValue) Render(ctx context.Context) (string, error) {
	c := bddcontext.LoadContext(ctx)
	tmpl, err := c.Template.Parse(string(t))
	if err != nil {
		return string(t), fmt.Errorf("failed to parse template for value %q: %w", t, err)
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, c.TemplateValues); err != nil {
		return string(t), fmt.Errorf("failed to execute template for value %q: %w", t, err)
	}

	return buf.String(), nil
}

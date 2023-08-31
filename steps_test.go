package bdd_test

import (
	"context"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v21"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/tigh-latte/go-bdd"
	"github.com/tigh-latte/go-bdd/bddcontext"
	"github.com/tigh-latte/go-bdd/internal/data"
	"github.com/tigh-latte/go-bdd/mocks"
)

func Test_IPutFilesIntoS3(t *testing.T) {
	tests := map[string]struct {
		tbl *godog.Table

		testDataFunc  func(...string) ([]byte, error)
		testFS        fs.FS
		putObjectFunc func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)

		expParts func() []string
		expErr   error
	}{
		"object successfully put": {
			tbl: &godog.Table{Rows: []*messages.PickleTableRow{{}, {
				Cells: []*messages.PickleTableCell{{Value: "mybucket"}, {Value: "myfile"}, {Value: "mykey"}},
			}}},
			testFS: fstest.MapFS{
				"myfile": {
					Data: []byte("hello"),
				},
			},
			expParts: func() []string {
				return []string{"myfile"}
			},
		},
		"multiple objects succesfully put": {
			tbl: &godog.Table{Rows: []*messages.PickleTableRow{{}, {
				Cells: []*messages.PickleTableCell{{Value: "myfirstbucket"}, {Value: "myfirstfile"}, {Value: "myfirstkey"}},
			}, {
				Cells: []*messages.PickleTableCell{{Value: "mysecondbucket"}, {Value: "mysecondfile"}, {Value: "mysecondkey"}},
			}}},
			testFS: fstest.MapFS{
				"myfirstfile": {
					Data: []byte("hello 0"),
				},
				"mysecondfile": {
					Data: []byte("hello 1"),
				},
			},
			expParts: func() func() []string {
				var idx int
				return func() []string {
					defer func() { idx++ }()
					return [][]string{
						{"myfirstfile"},
						{"mysecondfile"},
					}[idx]
				}
			}(),
		},
		"no object returns fine": {
			tbl: &godog.Table{Rows: []*messages.PickleTableRow{{}}},
		},
		"error on put is handled": {
			tbl: &godog.Table{Rows: []*messages.PickleTableRow{{}, {
				Cells: []*messages.PickleTableCell{{Value: "mybucket"}, {Value: "myfile"}, {Value: "mykey"}},
			}}},
			testFS: fstest.MapFS{
				"myfile": {
					Data: []byte("hello"),
				},
			},
			putObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
				return nil, errors.New("oh wow")
			},
			expParts: func() []string { return []string{"myfile"} },
			expErr:   errors.New("failed to put file 'myfile': oh wow"),
		},
	}

	for name, test := range tests {
		test := test

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := bddcontext.WithContext(context.Background(), &bddcontext.Context{
				TestData: &data.DataDir{
					Prefix: "",
					FS:     test.testFS,
				},
				S3Client: &mocks.S3Mock{
					PutObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
						if test.putObjectFunc != nil {
							return test.putObjectFunc(ctx, params, optFns...)
						}
						return &s3.PutObjectOutput{}, nil
					},
				},
			})
			err := bdd.IPutFilesIntoS3(ctx, test.tbl)

			if test.expErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, test.expErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

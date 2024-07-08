package bdd_test

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"regexp"
	"testing"
	"testing/fstest"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v21"
	"github.com/stretchr/testify/assert"
	"github.com/tigh-latte/go-bdd"
	"github.com/tigh-latte/go-bdd/bddcontext"
	"github.com/tigh-latte/go-bdd/internal/data"
	"github.com/tigh-latte/go-bdd/mocks"
)

type adder struct {
	steps []*regexp.Regexp
}

func (a *adder) Step(e any, _ any) {
	expr := e.(string)
	a.steps = append(a.steps, regexp.MustCompile(expr))
}

func (a *adder) Given(e any, _ any) {
	expr := e.(string)
	a.steps = append(a.steps, regexp.MustCompile(expr))
}

func TestISendARequestToRegex(t *testing.T) {
	tests := map[string]struct {
		sentence string

		expVerb string
		expHost string
		expPort string
		expPath string
	}{
		"GET full url": {
			sentence: `I send a GET request to "wow.test.localhost:8000/api/v1/holyhell"`,
			expVerb:  http.MethodGet,
			expPath:  "/api/v1/holyhell",
			expPort:  ":8000",
			expHost:  "wow.test.localhost",
		},
		"GET port:path only": {
			sentence: `I send a GET request to ":6000/api/v1/holyhell"`,
			expVerb:  http.MethodGet,
			expPort:  ":6000",
			expPath:  "/api/v1/holyhell",
		},
		"GET host/path only": {
			sentence: `I send a GET request to "wow.test.localhost/api/v1/holyhell"`,
			expVerb:  http.MethodGet,
			expHost:  "wow.test.localhost",
			expPath:  "/api/v1/holyhell",
		},
		"GET path only": {
			sentence: `I send a GET request to "/api/v1/holyhell"`,
			expVerb:  http.MethodGet,
			expPath:  "/api/v1/holyhell",
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			adder := &adder{
				steps: make([]*regexp.Regexp, 0),
			}
			bdd.InitSteps(adder)

			var matches [][]string
			for _, step := range adder.steps {
				matches = step.FindAllStringSubmatch(test.sentence, -1)
				if len(matches) > 0 {
					break
				}
			}
			if len(matches) == 0 {
				t.Fatalf("no matches: %q", test.sentence)
			}

			match := matches[0]

			var (
				verb = match[1]
				host = match[2]
				port = match[3]
				path = match[4]
			)

			if test.expVerb != verb {
				t.Fatalf("wrong verb. want=%q got=%q", test.expVerb, verb)
			}
			if test.expHost != host {
				t.Fatalf("wrong host. want=%q got=%q", test.expHost, host)
			}
			if test.expPort != port {
				t.Fatalf("wrong port. want=%q got=%q", test.expPort, port)
			}
			if test.expPath != path {
				t.Fatalf("wrong path. want=%q got=%q", test.expPath, path)
			}
		})
	}
}

func TestISendARequestToWithJSONAsStringRegex(t *testing.T) {
	tests := map[string]struct {
		sentence string

		expVerb string
		expHost string
		expPort string
		expPath string
	}{
		"GET full url": {
			sentence: `I send a GET request to "wow.test.localhost:8000/api/v1/holyhell" with JSON:`,
			expVerb:  http.MethodGet,
			expPath:  "/api/v1/holyhell",
			expPort:  ":8000",
			expHost:  "wow.test.localhost",
		},
		"GET port:path only": {
			sentence: `I send a GET request to ":6000/api/v1/holyhell" with JSON:`,
			expVerb:  http.MethodGet,
			expPort:  ":6000",
			expPath:  "/api/v1/holyhell",
		},
		"GET host/path only": {
			sentence: `I send a GET request to "wow.test.localhost/api/v1/holyhell" with JSON:`,
			expVerb:  http.MethodGet,
			expHost:  "wow.test.localhost",
			expPath:  "/api/v1/holyhell",
		},
		"GET path only": {
			sentence: `I send a GET request to "/api/v1/holyhell" with JSON:`,
			expVerb:  http.MethodGet,
			expPath:  "/api/v1/holyhell",
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			adder := &adder{
				steps: make([]*regexp.Regexp, 0),
			}
			bdd.InitSteps(adder)

			var matches [][]string
			for _, step := range adder.steps {
				matches = step.FindAllStringSubmatch(test.sentence, -1)
				if len(matches) > 0 {
					break
				}
			}
			if len(matches) == 0 {
				t.Fatalf("no matches: %q", test.sentence)
			}

			match := matches[0]

			var (
				verb = match[1]
				host = match[2]
				port = match[3]
				path = match[4]
			)

			if test.expVerb != verb {
				t.Fatalf("wrong verb. want=%q got=%q", test.expVerb, verb)
			}
			if test.expHost != host {
				t.Fatalf("wrong host. want=%q got=%q", test.expHost, host)
			}
			if test.expPort != port {
				t.Fatalf("wrong port. want=%q got=%q", test.expPort, port)
			}
			if test.expPath != path {
				t.Fatalf("wrong path. want=%q got=%q", test.expPath, path)
			}
		})
	}
}

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
				TestData: &data.Dir{
					Prefix: "",
					FS:     test.testFS,
				},
				S3: &bddcontext.S3Context{
					Client: &mocks.S3Mock{
						PutObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
							if test.putObjectFunc != nil {
								return test.putObjectFunc(ctx, params, optFns...)
							}
							return &s3.PutObjectOutput{}, nil
						},
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

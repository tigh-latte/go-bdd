package bdd

import (
	"io/fs"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/viper"
	"github.com/tigh-latte/go-bdd/clients"
	"github.com/tigh-latte/go-bdd/internal/websocket"

	"github.com/cucumber/godog"
	"github.com/tigh-latte/go-bdd/config"
	"github.com/tigh-latte/go-bdd/internal/data"
)

type TestSuiteOptionFunc func(t *testSuiteOpts)

type User struct {
	Username string
	APIKey   string
}

type testSuiteOpts struct {
	db    *dbOptions
	s3    *clients.S3Options
	sqs   *clients.SQSOptions
	mongo *clients.MongoOptions
	rmq   *rmqOptions
	grpcs []grpcOptions
	ws    *wsOptions

	concurrency int

	featureFS     fs.FS
	testDataDir   *data.DataDir
	rabbitDataDir *data.DataDir
	httpDataDir   *data.DataDir
	sqsDataDir    fs.FS
	mongoDataDir  *data.DataDir
	wsDataDir     *data.DataDir

	cookies      []*http.Cookie
	alwaysIgnore []string

	customBeforeSuiteFunc TestSuiteHookFunc
	customAfterSuiteFunc  TestSuiteHookFunc

	customBeforeScenarioFunc godog.BeforeScenarioHook
	customAfterScenarioFunc  godog.AfterScenarioHook

	customBeforeStepFunc godog.BeforeStepHook
	customAfterStepFunc  godog.AfterStepHook

	customStepFunc      []TestCustomStepFunc
	customRequireFuncs  RequireFuncs
	customTemplateFuncs template.FuncMap

	customViperConfigFunc ViperConfigFunc
}

func (o *testSuiteOpts) applyConfig() {
	// Apply mongo url
	mongoURI := viper.GetString("mongo.uri")
	if mongoURI != "" {
		o.mongo = &clients.MongoOptions{
			URI: mongoURI,
		}
	}

	for _, cookie := range config.Cookies() {
		parts := strings.SplitN(cookie, "=", 2)
		o.cookies = append(o.cookies, &http.Cookie{
			Name:    parts[0],
			Value:   parts[1],
			Expires: time.Now().Add(time.Minute * 60),
		})
	}
}

type grpcOptions struct {
	Host string
}

type dbOptions struct{}

type rmqOptions struct {
	Host            string
	DeliveryTimeout time.Duration
	Subs            []RabbitMQSubscription
}

type RabbitMQSubscription struct {
	Exchange   string
	RoutingKey string
}

type wsOptions struct {
	Timeout time.Duration
	Client  websocket.Client
}

func WithS3(host, key, secret string) TestSuiteOptionFunc {
	return func(t *testSuiteOpts) {
		t.s3 = &clients.S3Options{
			Host:   host,
			Key:    key,
			Secret: secret,
		}
	}
}

func WithSQS(host, key, secret string) TestSuiteOptionFunc {
	return func(t *testSuiteOpts) {
		t.sqs = &clients.SQSOptions{
			Host:   host,
			Key:    key,
			Secret: secret,
		}
	}
}

// WithSQSTestData takes an `fs.FS` of which to retrieve sqs message bodies from.
// This function assumes the data will be in a directory titled `sqs`.
//
// Usage example (using `embed.FS`):
//
//	//go:embed sqs/*
//	var sqsData embed.FS
//	. . .
//	cucumber.NewSuite("test", cucumber.WithSQSTestData(sqsData))
func WithSQSTestData(fsys fs.FS) TestSuiteOptionFunc {
	return func(t *testSuiteOpts) {
		t.sqsDataDir = &data.DataDir{
			FS:     fsys,
			Prefix: "sqs",
		}
	}
}

func WithMongo(uri string) TestSuiteOptionFunc {
	return func(t *testSuiteOpts) {
		t.mongo = &clients.MongoOptions{
			URI: uri,
		}
	}
}

func WithRabbitMQ(host string, deliveryTimeout time.Duration, subs ...RabbitMQSubscription) TestSuiteOptionFunc {
	return func(t *testSuiteOpts) {
		t.rmq = &rmqOptions{
			Host:            host,
			DeliveryTimeout: deliveryTimeout,
			Subs:            subs,
		}
	}
}

func WithFeatureFS(fsys fs.FS) TestSuiteOptionFunc {
	return func(t *testSuiteOpts) {
		t.featureFS = fsys
	}
}

// WithRabbitMQTestData takes an `fs.FS` of which to retrieve rabbitmq message bodies from.
// This function assumes the data will be in a directory titled `rabbitmq`.
//
// Usage example (using `embed.FS`):
//
//	//go:embed rabbitmq/*
//	var rabbitData embed.FS
//	. . .
//	cucumber.NewSuite("test", cucumber.WithRabbitMQTestData(rabbitData))
func WithRabbitMQTestData(fsys fs.FS) TestSuiteOptionFunc {
	return func(t *testSuiteOpts) {
		t.rabbitDataDir = &data.DataDir{
			FS:     fsys,
			Prefix: "rabbitmq",
		}
	}
}

// WithGRPC FUNCTIONALITY NOT IMPLEMENTED YET.
func WithGRPC(host string) TestSuiteOptionFunc {
	return func(t *testSuiteOpts) {
		t.grpcs = append(t.grpcs, grpcOptions{
			Host: host,
		})
	}
}

func WithWebsockets(timeout time.Duration) TestSuiteOptionFunc {
	return func(t *testSuiteOpts) {
		t.ws = &wsOptions{
			Timeout: timeout,
			Client:  websocket.NewClient(),
		}
	}
}

// WithWebsocketsTestData takes an `fs.FS` of which to retrieve websocket message bodies from.
// This function assumes the data will be in a directory titled `websockets`.
//
// Usage example (using `embed.FS`):
//
//	//go:embed websockets/*
//	var websocketData embed.FS
//	. . .
//	cucumber.NewSuite("test", cucumber.WithWebsocketsTestData(websocketData))
func WithWebsocketsTestData(fs fs.FS) TestSuiteOptionFunc {
	return func(t *testSuiteOpts) {
		t.wsDataDir = &data.DataDir{
			FS:     fs,
			Prefix: "websockets",
		}
	}
}

// WithBeforeSuite executes the provided function before suite execution.
//
// Usage example:
//
//	cucumber.NewSuite("test", cucumber.WithBeforeSuite(func() {
//		// init custom service or something
//	}))
func WithBeforeSuite(fn TestSuiteHookFunc) TestSuiteOptionFunc {
	return func(t *testSuiteOpts) {
		t.customBeforeSuiteFunc = fn
	}
}

// WithAfterSuite executes the provided function before suite execution.
//
// Usage example:
//
//	cucumber.NewSuite("test", cucumber.WithAfterSuite(func() {
//		// cleanup custom services or something
//	}))
func WithAfterSuite(fn TestSuiteHookFunc) TestSuiteOptionFunc {
	return func(t *testSuiteOpts) {
		t.customAfterSuiteFunc = fn
	}
}

// WithBeforeScenario executes the provided function before each scenario execution, providing `cucumber.TestContext`
// and `godog.Scenario`.
//
// Usage example:
//
//	cucumber.NewSuite("test", cucumber.WithBeforeScenario(func(ctx *cucumber.TestContext, sn *godog.Scenario) {
//		// Set a header for all requests
//		ctx.WithHeaders(...)
//	}))
func WithBeforeScenario(fn godog.BeforeScenarioHook) TestSuiteOptionFunc {
	return func(t *testSuiteOpts) {
		t.customBeforeScenarioFunc = fn
	}
}

// WithAfterScenario executes the provided function after each scenario execution, providing `cucumber.TestContext`,
// `godog.Scenario`, and an `error` if one occured.
//
// Usage example:
//
//	cucumber.NewSuite("test", cucumber.WithAfterScenario(func(ctx *cucumber.TestContext, sn *godog.Scenario, err error) {
//		// cleanup scenario test data
//	}))
func WithAfterScenario(fn godog.AfterScenarioHook) TestSuiteOptionFunc {
	return func(t *testSuiteOpts) {
		t.customAfterScenarioFunc = fn
	}
}

// WithAfterStep executes the provided function before every step, providing the `cucumber.TestContext`,
// and the `godog.Step`.
//
// Usage example:
//
//	cucumber.NewSuite("test", cucumber.WithBeforeStep(func(ctx *cucumber.TestContext, step *godog.Step) {
//		// prep ctx or something
//	}))
func WithBeforeStep(fn godog.BeforeStepHook) TestSuiteOptionFunc {
	return func(t *testSuiteOpts) {
		t.customBeforeStepFunc = fn
	}
}

// WithAfterStep executes the provided function after every step, providing the `cucumber.TestContext`,
// `godog.Step`, and any `error` that occured during the step.
//
// Usage example:
//
//	cucumber.NewSuite("test", cucumber.WithAfterStep(func(ctx *cucumber.TestContext, step *godog.Step, err error) {
//		// cleandown ctx or something
//	}))
func WithAfterStep(fn godog.AfterStepHook) TestSuiteOptionFunc {
	return func(t *testSuiteOpts) {
		t.customAfterStepFunc = fn
	}
}

// WithCustomSteps takes a function which loads custom steps.
// Use this for steps sepcific to the implementing service, or for
// utility functions.
//
// Usage example:
//
//	cucumber.NewSuite("test", cucumber.WithCustomSteps(func(ctx *cucumber.TestContext, sa cucumber.StepAdder) {
//	    sa.Step(`I say "([^"]*)"`, iSay(ctx))
//	}))
//	. . .
//	func iSay(ctx *cucumber.TestContext) func(string) error {
//	    return func(sentence string) error {
//	        fmt.Println(sentence)
//	        return nil
//	    }
//	}
func WithCustomSteps(fns ...TestCustomStepFunc) TestSuiteOptionFunc {
	return func(t *testSuiteOpts) {
		t.customStepFunc = append(t.customStepFunc, fns...)
	}
}

// WithTestData takes an `fs.FS` of which to retrieve testing data from.
// This function assumes the data will be in a directory titled `testdata`.
//
// Usage example (using `embed.FS`):
//
//	//go:embed testdata/*
//	var testData embed.FS
//	. . .
//	cucumber.NewSuite("test", cucumber.WithTestData(testData))
func WithTestData(fsys fs.FS) TestSuiteOptionFunc {
	return func(t *testSuiteOpts) {
		t.testDataDir = &data.DataDir{
			FS:     fsys,
			Prefix: "testdata",
		}
	}
}

// WithHTTPData takes an `fs.FS` of which to retrieve http requests and responses.
// This function assumes the data will be in a directory titled `http`.
//
// Usage example (using `embed.FS`):
//
//	//go:embed http/*
//	var httpData embed.FS
//	. . .
//	cucumber.NewSuite("test", cucumber.WithHTTPData(httpData))
func WithHTTPData(fsys fs.FS) TestSuiteOptionFunc {
	return func(t *testSuiteOpts) {
		t.httpDataDir = &data.DataDir{
			FS:     fsys,
			Prefix: "http",
		}
	}
}

// WithMongoData takes an `fs.FS` of which to retrieve mongo document data.
// This function assumes the data will be in a directory titled `mongo`.
//
// Usage example (using `embed.FS`):
//
//	//go:embed mongo/*
//	var mongoData embed.FS
//	. . .
//	cucumber.NewSuite("test", cucumber.WithMongoData(httpData))
func WithMongoData(fsys fs.FS) TestSuiteOptionFunc {
	return func(t *testSuiteOpts) {
		t.mongoDataDir = &data.DataDir{
			FS:     fsys,
			Prefix: "mongo",
		}
	}
}

// WithCustomRequireFuncs add custom funcs to be executed via `@require=` tags,
// before a scenario is ran.
//
// Usage example:
//
//	cucumber.NewSuite("test", cucumber.WithCustomRequireFuncs(cucumber.RequireFuncs{
//		"customrequire": func(ctx *cucumber.TextContext) error { return nil },
//	}))
//
// @v3 @get @require=customrequire
// Scenario: Showcase custom require
func WithCustomRequireFuncs(fns RequireFuncs) TestSuiteOptionFunc {
	return func(t *testSuiteOpts) {
		for k, v := range fns {
			t.customRequireFuncs[k] = v
		}
	}
}

// WithAlwaysIgnoreFromResponse specific a list of keys via jsonpath to be
// ignored from every http response.
// Typically used for data that is completely non-deterministic data.
//
// Usage example:
//
//	cucumber.NewSuite("test", cucumber.WithAlwaysIgnoreFromResponse(
//		"..createdAt",
//		"..updatedAt",
//	))
func WithAlwaysIgnoreFromResponse(paths ...string) TestSuiteOptionFunc {
	return func(t *testSuiteOpts) {
		t.alwaysIgnore = append(t.alwaysIgnore, paths...)
	}
}

// WithCustomTemplateFuncs add custom functions to the template engine.
//
// Usage example:
//
//	cucumber.NewSuite("test", cucumber.WithCustomTemplateFuncs(template.FuncMap{
//		"myfunc": func() string { return "holy hell" },
//	}))
//
//	GET-test.json:
//	{ "wow": "{{ myfunc }}" }
func WithCustomTemplateFuncs(fns template.FuncMap) TestSuiteOptionFunc {
	return func(t *testSuiteOpts) {
		for k, v := range fns {
			t.customTemplateFuncs[k] = v
		}
	}
}

func WithViperConfigFunc(fn ViperConfigFunc) TestSuiteOptionFunc {
	return func(t *testSuiteOpts) {
		t.customViperConfigFunc = fn
	}
}

package bdd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/google/uuid"
	"github.com/makiuchi-d/gozxing"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/tigh-latte/go-bdd/bddcontext"
	"github.com/tigh-latte/go-bdd/clients"
	"github.com/tigh-latte/go-bdd/config"
	"github.com/tigh-latte/go-bdd/fake"
	configinternal "github.com/tigh-latte/go-bdd/internal/config"
	"github.com/zeroflucs-given/generics/collections/stack"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var godogOpts = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "progress",
	Paths:  []string{"features"},
	Strict: true,
}

func init() {
	godog.BindCommandLineFlags("godog.", &godogOpts)

	configinternal.BindFlags()

	pflag.Parse()
}

type Suite struct {
	suite  godog.TestSuite
	reqFns RequireFuncs
}

func NewSuite(name string, oo ...TestSuiteOptionFunc) *Suite {
	godogOpts.Paths = pflag.Args()

	opts := &testSuiteOpts{
		customBeforeSuiteFunc: noopTestSuiteHook,
		customAfterSuiteFunc:  noopTestSuiteHook,

		customBeforeScenarioFunc: noopBeforeScenarioHook,
		customAfterScenarioFunc:  noopAfterScenarioHook,

		customBeforeStepFunc: noopTestBeforeStepHook,
		customAfterStepFunc:  noopTestAfterStepHook,

		customStepFunc:        []TestCustomStepFunc{},
		customRequireFuncs:    RequireFuncs{},
		customTemplateFuncs:   make(template.FuncMap),
		customViperConfigFunc: func(v *viper.Viper) {},

		cookies:           make([]*http.Cookie, 0),
		alwaysIgnore:      make([]string, 0),
		globalHTTPHeaders: make(map[string][]string, 0),
	}
	for _, o := range oo {
		o(opts)
	}

	s := &Suite{
		reqFns: make(RequireFuncs, 0),
		suite: godog.TestSuite{
			Name:    name,
			Options: &godogOpts,
		},
	}
	s.suite.TestSuiteInitializer = s.initSuite(opts)
	s.suite.ScenarioInitializer = s.initScenario(opts)
	godogOpts.FS = opts.featureFS

	return s
}

func (s *Suite) Run() int {
	return s.suite.Run()
}

func (s *Suite) initSuite(opts *testSuiteOpts) func(ctx *godog.TestSuiteContext) {
	return func(ctx *godog.TestSuiteContext) {
		// Init config
		ctx.BeforeSuite(func() {
			config.Init()
			opts.customViperConfigFunc(viper.GetViper())
			opts.applyConfig()
		})

		var (
			comp compose.ComposeStack
			err  error
		)
		ctx.BeforeSuite(func() {
			if len(opts.dockerComposeOptions.paths) == 0 {
				return
			}
			comp, err = compose.NewDockerComposeWith(
				compose.WithStackFiles(opts.dockerComposeOptions.paths...),
				compose.StackIdentifier(strconv.FormatInt(time.Now().Unix(), 10)),
			)
			if err != nil {
				panic(err)
			}
			if err = comp.WithEnv(opts.dockerComposeOptions.env).Up(context.TODO()); err != nil {
				panic(err)
			}
			clients.ComposeStack = comp
		})

		ctx.BeforeSuite(func() {
			if err = clients.InitS3(opts.s3); err != nil {
				panic(err)
			}
			if err = clients.InitMongo(opts.mongo); err != nil {
				panic(err)
			}
			if err = clients.InitSQS(opts.sqs); err != nil {
				panic(err)
			}
			if err = clients.InitDynamoDB(opts.dynamodb); err != nil {
				panic(err)
			}
		})

		ctx.BeforeSuite(func() {
			for k, v := range opts.customRequireFuncs {
				if _, ok := s.reqFns[k]; ok {
					panic("cannot overwrite builtin require function")
				}
				s.reqFns[k] = v
			}
		})

		ctx.BeforeSuite(func() {
			if err = opts.customBeforeSuiteFunc(); err != nil {
				panic(fmt.Errorf("failed before suite hook: %w", err))
			}
		})

		ctx.AfterSuite(func() {
			if err = opts.customAfterSuiteFunc(); err != nil {
				panic(fmt.Errorf("failed after suite hook: %w", err))
			}
		})

		ctx.AfterSuite(func() {
			if comp == nil {
				return
			}
			if err = comp.Down(context.TODO()); err != nil {
				panic(err)
			}
		})
	}
}

func (s *Suite) initScenario(opts *testSuiteOpts) func(ctx *godog.ScenarioContext) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return func(ctx *godog.ScenarioContext) {
		sd := &bddcontext.Context{
			TemplateValues: make(map[string]any),
			QRCodes:        stack.NewStack[*gozxing.Result](20),
			MongoContext: &bddcontext.MongoContext{
				IDs:           stack.NewStack[primitive.ObjectID](20),
				DocumentIDMap: make(map[primitive.ObjectID]string, 0),
				ToIgnore:      make([]string, 0),
				TestData:      opts.mongoDataDir,
				Client:        clients.MongoClient,
			},
			Compose: &bddcontext.ComposeContext{
				Stack: clients.ComposeStack,
			},
			SQS: &bddcontext.SQSContext{
				MsgAttrs:   make(map[string]sqstypes.MessageAttributeValue),
				MessageIDs: stack.NewStack[string](20),
				TestData:   opts.sqsDataDir,
				Client:     clients.SQSClient,
			},
			DynamoDB: &bddcontext.DynamoDBContext{
				Client:   clients.DynamoDBClient,
				TestData: opts.dynamoDataDir,
			},
			HTTP: &bddcontext.HTTPContext{
				Headers:       make(http.Header, 0),
				Cookies:       make([]*http.Cookie, len(opts.cookies)),
				Requests:      stack.NewStack[json.RawMessage](20),
				Responses:     stack.NewStack[json.RawMessage](20),
				ResponseCodes: stack.NewStack[int](20),
				TestData:      opts.httpDataDir,
				QueryParams:   make(url.Values),
				ToIgnore:      make([]string, 0),
				Client:        &http.Client{Timeout: 30 * time.Second, Transport: transport},
				GlobalHeaders: make(map[string][]string, len(opts.globalHTTPHeaders)),
			},
			TestData:     opts.testDataDir,
			IgnoreAlways: make([]string, len(opts.alwaysIgnore)),
		}

		ctx.Before(func(ctx context.Context, sn *godog.Scenario) (context.Context, error) {
			sd.ID = sn.Id
			sd.Template = template.New(sd.ID)
			scenarioStart := time.Now().UTC()

			today := time.Date(
				scenarioStart.Year(),
				scenarioStart.Month(),
				scenarioStart.Day(),
				0, 0, 0, 0,
				scenarioStart.Location(),
			)
			yesterday := time.Date(
				scenarioStart.Year(),
				scenarioStart.Month(),
				scenarioStart.Day()-1,
				0, 0, 0, 0,
				scenarioStart.Location(),
			)
			tomorrow := time.Date(scenarioStart.Year(),
				scenarioStart.Month(),
				scenarioStart.Day()+1,
				0, 0, 0, 0,
				scenarioStart.Location(),
			)

			sd.ScenarioStart = scenarioStart
			sd.TemplateValues["__scenario_id"] = sn.Id
			sd.TemplateValues["__time_unix"] = sd.ScenarioStart.Unix()
			sd.TemplateValues["__time_unix_milli"] = sd.ScenarioStart.UnixMilli()
			sd.TemplateValues["__now"] = scenarioStart.String()
			sd.TemplateValues["__today"] = today.Local().Format("2006-01-02")
			sd.TemplateValues["__today_timestamp"] = today.Local().String()
			sd.TemplateValues["__yesterday"] = yesterday.Local().Format("2006-01-02")
			sd.TemplateValues["__yesterday_timestamp"] = yesterday.Local().String()
			sd.TemplateValues["__tomorrow"] = tomorrow.Local().Format("2006-01-02")
			sd.TemplateValues["__tomorrow_timestamp"] = tomorrow.Local().String()

			sd.Template.Funcs(sprig.TxtFuncMap())
			sd.Template.Funcs(template.FuncMap{
				"scenarioStart":   func() time.Time { return sd.ScenarioStart },
				"stepStart":       func() time.Time { return sd.StepStart },
				"unixEpochMillis": func(t time.Time) int64 { return t.UnixMilli() },
				"yearDay":         func(t time.Time) int { return t.YearDay() },
				"add": func(l, r any) int {
					left := toInt(l)
					right := toInt(r)
					return left + right
				},
				"sub": func(l, r any) int {
					left := toInt(l)
					right := toInt(r)
					return left - right
				},
				"assertFuture":     assertFuture,
				"assertJsonString": assertJsonString,
				"assertNotEmpty":   assertNotEmpty,
				"toInt":            toInt,
				"match":            match,
				"urlEncode":        urlEncode,
				"toString":         toString,
				"toJsonString":     toJsonString,
				"viper": func(key string) string {
					return viper.GetString(key)
				},
				"toHostname":      toHostname,
				"randomName":      fake.Name,
				"randomFirstName": fake.FirstName,
				"randomLastName":  fake.LastName,
				"randomEmail":     fake.Email,
				"randomSentence":  fake.Sentence,
				"randomString": func(l int) string {
					const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"
					var b strings.Builder
					b.Grow(l)
					return b.String()
				},
				"upper": strings.ToUpper,
				"lower": strings.ToLower,
				"uuid": func() string {
					return uuid.New().String()
				},
			})
			sd.Template.Funcs(opts.customTemplateFuncs)

			if opts.s3 != nil {
				sd.S3 = &bddcontext.S3Context{
					Client: clients.S3Client,
				}
			}
			if opts.ws != nil {
				sd.WS = &bddcontext.WebsocketContext{
					Host:    "TODO",
					Timeout: opts.ws.Timeout,
					Connections: make(map[string]struct {
						SessionID string
						Messages  *stack.Stack[[]byte]
					}, 0),
					TestData: opts.wsDataDir,
					Client:   opts.ws.Client,
				}
			}

			return bddcontext.WithContext(ctx, sd), nil
		})

		// copy gobal data into scenario context.
		ctx.Before(func(ctx context.Context, sn *godog.Scenario) (context.Context, error) {
			bctx := bddcontext.LoadContext(ctx)
			copy(bctx.HTTP.Cookies, opts.cookies)
			copy(bctx.IgnoreAlways, opts.alwaysIgnore)
			maps.Copy(bctx.HTTP.GlobalHeaders, opts.globalHTTPHeaders)
			return bddcontext.WithContext(ctx, bctx), nil
		})

		ctx.Before(func(ctx context.Context, sn *godog.Scenario) (context.Context, error) {
			var err error
			for _, tag := range sn.Tags {
				fnName, ok := strings.CutPrefix(tag.Name, TagRequire)
				if !ok {
					continue
				}
				fn, ok := s.reqFns[fnName]
				if !ok {
					return ctx, errors.New("unregistered required func")
				}
				if ctx, err = fn(ctx); err != nil {
					return ctx, err
				}
			}
			return ctx, nil
		})
		ctx.Before(func(ctx context.Context, sn *godog.Scenario) (context.Context, error) {
			return opts.customBeforeScenarioFunc(ctx, sn)
		})

		ctx.StepContext().Before(func(ctx context.Context, st *godog.Step) (context.Context, error) {
			// Whatever we want to do can be added here.
			sd.StepStart = time.Now().UTC()
			return ctx, nil
		})
		ctx.StepContext().Before(func(ctx context.Context, st *godog.Step) (context.Context, error) {
			return opts.customBeforeStepFunc(ctx, st)
		})

		ctx.StepContext().After(func(ctx context.Context, st *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {
			return opts.customAfterStepFunc(ctx, st, status, err)
		})

		ctx.StepContext().After(func(ctx context.Context, st *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {
			// Whatever we want to do can be added here.
			return ctx, nil
		})
		ctx.After(func(ctx context.Context, sn *godog.Scenario, err error) (context.Context, error) {
			return opts.customAfterScenarioFunc(ctx, sn, err)
		})
		ctx.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
			if err == nil {
				return ctx, nil
			}

			fmt.Printf("fake.Seed=%s\n", fake.GetInfo())

			b := bddcontext.LoadContext(ctx)
			return ctx, b.Compose.Stack.Down(ctx)
		})
		ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
			if opts.ws != nil {
				for _, connection := range sd.WS.Connections {
					sd.WS.Client.Close(ctx, connection.SessionID)
				}
			}

			return ctx, nil
		})

		InitSteps(ctx)

		for _, fn := range opts.customStepFunc {
			fn(ctx)
		}
	}
}

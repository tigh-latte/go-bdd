package bddcontext

import (
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/url"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/makiuchi-d/gozxing"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/tigh-latte/go-bdd/internal/websocket"
	"github.com/zeroflucs-given/generics/collections/stack"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type contextKey struct{}

func WithContext(parent context.Context, bctx *Context) context.Context {
	return context.WithValue(parent, contextKey{}, bctx)
}

func LoadContext(ctx context.Context) *Context {
	return ctx.Value(contextKey{}).(*Context)
}

type Context struct {
	ID string

	TestID         string
	ScenarioStart  time.Time
	StepStart      time.Time
	TemplateValues map[string]any
	Template       *template.Template

	TestData fs.FS

	HTTP *HTTPContext

	SQS *SQSContext

	DynamoDB *DynamoDBContext

	S3 *S3Context

	Compose *ComposeContext

	QRCodes *stack.Stack[*gozxing.Result]

	WS *WebsocketContext

	MongoContext *MongoContext

	IgnoreAlways []string
}

type ComposeContext struct {
	Stack compose.ComposeStack
}

type S3Context struct {
	Client interface {
		s3.ListObjectsV2APIClient
		s3.HeadObjectAPIClient

		PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
		DeleteObjects(context.Context, *s3.DeleteObjectsInput, ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
	}
}

type HTTPContext struct {
	Endpoint      string
	GlobalHeaders map[string][]string
	Headers       http.Header
	Cookies       []*http.Cookie
	QueryParams   url.Values
	ToIgnore      []string

	Requests      *stack.Stack[json.RawMessage]
	Responses     *stack.Stack[json.RawMessage]
	ResponseCodes *stack.Stack[int]

	TestData fs.FS

	Client *http.Client
}

type SQSContext struct {
	MsgAttrs map[string]sqstypes.MessageAttributeValue

	MessageIDs *stack.Stack[string]

	Client interface {
		GetQueueUrl(ctx context.Context, params *sqs.GetQueueUrlInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error)
		SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	}

	TestData fs.FS
}

type DynamoDBContext struct {
	Client   *dynamodb.Client
	TestData fs.FS
}

type MongoContext struct {
	DocumentIDMap map[primitive.ObjectID]string
	IDs           *stack.Stack[primitive.ObjectID]
	ToIgnore      []string

	TestData fs.FS

	Client *mongo.Client
}

type WebsocketContext struct {
	Host    string
	Timeout time.Duration

	Connections map[string]struct {
		SessionID string
		Messages  *stack.Stack[[]byte]
	}

	TestData fs.FS

	Client websocket.Client
}

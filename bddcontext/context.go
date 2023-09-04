package bddcontext

import (
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/url"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/tigh-latte/go-bdd/internal/websocket"
	"github.com/zeroflucs-given/generics/collections/stack"
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
	Time           time.Time
	TemplateValues map[string]any
	Template       *template.Template

	TestData fs.FS

	HTTP *HTTPContext

	S3Client interface {
		s3.ListObjectsV2APIClient
		s3.HeadObjectAPIClient

		PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
		DeleteObjects(context.Context, *s3.DeleteObjectsInput, ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
	}

	WS *WebsocketContext

	IgnoreAlways []string
}

type HTTPContext struct {
	Endpoint    string
	Headers     http.Header
	Cookies     []*http.Cookie
	QueryParams url.Values
	ToIgnore    []string

	Requests      *stack.Stack[json.RawMessage]
	Responses     *stack.Stack[json.RawMessage]
	ResponseCodes *stack.Stack[int]

	TestData fs.FS

	Client *http.Client
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

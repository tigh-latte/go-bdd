package noop

import (
	"context"
	"github.com/tigh-latte/go-bdd/internal/websocket"
)

type WebsocketClientMock struct{}

func (w *WebsocketClientMock) Dial(ctx context.Context, host string) (string, error) {
	return "noop", nil
}

func (w *WebsocketClientMock) Read(ctx context.Context, sid string, event string) (<-chan websocket.Event, error) {
	return nil, nil
}

func (w *WebsocketClientMock) Write(ctx context.Context, sid string, event string, message []byte) error {
	return nil
}

func (w *WebsocketClientMock) Close(ctx context.Context, sid string) error {
	return nil
}

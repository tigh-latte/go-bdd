package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	gosocketio "github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
	"github.com/pkg/errors"
	"github.com/tigh-latte/go-bdd/internal/syncmap"
)

const (
	EventNameAuthResult  string = "authResult"
	EventNameEventNotice string = "eventNotice"
)

type Client interface {
	Dial(ctx context.Context, host string) (string, error)
	Read(ctx context.Context, sid string, event string) (<-chan Event, error)
	Write(ctx context.Context, sid string, event string, message []byte) error
	Close(ctx context.Context, sid string) error
}

func NewClient() Client {
	return &client{
		events:      syncmap.NewMap[string, chan Event](),
		connections: syncmap.NewMap[string, *gosocketio.Client](),
	}
}

type Event struct {
	Name    string
	Message map[string]any
}

type client struct {
	events      syncmap.Map[string, chan Event]
	connections syncmap.Map[string, *gosocketio.Client]
}

func (c *client) Dial(ctx context.Context, host string) (string, error) {
	u, err := url.Parse(host)
	if err != nil {
		return "", err
	}

	if u.Port() == "" {
		return "", errors.New("websocket port is missing")
	}

	secure := u.Scheme == "https"
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return "", err
	}

	conn, err := gosocketio.Dial(
		gosocketio.GetUrl(u.Hostname(), port, secure),
		transport.GetDefaultWebsocketTransport(),
	)
	if err != nil {
		return "", err
	}

	if conn == nil {
		return "", errors.New("websocket connection is missing")
	}

	var sID string
	sessionID := make(chan string)

	conn.On(gosocketio.OnConnection, func(sc *gosocketio.Channel) {
		sessionID <- sc.Id()
	})

	conn.On(gosocketio.OnError, func(_ *gosocketio.Channel) {
		fmt.Println("websocket error event received")
	})

	conn.On(gosocketio.OnDisconnection, func(c *gosocketio.Channel) {})

	conn.On(EventNameAuthResult, c.handlerAuth)
	conn.On(EventNameEventNotice, c.handlerEventNotice)

	if sID == "" {
		timer := time.NewTimer(time.Second * 30)

		select {
		case sID = <-sessionID:
		case <-timer.C:
			conn.Close()
			return "", errors.New("websocket session ID retrieval timeout")
		}

		timer.Stop()
	}

	c.connections.Store(sID, conn)
	return sID, nil
}

func (c *client) Write(ctx context.Context, sid, event string, message []byte) error {
	conn, ok := c.connections.Get(sid)
	if !ok {
		return errors.New("websocket connection doesn't exist")
	}

	if !conn.IsAlive() {
		return errors.New("websocket connection is closed")
	}

	if err := conn.Emit(event, json.RawMessage(message)); err != nil {
		return err
	}

	return nil
}

func (c *client) Read(ctx context.Context, sid string, e string) (<-chan Event, error) {
	conn, ok := c.connections.Get(sid)
	if !ok {
		return nil, errors.New("websocket connection doesn't exist")
	}

	if !conn.IsAlive() {
		return nil, errors.New("websocket connection is closed")
	}

	eventName := sid + "-" + e
	if !c.events.Exists(eventName) {
		c.events.Store(eventName, make(chan Event, 100))
	}

	ch, _ := c.events.Get(eventName)
	return ch, nil
}

func (c *client) Close(ctx context.Context, sid string) error {
	conn, ok := c.connections.Get(sid)
	if !ok {
		return errors.New("websocket connection doesn't exist")
	}

	c.events.Range(func(k string, ch chan Event) bool {
		if !strings.HasPrefix(k, sid) {
			return true
		}

		for len(ch) > 0 {
			<-ch
		}

		close(ch)
		return true
	})

	conn.Close()
	return nil
}

func (c *client) handlerAuth(sc *gosocketio.Channel, msg any) {
	data, ok := msg.(map[string]any)
	if !ok {
		fmt.Printf("error with auth result response: '%v' \n", msg)
		return
	}

	authed, ok := data["authenticated"]
	if !ok {
		fmt.Printf("auth: received object does not contain 'authenticated': '%v'\n", data)
		return
	}

	if !authed.(bool) {
		fmt.Printf("auth: failed authentication")
		return
	}

	eventName := sc.Id() + "-" + EventNameAuthResult
	if !c.events.Exists(eventName) {
		c.events.Store(eventName, make(chan Event, 100))
	}

	ch, _ := c.events.Get(eventName)
	ch <- Event{
		Name:    EventNameAuthResult,
		Message: data,
	}
}

func (c *client) handlerEventNotice(sc *gosocketio.Channel, msg any) {
	data, ok := msg.(map[string]any)
	if !ok {
		fmt.Printf("error with auth result response: '%v' \n", msg)
	}

	eventInfo, ok := data["eventInfo"]
	if !ok {
		fmt.Printf("eventInfo: received object does not contain 'eventInfo': '%v'\n", data)
		return
	}

	e, ok := eventInfo.(map[string]any)["event"]
	if !ok {
		fmt.Printf("eventInfo: received object does not contain 'event': '%v'\n", data)
		return
	}

	if e.(string) == "" {
		fmt.Printf("eventInfo: received object 'event' is empty: '%v'\n", data)
		return
	}

	eventName := sc.Id() + "-" + e.(string)

	if !c.events.Exists(eventName) {
		c.events.Store(eventName, make(chan Event, 100))
	}

	ch, _ := c.events.Get(eventName)
	ch <- Event{
		Name:    e.(string),
		Message: data,
	}
}

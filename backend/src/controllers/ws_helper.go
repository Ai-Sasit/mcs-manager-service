package controllers

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	ws "github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
)

const (
	wsWriteWait      = 10 * time.Second
	wsPongWait       = 60 * time.Second
	wsPingPeriod     = (wsPongWait * 9) / 10
	wsMaxMessageSize = 65536
	wsOutboxSize     = 256
)

type wsEnvelope struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
	Ts   string `json:"ts,omitempty"`
}

func ensureWebSocketUpgrade(c fiber.Ctx) error {
	header := c.RequestCtx().Request.Header
	upgrade := strings.ToLower(string(header.Peek("Upgrade")))
	connection := strings.ToLower(string(header.Peek("Connection")))
	key := header.Peek("Sec-WebSocket-Key")

	if upgrade != "websocket" || !strings.Contains(connection, "upgrade") || len(key) == 0 {
		return c.Status(fiber.StatusUpgradeRequired).JSON(fiber.Map{
			"status":  "error",
			"message": "WebSocket upgrade required",
		})
	}
	return nil
}

type wsClient struct {
	conn   *ws.Conn
	outbox chan []byte
	done   chan struct{}
	once   sync.Once
}

func newWSClient(conn *ws.Conn) *wsClient {
	conn.SetReadLimit(wsMaxMessageSize)
	conn.SetReadDeadline(time.Now().Add(wsPongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(wsPongWait))
		return nil
	})

	client := &wsClient{
		conn:   conn,
		outbox: make(chan []byte, wsOutboxSize),
		done:   make(chan struct{}),
	}
	go client.writePump()
	return client
}

func (c *wsClient) close() {
	c.once.Do(func() {
		close(c.done)
	})
}

func (c *wsClient) wait() {
	<-c.done
}

func (c *wsClient) send(payload []byte) bool {
	select {
	case <-c.done:
		return false
	default:
	}
	select {
	case <-c.done:
		return false
	case c.outbox <- payload:
		return true
	default:
		c.close()
		return false
	}
}

func (c *wsClient) sendEnvelope(messageType string, data string) bool {
	payload, err := json.Marshal(wsEnvelope{
		Type: messageType,
		Data: data,
		Ts:   time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return false
	}
	return c.send(payload)
}

func (c *wsClient) writePump() {
	ticker := time.NewTicker(wsPingPeriod)
	defer ticker.Stop()

	for {
		select {
		case payload := <-c.outbox:
			c.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if err := c.conn.WriteMessage(ws.TextMessage, payload); err != nil {
				c.close()
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if err := c.conn.WriteMessage(ws.PingMessage, nil); err != nil {
				c.close()
				return
			}
		case <-c.done:
			c.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			_ = c.conn.WriteMessage(ws.CloseMessage, ws.FormatCloseMessage(ws.CloseNormalClosure, ""))
			return
		}
	}
}

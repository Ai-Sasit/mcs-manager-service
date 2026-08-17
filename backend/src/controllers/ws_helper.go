package controllers

import (
	"encoding/json"
	"mc-manage-backend/src/utils"
	"strings"
	"sync"
	"time"

	ws "github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
)

const (
	wsWriteWait         = 10 * time.Second
	wsEnqueueWait       = 5 * time.Second
	wsPongWait          = 60 * time.Second
	wsPingPeriod        = 25 * time.Second
	wsMaxMessageSize    = 65536
	wsOutboxSize        = 256
	wsCloseUnauthorized = 4401

	wsCloseJobComplete = 4000 // setup job finished successfully
	wsCloseJobFailed   = 4001 // setup job failed
)

type wsEnvelope struct {
	Type          string `json:"type"`
	Data          string `json:"data,omitempty"`
	Ts            string `json:"ts,omitempty"`
	StreamID      string `json:"stream_id,omitempty"`
	EventID       uint64 `json:"event_id,omitempty"`
	OldestEventID uint64 `json:"oldest_event_id,omitempty"`
	LatestEventID uint64 `json:"latest_event_id,omitempty"`
}

var upgrader = ws.FastHTTPUpgrader{
	HandshakeTimeout: wsWriteWait,
	CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
		return true
	},
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

func upgradeAuthorizedWebSocket(c fiber.Ctx, handler func(*ws.Conn)) error {
	if err := ensureWebSocketUpgrade(c); err != nil {
		return err
	}

	token := strings.Clone(c.Query("token"))
	unauthorizedReason := ""
	claims, err := utils.ValidateJWT(token)
	if token == "" || err != nil {
		unauthorizedReason = "Invalid or expired token"
	} else if state == nil || state.UserService == nil || !state.UserService.IsTokenValid(claims.Username, token) {
		unauthorizedReason = "Token has been revoked"
	}

	return upgrader.Upgrade(c.RequestCtx(), func(conn *ws.Conn) {
		if unauthorizedReason != "" {
			_ = conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			_ = conn.WriteMessage(ws.CloseMessage, ws.FormatCloseMessage(wsCloseUnauthorized, unauthorizedReason))
			_ = conn.Close()
			return
		}
		handler(conn)
	})
}

type wsWriteRequest struct {
	messageType int
	payload     []byte
	final       bool
}

type wsClient struct {
	conn      *ws.Conn
	outbox    chan wsWriteRequest
	done      chan struct{}
	writeDone chan struct{}
	once      sync.Once

	closeMu     sync.RWMutex
	closeCode   int
	closeReason string
}

func newWSClient(conn *ws.Conn) *wsClient {
	conn.SetReadLimit(wsMaxMessageSize)
	conn.SetReadDeadline(time.Now().Add(wsPongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(wsPongWait))
		return nil
	})

	client := &wsClient{
		conn:      conn,
		outbox:    make(chan wsWriteRequest, wsOutboxSize),
		done:      make(chan struct{}),
		writeDone: make(chan struct{}),
		closeCode: ws.CloseNormalClosure,
	}
	go client.writePump()
	return client
}

func (c *wsClient) close() {
	c.closeWithCode(ws.CloseNormalClosure, "")
}

func (c *wsClient) closeWithCode(code int, reason string) {
	c.once.Do(func() {
		c.closeMu.Lock()
		c.closeCode = code
		c.closeReason = reason
		c.closeMu.Unlock()
		close(c.done)
	})
}

func (c *wsClient) sendClose(code int, text string) bool {
	return c.enqueue(wsWriteRequest{
		messageType: ws.CloseMessage,
		payload:     ws.FormatCloseMessage(code, text),
		final:       true,
	})
}

func (c *wsClient) wait() {
	<-c.writeDone
}

func (c *wsClient) send(payload []byte) bool {
	return c.enqueue(wsWriteRequest{messageType: ws.TextMessage, payload: payload})
}

func (c *wsClient) enqueue(request wsWriteRequest) bool {
	select {
	case <-c.done:
		return false
	default:
	}
	timer := time.NewTimer(wsEnqueueWait)
	defer timer.Stop()
	select {
	case <-c.done:
		return false
	case c.outbox <- request:
		return true
	case <-timer.C:
		c.closeWithCode(ws.CloseTryAgainLater, "Client cannot keep up with the log stream")
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

func (c *wsClient) sendLogEvent(messageType string, event utils.LogEvent) bool {
	payload, err := json.Marshal(wsEnvelope{
		Type:     messageType,
		Data:     event.Data,
		Ts:       event.Ts,
		StreamID: event.StreamID,
		EventID:  event.EventID,
	})
	if err != nil {
		return false
	}
	return c.send(payload)
}

func (c *wsClient) sendStreamControl(messageType string, replay utils.LogReplay, data string) bool {
	payload, err := json.Marshal(wsEnvelope{
		Type:          messageType,
		Data:          data,
		Ts:            time.Now().UTC().Format(time.RFC3339Nano),
		StreamID:      replay.StreamID,
		OldestEventID: replay.OldestEventID,
		LatestEventID: replay.LatestEventID,
	})
	if err != nil {
		return false
	}
	return c.send(payload)
}

func (c *wsClient) writePump() {
	ticker := time.NewTicker(wsPingPeriod)
	defer func() {
		ticker.Stop()
		close(c.writeDone)
	}()

	for {
		select {
		case <-c.done:
			c.writeConfiguredClose()
			return
		default:
		}
		select {
		case request := <-c.outbox:
			_ = c.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if err := c.conn.WriteMessage(request.messageType, request.payload); err != nil {
				c.closeWithCode(ws.CloseAbnormalClosure, "Write failed")
				return
			}
			if request.final {
				c.closeWithCode(ws.CloseNormalClosure, "")
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if err := c.conn.WriteMessage(ws.PingMessage, nil); err != nil {
				c.closeWithCode(ws.CloseAbnormalClosure, "Ping failed")
				return
			}
		case <-c.done:
			c.writeConfiguredClose()
			return
		}
	}
}

func (c *wsClient) writeConfiguredClose() {
	c.closeMu.RLock()
	code := c.closeCode
	reason := c.closeReason
	c.closeMu.RUnlock()
	if code == ws.CloseAbnormalClosure {
		return
	}
	_ = c.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
	_ = c.conn.WriteMessage(ws.CloseMessage, ws.FormatCloseMessage(code, reason))
}

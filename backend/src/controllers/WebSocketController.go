package controllers

import (
	"encoding/json"
	"mc-manage-backend/src/utils"
	"strconv"
	"strings"

	ws "github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
)

var upgrader = ws.FastHTTPUpgrader{
	CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
		return true
	},
}

type terminalCommandMessage struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

// WsLogs streams server stdout/stderr to a WebSocket client
func WsLogs(c fiber.Ctx) error {
	id := c.Params("id")
	if err := ensureWebSocketUpgrade(c); err != nil {
		return err
	}
	logger.Info("[WsLogs] Client connected server="+id, nil)

	err := upgrader.Upgrade(c.RequestCtx(), func(conn *ws.Conn) {
		defer conn.Close()
		client := newWSClient(conn)

		broker := state.GetLogBroker(id)
		ch := broker.Subscribe()
		defer func() {
			client.close()
			broker.Unsubscribe(ch)
			logger.Info("[WsLogs] Client disconnected server="+id, nil)
		}()

		go func() {
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					client.close()
					return
				}
			}
		}()

		for {
			select {
			case line, ok := <-ch:
				if !ok {
					client.close()
					return
				}
				if !client.sendEnvelope("log", line) {
					return
				}
			case <-client.done:
				return
			}
		}
	})

	if err != nil {
		logger.Error("[WsLogs] Upgrade failed server="+id+": "+err.Error(), nil)
		return nil
	}
	// Return nil after successful upgrade — Fiber must not write to hijacked conn
	return nil
}

// WsTerminal provides bidirectional terminal: logs streamed out, commands sent in
func WsTerminal(c fiber.Ctx) error {
	id := c.Params("id")
	if err := ensureWebSocketUpgrade(c); err != nil {
		return err
	}
	logger.Info("[WsTerminal] Client connected server="+id, nil)

	err := upgrader.Upgrade(c.RequestCtx(), func(conn *ws.Conn) {
		defer conn.Close()
		client := newWSClient(conn)

		broker := state.GetLogBroker(id)
		ch := broker.Subscribe()
		defer func() {
			client.close()
			broker.Unsubscribe(ch)
			logger.Info("[WsTerminal] Client disconnected server="+id, nil)
		}()

		go func() {
			for {
				select {
				case line, ok := <-ch:
					if !ok {
						client.close()
						return
					}
					if !client.sendEnvelope("terminal_output", line) {
						return
					}
				case <-client.done:
					return
				}
			}
		}()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			cmd := parseTerminalCommand(msg)
			if cmd != "" {
				if err := state.SendCommand(id, cmd); err != nil {
					client.sendEnvelope("terminal_output", "[ERROR] "+err.Error())
				}
			}
		}

		client.close()
		client.wait()
	})

	if err != nil {
		logger.Error("[WsTerminal] Upgrade failed server="+id+": "+err.Error(), nil)
		return nil
	}
	return nil
}

func parseTerminalCommand(msg []byte) string {
	var payload terminalCommandMessage
	if err := json.Unmarshal(msg, &payload); err == nil && payload.Type == "command" {
		return strings.TrimSpace(payload.Data)
	}
	return strings.TrimSpace(string(msg))
}

// WsBackendLogs streams internal backend logs to a WebSocket client
func WsBackendLogs(c fiber.Ctx) error {
	if err := ensureWebSocketUpgrade(c); err != nil {
		return err
	}
	logger.Info("[WsBackendLogs] Client connected", nil)

	err := upgrader.Upgrade(c.RequestCtx(), func(conn *ws.Conn) {
		defer conn.Close()
		client := newWSClient(conn)

		ch := utils.BackendLogBroker.Subscribe()
		defer func() {
			client.close()
			utils.BackendLogBroker.Unsubscribe(ch)
			logger.Info("[WsBackendLogs] Client disconnected", nil)
		}()

		go func() {
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					client.close()
					return
				}
			}
		}()

		for {
			select {
			case line, ok := <-ch:
				if !ok {
					client.close()
					return
				}
				if !client.sendEnvelope("log", line) {
					return
				}
			case <-client.done:
				return
			}
		}
	})

	if err != nil {
		logger.Error("[WsBackendLogs] Upgrade failed: "+err.Error(), nil)
		return nil
	}
	return nil
}

func WsServerSetup(c fiber.Ctx) error {
	jobID := c.Params("job_id")
	job, ok := state.SetupJobs().Get(jobID)
	if !ok {
		return c.Status(fiber.StatusNotFound).SendString("Setup job not found")
	}
	if err := ensureWebSocketUpgrade(c); err != nil {
		return err
	}

	logger.Info("[WsServerSetup] Client connected job="+jobID, nil)
	lastEventID, _ := strconv.ParseInt(c.Query("last_event_id"), 10, 64)
	err := upgrader.Upgrade(c.RequestCtx(), func(conn *ws.Conn) {
		defer conn.Close()
		client := newWSClient(conn)

		ch, unsubscribe := job.SubscribeAfter(lastEventID)
		defer func() {
			client.close()
			unsubscribe()
			logger.Info("[WsServerSetup] Client disconnected job="+jobID, nil)
		}()

		go func() {
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					client.close()
					return
				}
			}
		}()

		for {
			select {
			case event, ok := <-ch:
				if !ok {
					client.close()
					return
				}
				if !client.send(event.JSON()) {
					return
				}
			case <-client.done:
				return
			}
		}
	})

	if err != nil {
		logger.Error("[WsServerSetup] Upgrade failed job="+jobID+": "+err.Error(), nil)
		return nil
	}
	return nil
}

package controllers

import (
	"mc-manage-backend/src/utils"
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

// WsLogs streams server stdout/stderr to a WebSocket client
func WsLogs(c fiber.Ctx) error {
	id := c.Params("id")
	logger.Info("[WsLogs] Client connected server="+id, nil)

	err := upgrader.Upgrade(c.RequestCtx(), func(conn *ws.Conn) {
		defer conn.Close()

		broker := state.GetLogBroker(id)
		ch := broker.Subscribe()
		defer func() {
			broker.Unsubscribe(ch)
			logger.Info("[WsLogs] Client disconnected server="+id, nil)
		}()

		done := make(chan struct{})
		go func() {
			defer close(done)
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()

		for {
			select {
			case line, ok := <-ch:
				if !ok {
					return
				}
				if err := conn.WriteMessage(ws.TextMessage, []byte(line)); err != nil {
					return
				}
			case <-done:
				return
			}
		}
	})

	if err != nil {
		logger.Error("[WsLogs] Upgrade failed server="+id+": "+err.Error(), nil)
		return err
	}
	// Return nil after successful upgrade — Fiber must not write to hijacked conn
	return nil
}

// WsTerminal provides bidirectional terminal: logs streamed out, commands sent in
func WsTerminal(c fiber.Ctx) error {
	id := c.Params("id")
	logger.Info("[WsTerminal] Client connected server="+id, nil)

	err := upgrader.Upgrade(c.RequestCtx(), func(conn *ws.Conn) {
		defer conn.Close()

		broker := state.GetLogBroker(id)
		ch := broker.Subscribe()
		defer func() {
			broker.Unsubscribe(ch)
			logger.Info("[WsTerminal] Client disconnected server="+id, nil)
		}()

		// Stream logs to client
		done := make(chan struct{})
		go func() {
			defer close(done)
			for {
				select {
				case line, ok := <-ch:
					if !ok {
						return
					}
					if err := conn.WriteMessage(ws.TextMessage, []byte(line)); err != nil {
						return
					}
				case <-done:
					return
				}
			}
		}()

		// Read commands from client
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			cmd := strings.TrimSpace(string(msg))
			if cmd != "" {
				if err := state.SendCommand(id, cmd); err != nil {
					conn.WriteMessage(ws.TextMessage, []byte("[ERROR] "+err.Error()))
				}
			}
		}

		// Signal done and wait for log streaming goroutine to finish
		close(done)
		<-done
	})

	if err != nil {
		logger.Error("[WsTerminal] Upgrade failed server="+id+": "+err.Error(), nil)
		return err
	}
	return nil
}

// WsBackendLogs streams internal backend logs to a WebSocket client
func WsBackendLogs(c fiber.Ctx) error {
	logger.Info("[WsBackendLogs] Client connected", nil)

	err := upgrader.Upgrade(c.RequestCtx(), func(conn *ws.Conn) {
		defer conn.Close()

		ch := utils.BackendLogBroker.Subscribe()
		defer func() {
			utils.BackendLogBroker.Unsubscribe(ch)
			logger.Info("[WsBackendLogs] Client disconnected", nil)
		}()

		done := make(chan struct{})
		go func() {
			defer close(done)
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()

		for {
			select {
			case line, ok := <-ch:
				if !ok {
					return
				}
				if err := conn.WriteMessage(ws.TextMessage, []byte(line)); err != nil {
					return
				}
			case <-done:
				return
			}
		}
	})

	if err != nil {
		logger.Error("[WsBackendLogs] Upgrade failed: "+err.Error(), nil)
		return err
	}
	return nil
}

func WsServerSetup(c fiber.Ctx) error {
	jobID := c.Params("job_id")
	job, ok := state.SetupJobs().Get(jobID)
	if !ok {
		return c.Status(fiber.StatusNotFound).SendString("Setup job not found")
	}

	logger.Info("[WsServerSetup] Client connected job="+jobID, nil)
	err := upgrader.Upgrade(c.RequestCtx(), func(conn *ws.Conn) {
		defer conn.Close()

		ch, unsubscribe := job.Subscribe()
		defer func() {
			unsubscribe()
			logger.Info("[WsServerSetup] Client disconnected job="+jobID, nil)
		}()

		done := make(chan struct{})
		go func() {
			defer close(done)
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()

		for {
			select {
			case event, ok := <-ch:
				if !ok {
					return
				}
				if err := conn.WriteMessage(ws.TextMessage, event.JSON()); err != nil {
					return
				}
			case <-done:
				return
			}
		}
	})

	if err != nil {
		logger.Error("[WsServerSetup] Upgrade failed job="+jobID+": "+err.Error(), nil)
		return err
	}
	return nil
}

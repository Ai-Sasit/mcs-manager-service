package controllers

import (
	"encoding/json"
	"mc-manage-backend/src/services"
	"mc-manage-backend/src/utils"
	"strconv"
	"strings"
	"time"

	ws "github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
)

type terminalCommandMessage struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type resourceSnapshotMessage struct {
	Type string                    `json:"type"`
	Ts   string                    `json:"ts"`
	Data services.ResourceSnapshot `json:"data"`
}

func parseLogCursor(c fiber.Ctx) (string, uint64) {
	streamID := strings.Clone(c.Query("stream_id"))
	lastEventID, _ := strconv.ParseUint(c.Query("last_event_id"), 10, 64)
	return streamID, lastEventID
}

func streamLogBroker(client *wsClient, broker *utils.LogBroker, messageType, streamID string, lastEventID uint64) {
	replay, subscription := broker.SubscribeAfter(streamID, lastEventID)
	defer subscription.Cancel()

	if replay.Reset && !client.sendStreamControl("stream_reset", replay, "The log stream restarted; replaying retained output.") {
		return
	}
	if replay.Gap && !client.sendStreamControl("stream_gap", replay, "Some log events are no longer retained; replaying the available range.") {
		return
	}
	for _, event := range replay.Events {
		select {
		case <-subscription.Done:
			if subscription.Reason() == utils.LogSubscriptionSlowConsumer {
				client.closeWithCode(ws.CloseTryAgainLater, "Log subscriber fell behind during replay")
			} else {
				client.close()
			}
			return
		default:
		}
		if !client.sendLogEvent(messageType, event) {
			return
		}
	}

	for {
		select {
		case event := <-subscription.Events:
			if !client.sendLogEvent(messageType, event) {
				return
			}
		case <-subscription.Done:
			if subscription.Reason() == utils.LogSubscriptionSlowConsumer {
				client.closeWithCode(ws.CloseTryAgainLater, "Log subscriber fell behind")
			} else {
				client.close()
			}
			return
		case <-client.done:
			return
		}
	}
}

// WsLogs streams server stdout/stderr to a WebSocket client
func WsLogs(c fiber.Ctx) error {
	id := strings.Clone(c.Params("id"))
	streamID, lastEventID := parseLogCursor(c)
	err := upgradeAuthorizedWebSocket(c, func(conn *ws.Conn) {
		defer conn.Close()
		client := newWSClient(conn)
		logger.Info("[WsLogs] Client connected server="+id, nil)

		broker := state.GetLogBroker(id)
		defer func() {
			client.close()
			client.wait()
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

		streamLogBroker(client, broker, "log", streamID, lastEventID)
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
	id := strings.Clone(c.Params("id"))
	streamID, lastEventID := parseLogCursor(c)
	err := upgradeAuthorizedWebSocket(c, func(conn *ws.Conn) {
		defer conn.Close()
		client := newWSClient(conn)
		logger.Info("[WsTerminal] Client connected server="+id, nil)

		broker := state.GetLogBroker(id)
		defer func() {
			client.close()
			client.wait()
			logger.Info("[WsTerminal] Client disconnected server="+id, nil)
		}()

		go streamLogBroker(client, broker, "terminal_output", streamID, lastEventID)

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
	streamID, lastEventID := parseLogCursor(c)
	err := upgradeAuthorizedWebSocket(c, func(conn *ws.Conn) {
		defer conn.Close()
		client := newWSClient(conn)
		logger.Info("[WsBackendLogs] Client connected", nil)

		defer func() {
			client.close()
			client.wait()
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

		streamLogBroker(client, utils.BackendLogBroker, "log", streamID, lastEventID)
	})

	if err != nil {
		logger.Error("[WsBackendLogs] Upgrade failed: "+err.Error(), nil)
		return nil
	}
	return nil
}

func WsServerSetup(c fiber.Ctx) error {
	jobID := strings.Clone(c.Params("job_id"))
	job, ok := state.SetupJobs().Get(jobID)
	if !ok {
		return c.Status(fiber.StatusNotFound).SendString("Setup job not found")
	}

	lastEventID, _ := strconv.ParseInt(c.Query("last_event_id"), 10, 64)
	err := upgradeAuthorizedWebSocket(c, func(conn *ws.Conn) {
		defer conn.Close()
		client := newWSClient(conn)
		logger.Info("[WsServerSetup] Client connected job="+jobID, nil)

		ch, unsubscribe := job.SubscribeAfter(lastEventID)
		defer func() {
			client.close()
			client.wait()
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
					if client.sendClose(wsCloseJobComplete, "Setup complete") {
						client.wait()
					}
					return
				}
				if !client.send(event.JSON()) {
					return
				}
				// If this is the terminal event, close gracefully after sending it.
				if event.Status == "success" || event.Status == "failed" {
					code := wsCloseJobComplete
					reason := "Setup complete"
					if event.Status == "failed" {
						code = wsCloseJobFailed
						reason = "Setup failed"
					}
					if client.sendClose(code, reason) {
						client.wait()
					}
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

func WsSystemResources(c fiber.Ctx) error {
	err := upgradeAuthorizedWebSocket(c, func(conn *ws.Conn) {
		defer conn.Close()
		client := newWSClient(conn)
		ticker := time.NewTicker(2 * time.Second)
		logger.Info("[WsSystemResources] Client connected", nil)
		defer func() {
			ticker.Stop()
			client.close()
			client.wait()
			logger.Info("[WsSystemResources] Client disconnected", nil)
		}()

		go func() {
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					client.close()
					return
				}
			}
		}()

		sendSnapshot := func() bool {
			snapshot, err := services.CollectResourceSnapshot(state)
			if err != nil {
				return client.sendEnvelope("error", "Failed to collect resource snapshot: "+err.Error())
			}
			payload, err := json.Marshal(resourceSnapshotMessage{
				Type: "resource_snapshot",
				Ts:   time.Now().Format(time.RFC3339),
				Data: snapshot,
			})
			if err != nil {
				return false
			}
			return client.send(payload)
		}

		if !sendSnapshot() {
			return
		}

		for {
			select {
			case <-ticker.C:
				if !sendSnapshot() {
					return
				}
			case <-client.done:
				return
			}
		}
	})

	if err != nil {
		logger.Error("[WsSystemResources] Upgrade failed: "+err.Error(), nil)
		return nil
	}
	return nil
}

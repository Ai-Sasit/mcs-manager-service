package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mc-manage-backend/src/models"
	"mc-manage-backend/src/utils"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var logger = utils.NewLogger("mc-manage")

// LogBroker fans out lines to multiple WebSocket subscribers
type LogBroker struct {
	mu          sync.Mutex
	subscribers map[chan string]struct{}
	history     []string
}

func newLogBroker() *LogBroker {
	return &LogBroker{
		subscribers: make(map[chan string]struct{}),
	}
}

func (b *LogBroker) Subscribe() chan string {
	ch := make(chan string, 256)
	b.mu.Lock()
	for _, line := range b.history {
		select {
		case ch <- line:
		default:
		}
	}
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *LogBroker) Unsubscribe(ch chan string) {
	b.mu.Lock()
	if _, ok := b.subscribers[ch]; ok {
		delete(b.subscribers, ch)
		close(ch)
	}
	b.mu.Unlock()
}

func (b *LogBroker) Publish(line string) {
	b.mu.Lock()
	b.history = append(b.history, line)
	if len(b.history) > 500 {
		b.history = b.history[1:]
	}
	for ch := range b.subscribers {
		select {
		case ch <- line:
		default:
		}
	}
	b.mu.Unlock()
}

func (b *LogBroker) Close() {
	b.mu.Lock()
	for ch := range b.subscribers {
		close(ch)
		delete(b.subscribers, ch)
	}
	b.mu.Unlock()
}

// logWriter implements io.Writer and publishes lines to a LogBroker
type logWriter struct {
	broker *LogBroker
	buf    []byte
	mu     sync.Mutex
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf = append(w.buf, p...)
	for {
		idx := bytes.IndexByte(w.buf, '\n')
		if idx < 0 {
			break
		}
		line := strings.TrimRight(string(w.buf[:idx]), "\r")
		w.buf = w.buf[idx+1:]
		if line != "" {
			w.broker.Publish(line)
		}
	}
	return len(p), nil
}

type AppState struct {
	mu          sync.RWMutex
	Servers     map[string]*models.ServerConfig
	processes   map[string]*exec.Cmd
	stdinPipes  map[string]io.WriteCloser
	logBrokers  map[string]*LogBroker
	Scheduler   *SchedulerService
	UserService *UserService
	DataDir     string
}

func NewAppState() *AppState {
	cwd, _ := os.Getwd()
	dataDir := filepath.Join(cwd, "data")
	os.MkdirAll(dataDir, os.ModePerm)
	os.MkdirAll(filepath.Join(dataDir, "servers"), os.ModePerm)

	state := &AppState{
		Servers:    make(map[string]*models.ServerConfig),
		processes:  make(map[string]*exec.Cmd),
		stdinPipes: make(map[string]io.WriteCloser),
		logBrokers: make(map[string]*LogBroker),
		DataDir:    dataDir,
	}

	state.loadServers()
	state.Scheduler = NewSchedulerService(state)
	state.UserService = NewUserService(dataDir)
	return state
}

func (s *AppState) loadServers() {
	path := filepath.Join(s.DataDir, "servers.json")
	data, err := os.ReadFile(path)
	if err != nil {
		logger.Info("[AppState] No existing servers.json, starting fresh", nil)
		return
	}

	var servers map[string]*models.ServerConfig
	if err := json.Unmarshal(data, &servers); err != nil {
		logger.Error("[AppState] Failed to parse servers.json: "+err.Error(), nil)
		return
	}

	for _, srv := range servers {
		srv.Status = models.StatusStopped
	}
	s.Servers = servers
	logger.Info(fmt.Sprintf("[AppState] Loaded %d server(s)", len(servers)), nil)
}

func (s *AppState) Save() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.DataDir, "servers.json")
	data, err := json.MarshalIndent(s.Servers, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(path, data, 0644)
}

func (s *AppState) GetServer(id string) (*models.ServerConfig, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	srv, ok := s.Servers[id]
	return srv, ok
}

func (s *AppState) ListServers() []*models.ServerConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*models.ServerConfig, 0, len(s.Servers))
	for _, srv := range s.Servers {
		list = append(list, srv)
	}
	return list
}

func (s *AppState) AddServer(config *models.ServerConfig) {
	s.mu.Lock()
	s.Servers[config.ID] = config
	s.mu.Unlock()
	s.Save()
}

func (s *AppState) RemoveServer(id string) (*models.ServerConfig, bool) {
	s.mu.Lock()
	srv, ok := s.Servers[id]
	if ok {
		delete(s.Servers, id)
	}
	s.mu.Unlock()
	if ok {
		s.Save()
	}
	return srv, ok
}

func (s *AppState) StartServer(id string) error {
	s.mu.Lock()
	srv, ok := s.Servers[id]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("server not found")
	}
	if srv.Status == models.StatusRunning {
		s.mu.Unlock()
		return fmt.Errorf("server is already running")
	}

	dir := srv.ServerDir
	logger.Info(fmt.Sprintf("[StartServer] Starting id=%s edition=%s", id, srv.Edition), nil)
	s.getLogBrokerLocked(id).Publish(fmt.Sprintf("[%s] [SYSTEM/INFO] Starting %s server (edition=%s, version=%s)...",
		time.Now().Format("15:04:05"), srv.Name, srv.Edition, srv.Version))

	var cmd *exec.Cmd
	switch srv.Edition {
	case models.EditionJava:
		cmd = exec.Command("java",
			fmt.Sprintf("-Xmx%dM", srv.MemoryMB),
			fmt.Sprintf("-Xms%dM", srv.MemoryMB),
			"-jar", "server.jar", "nogui",
		)
	case models.EditionBedrock:
		var exe string
		if runtime.GOOS == "windows" {
			exe = filepath.Join(dir, "bedrock_server.exe")
		} else {
			exe = filepath.Join(dir, "bedrock_server")
			os.Chmod(exe, 0755)
		}
		cmd = exec.Command(exe)
	default:
		s.mu.Unlock()
		return fmt.Errorf("unknown edition: %s", srv.Edition)
	}

	cmd.Dir = dir

	broker := s.getLogBrokerLocked(id)
	lw := &logWriter{broker: broker}
	cmd.Stdout = lw
	cmd.Stderr = lw

	stdin, stdinErr := cmd.StdinPipe()

	if err := cmd.Start(); err != nil {
		s.mu.Unlock()
		logger.Error("[StartServer] Process failed to start id="+id+": "+err.Error(), nil)
		s.publishLog(id, "ERROR", "Failed to start process: "+err.Error())
		return fmt.Errorf("failed to start: %w", err)
	}

	srv.Status = models.StatusRunning
	s.processes[id] = cmd
	if stdinErr == nil {
		s.stdinPipes[id] = stdin
	}
	s.mu.Unlock()
	logger.Info("[StartServer] Process started id="+id, nil)
	s.publishLog(id, "INFO", "Process started. Waiting for server to be ready...")

	// Watch for process exit in a goroutine
	go func() {
		err := cmd.Wait()
		if err != nil {
			logger.Warn(fmt.Sprintf("[StartServer] Process exited id=%s: %s", id, err.Error()), nil)
			s.publishLog(id, "WARN", "Server process exited unexpectedly: "+err.Error())
		} else {
			logger.Info("[StartServer] Process exited cleanly id="+id, nil)
			s.publishLog(id, "INFO", "Server process stopped.")
		}

		s.mu.Lock()
		// Only clean up if this is still the active process for this server
		if s.processes[id] == cmd {
			delete(s.processes, id)
			if pipe, ok := s.stdinPipes[id]; ok {
				pipe.Close()
				delete(s.stdinPipes, id)
			}
			if srv, ok := s.Servers[id]; ok {
				srv.Status = models.StatusStopped
			}
		}
		s.mu.Unlock()
		s.Save()
	}()

	s.Save()
	return nil
}

func (s *AppState) StopServer(id string) error {
	s.mu.Lock()
	cmd, hasProcess := s.processes[id]
	stdin, hasStdin := s.stdinPipes[id]
	srv := s.Servers[id]
	s.mu.Unlock()

	if hasProcess && cmd.Process != nil {
		// Try graceful shutdown first: send "stop" command
		if hasStdin {
			logger.Info("[StopServer] Sending graceful stop command id="+id, nil)
			s.publishLog(id, "INFO", "Sending stop command to server...")
			fmt.Fprintln(stdin, "stop")
		}

		// Wait up to 5 seconds for graceful exit
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		select {
		case <-done:
			logger.Info("[StopServer] Process exited gracefully id="+id, nil)
			s.publishLog(id, "INFO", "Server stopped gracefully.")
		case <-time.After(5 * time.Second):
			logger.Warn("[StopServer] Graceful timeout, force killing id="+id, nil)
			s.publishLog(id, "WARN", "Graceful stop timed out — force killing process.")
			cmd.Process.Kill()
			<-done
			logger.Info("[StopServer] Process force killed id="+id, nil)
			s.publishLog(id, "WARN", "Process force killed.")
		}
	} else {
		logger.Info("[StopServer] No running process id="+id, nil)
	}

	// Clean up resources
	s.mu.Lock()
	if s.processes[id] == cmd {
		delete(s.processes, id)
	}
	if pipe, ok := s.stdinPipes[id]; ok {
		pipe.Close()
		delete(s.stdinPipes, id)
	}
	if srv != nil {
		srv.Status = models.StatusStopped
	}
	s.mu.Unlock()

	s.Save()
	return nil
}

func (s *AppState) RestartServer(id string) error {
	logger.Info("[RestartServer] Stopping id="+id, nil)
	s.publishLog(id, "INFO", "Restart requested — stopping server...")
	s.StopServer(id)
	time.Sleep(2 * time.Second)
	logger.Info("[RestartServer] Restarting id="+id, nil)
	s.publishLog(id, "INFO", "Restarting server now...")
	return s.StartServer(id)
}

func (s *AppState) StopAllServers() {
	logger.Info("[StopAllServers] Stopping all running servers", nil)
	s.mu.Lock()
	for id, cmd := range s.processes {
		if cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait()
			logger.Info("[StopAllServers] Killed id="+id, nil)
		}
		if srv, ok := s.Servers[id]; ok {
			srv.Status = models.StatusStopped
		}
	}
	s.processes = make(map[string]*exec.Cmd)
	s.stdinPipes = make(map[string]io.WriteCloser)
	s.mu.Unlock()
	s.Save()
	logger.Info("[StopAllServers] Done", nil)
}

func (s *AppState) getLogBrokerLocked(id string) *LogBroker {
	if _, ok := s.logBrokers[id]; !ok {
		s.logBrokers[id] = newLogBroker()
	}
	return s.logBrokers[id]
}

func (s *AppState) GetLogBroker(id string) *LogBroker {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.getLogBrokerLocked(id)
}

// publishLog writes a timestamped [SYSTEM] line into a server's log broker
// so it appears in the frontend live log alongside the Minecraft output.
func (s *AppState) publishLog(id, level, msg string) {
	ts := time.Now().Format("15:04:05")
	line := fmt.Sprintf("[%s] [SYSTEM/%s] %s", ts, level, msg)
	s.mu.Lock()
	broker := s.getLogBrokerLocked(id)
	s.mu.Unlock()
	broker.Publish(line)
}

func (s *AppState) SendCommand(id string, cmd string) error {
	s.mu.RLock()
	stdin, ok := s.stdinPipes[id]
	_, running := s.processes[id]
	s.mu.RUnlock()
	if !ok || !running {
		return fmt.Errorf("server is not running")
	}
	_, err := fmt.Fprintln(stdin, cmd)
	if err != nil {
		return fmt.Errorf("failed to send command: server may be shutting down")
	}
	return err
}

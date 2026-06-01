package services

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
)

type SetupEvent struct {
	Type     string `json:"type"`
	Step     string `json:"step"`
	Status   string `json:"status"`
	Message  string `json:"message"`
	Percent  int    `json:"percent"`
	ServerID string `json:"server_id,omitempty"`
}

type SetupJob struct {
	ID      string
	events  []SetupEvent
	clients map[chan SetupEvent]struct{}
	done    bool
	mu      sync.RWMutex
}

type SetupJobManager struct {
	jobs map[string]*SetupJob
	mu   sync.RWMutex
}

func NewSetupJobManager() *SetupJobManager {
	return &SetupJobManager{jobs: make(map[string]*SetupJob)}
}

func (m *SetupJobManager) Create() *SetupJob {
	job := &SetupJob{
		ID:      uuid.New().String(),
		events:  make([]SetupEvent, 0, 12),
		clients: make(map[chan SetupEvent]struct{}),
	}
	m.mu.Lock()
	m.jobs[job.ID] = job
	m.mu.Unlock()
	return job
}

func (m *SetupJobManager) Get(id string) (*SetupJob, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	job, ok := m.jobs[id]
	return job, ok
}

func (m *SetupJobManager) DeleteAfter(id string, delay time.Duration) {
	go func() {
		time.Sleep(delay)
		m.mu.Lock()
		delete(m.jobs, id)
		m.mu.Unlock()
	}()
}

func (j *SetupJob) Publish(event SetupEvent) {
	j.mu.Lock()
	j.events = append(j.events, event)
	if event.Status == "success" || event.Status == "failed" {
		j.done = true
	}
	for ch := range j.clients {
		select {
		case ch <- event:
		default:
		}
	}
	if j.done {
		for ch := range j.clients {
			close(ch)
			delete(j.clients, ch)
		}
	}
	j.mu.Unlock()
}

func (j *SetupJob) Subscribe() (<-chan SetupEvent, func()) {
	ch := make(chan SetupEvent, 32)

	j.mu.Lock()
	for _, event := range j.events {
		ch <- event
	}
	if j.done {
		close(ch)
		j.mu.Unlock()
		return ch, func() {}
	}
	j.clients[ch] = struct{}{}
	j.mu.Unlock()

	unsubscribe := func() {
		j.mu.Lock()
		if _, ok := j.clients[ch]; ok {
			delete(j.clients, ch)
			close(ch)
		}
		j.mu.Unlock()
	}

	return ch, unsubscribe
}

func (e SetupEvent) JSON() []byte {
	data, err := json.Marshal(e)
	if err != nil {
		return []byte(`{"type":"setup","step":"unknown","status":"failed","message":"Failed to encode setup event","percent":100}`)
	}
	return data
}

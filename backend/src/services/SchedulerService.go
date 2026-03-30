package services

import (
	"encoding/json"
	"fmt"
	"mc-manage-backend/src/utils"
	"os"
	"path/filepath"
	"sync"

	"github.com/robfig/cron/v3"
)

type ScheduleEntry struct {
	ID       string       `json:"id"`
	ServerID string       `json:"server_id"`
	Task     string       `json:"task"`    // "backup", "restart", "command"
	Cron     string       `json:"cron"`    // e.g. "0 0 * * *"
	Command  string       `json:"command"` // Only for task="command"
	Enabled  bool         `json:"enabled"`
	EntryID  cron.EntryID `json:"-"`
}

type SchedulerService struct {
	mu      sync.RWMutex
	cron    *cron.Cron
	entries map[string]*ScheduleEntry
	state   *AppState
	dataDir string
}

func NewSchedulerService(state *AppState) *SchedulerService {
	s := &SchedulerService{
		cron:    cron.New(),
		entries: make(map[string]*ScheduleEntry),
		state:   state,
		dataDir: state.DataDir,
	}
	s.loadEntries()
	s.cron.Start()
	return s
}

func (s *SchedulerService) loadEntries() {
	path := filepath.Join(s.dataDir, "schedules.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var saved []ScheduleEntry
	if err := json.Unmarshal(data, &saved); err != nil {
		logger.Error("[SchedulerService] Failed to load schedules: "+err.Error(), nil)
		return
	}

	for _, entry := range saved {
		e := entry // copy
		s.AddSchedule(&e)
	}
}

func (s *SchedulerService) save() {
	s.mu.RLock()
	var list []ScheduleEntry
	for _, e := range s.entries {
		list = append(list, *e)
	}
	s.mu.RUnlock()

	path := filepath.Join(s.dataDir, "schedules.json")
	data, _ := json.MarshalIndent(list, "", "  ")
	os.WriteFile(path, data, 0644)
}

func (s *SchedulerService) AddSchedule(e *ScheduleEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if e.Enabled {
		id, err := s.cron.AddFunc(e.Cron, func() {
			s.runTask(e)
		})
		if err != nil {
			return err
		}
		e.EntryID = id
	}

	s.entries[e.ID] = e
	s.save()
	return nil
}

func (s *SchedulerService) DeleteSchedule(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if e, ok := s.entries[id]; ok {
		if e.Enabled {
			s.cron.Remove(e.EntryID)
		}
		delete(s.entries, id)
		s.save()
	}
}

func (s *SchedulerService) ToggleSchedule(id string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.entries[id]
	if !ok {
		return fmt.Errorf("schedule not found")
	}

	if e.Enabled == enabled {
		return nil
	}

	if enabled {
		id, err := s.cron.AddFunc(e.Cron, func() {
			s.runTask(e)
		})
		if err != nil {
			return err
		}
		e.EntryID = id
	} else {
		s.cron.Remove(e.EntryID)
	}

	e.Enabled = enabled
	s.save()
	return nil
}

func (s *SchedulerService) runTask(e *ScheduleEntry) {
	utils.LogAudit("system", "SCHEDULE_TASK_START", e.ServerID, fmt.Sprintf("Running task: %s", e.Task))

	switch e.Task {
	case "backup":
		// Backup logic
		srv, ok := s.state.GetServer(e.ServerID)
		if !ok {
			return
		}
		backupDir := filepath.Join("data", "backups", e.ServerID)
		os.MkdirAll(backupDir, 0755)
		target := filepath.Join(backupDir, fmt.Sprintf("%s_autobackup_%s.zip", srv.Name, utils.GenerateID()))
		utils.CreateZip(srv.ServerDir, target)

	case "restart":
		s.state.RestartServer(e.ServerID)

	case "command":
		if e.Command != "" {
			s.state.SendCommand(e.ServerID, e.Command)
		}
	}
}

func (s *SchedulerService) ListSchedules() []ScheduleEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var list []ScheduleEntry
	for _, e := range s.entries {
		list = append(list, *e)
	}
	return list
}

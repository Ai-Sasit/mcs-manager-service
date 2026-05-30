package services

import (
	"fmt"
	"mc-manage-backend/src/utils"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ScheduleEntry struct {
	ID       string       `bson:"id" json:"id"`
	ServerID string       `bson:"server_id" json:"server_id"`
	Task     string       `bson:"task" json:"task"`       // "backup", "restart", "command"
	Cron     string       `bson:"cron" json:"cron"`       // e.g. "0 0 * * *"
	Command  string       `bson:"command" json:"command"` // Only for task="command"
	Enabled  bool         `bson:"enabled" json:"enabled"`
	EntryID  cron.EntryID `bson:"-" json:"-"`
}

type SchedulerService struct {
	mu      sync.RWMutex
	cron    *cron.Cron
	entries map[string]*ScheduleEntry
	state   *AppState
	col     *mongo.Collection
}

func NewSchedulerService(state *AppState) *SchedulerService {
	s := &SchedulerService{
		cron:    cron.New(),
		entries: make(map[string]*ScheduleEntry),
		state:   state,
		col:     utils.GetCollection(utils.DB, "schedules"),
	}
	s.loadEntries()
	s.cron.Start()
	return s
}

func (s *SchedulerService) loadEntries() {
	ctx, cancel := utils.MongoContext(10 * time.Second)
	defer cancel()

	cursor, err := s.col.Find(ctx, bson.M{})
	if err != nil {
		logger.Error("[SchedulerService] Failed to load schedules from MongoDB: "+err.Error(), nil)
		return
	}
	defer cursor.Close(ctx)

	var saved []ScheduleEntry
	if err := cursor.All(ctx, &saved); err != nil {
		logger.Error("[SchedulerService] Failed to decode schedules from MongoDB: "+err.Error(), nil)
		return
	}

	for _, entry := range saved {
		e := entry
		if err := s.registerSchedule(&e); err != nil {
			logger.Error("[SchedulerService] Failed to register schedule "+e.ID+": "+err.Error(), nil)
			continue
		}
		s.entries[e.ID] = &e
	}
}

func (s *SchedulerService) saveEntry(e *ScheduleEntry) {
	ctx, cancel := utils.MongoContext(10 * time.Second)
	defer cancel()
	if _, err := s.col.ReplaceOne(ctx, bson.M{"id": e.ID}, e, options.Replace().SetUpsert(true)); err != nil {
		logger.Error("[SchedulerService] Failed to save schedule "+e.ID+": "+err.Error(), nil)
	}
}

func (s *SchedulerService) deleteEntry(id string) {
	ctx, cancel := utils.MongoContext(10 * time.Second)
	defer cancel()
	if _, err := s.col.DeleteOne(ctx, bson.M{"id": id}); err != nil {
		logger.Error("[SchedulerService] Failed to delete schedule "+id+": "+err.Error(), nil)
	}
}

func (s *SchedulerService) registerSchedule(e *ScheduleEntry) error {
	if !e.Enabled {
		return nil
	}
	id, err := s.cron.AddFunc(e.Cron, func() {
		s.runTask(e)
	})
	if err != nil {
		return err
	}
	e.EntryID = id
	return nil
}

func (s *SchedulerService) AddSchedule(e *ScheduleEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.registerSchedule(e); err != nil {
		return err
	}

	s.entries[e.ID] = e
	s.saveEntry(e)
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
		s.deleteEntry(id)
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
		e.Enabled = true
		if err := s.registerSchedule(e); err != nil {
			e.Enabled = false
			return err
		}
	} else {
		s.cron.Remove(e.EntryID)
		e.EntryID = 0
		e.Enabled = false
	}

	s.saveEntry(e)
	return nil
}

func (s *SchedulerService) runTask(e *ScheduleEntry) {
	utils.LogAudit("system", "SCHEDULE_TASK_START", e.ServerID, fmt.Sprintf("Running task: %s", e.Task))

	switch e.Task {
	case "backup":
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

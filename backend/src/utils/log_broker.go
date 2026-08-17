package utils

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultLogHistorySize  = 10000
	logSubscriberQueueSize = 512

	LogSubscriptionCancelled    = "cancelled"
	LogSubscriptionSlowConsumer = "slow_consumer"
	LogSubscriptionBrokerClosed = "broker_closed"
)

type LogEvent struct {
	EventID  uint64 `json:"event_id"`
	StreamID string `json:"stream_id"`
	Data     string `json:"data"`
	Ts       string `json:"ts"`
}

type LogReplay struct {
	Events        []LogEvent
	StreamID      string
	Gap           bool
	Reset         bool
	OldestEventID uint64
	LatestEventID uint64
}

type logSubscriber struct {
	events chan LogEvent
	done   chan struct{}

	mu     sync.RWMutex
	reason string
	once   sync.Once
}

func newLogSubscriber() *logSubscriber {
	return &logSubscriber{
		events: make(chan LogEvent, logSubscriberQueueSize),
		done:   make(chan struct{}),
	}
}

func (s *logSubscriber) close(reason string) {
	s.once.Do(func() {
		s.mu.Lock()
		s.reason = reason
		s.mu.Unlock()
		close(s.done)
	})
}

func (s *logSubscriber) closeReason() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.reason
}

type LogSubscription struct {
	Events <-chan LogEvent
	Done   <-chan struct{}

	broker *LogBroker
	sub    *logSubscriber
	once   sync.Once
}

func (s *LogSubscription) Cancel() {
	if s == nil || s.broker == nil || s.sub == nil {
		return
	}
	s.once.Do(func() {
		s.broker.removeSubscriber(s.sub, LogSubscriptionCancelled)
	})
}

func (s *LogSubscription) Reason() string {
	if s == nil || s.sub == nil {
		return ""
	}
	return s.sub.closeReason()
}

// LogBroker retains a bounded, ordered event history and atomically attaches
// live subscribers after producing their replay window.
type LogBroker struct {
	mu          sync.Mutex
	subscribers map[*logSubscriber]struct{}
	history     []LogEvent
	capacity    int
	start       int
	nextID      uint64
	streamID    string
	closed      bool
}

func NewLogBroker() *LogBroker {
	return &LogBroker{
		subscribers: make(map[*logSubscriber]struct{}),
		history:     make([]LogEvent, 0, DefaultLogHistorySize),
		capacity:    DefaultLogHistorySize,
		streamID:    uuid.NewString(),
	}
}

func (b *LogBroker) Publish(line string) uint64 {
	return b.PublishAt(line, time.Now())
}

func (b *LogBroker) PublishAt(line string, eventTime time.Time) uint64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return 0
	}

	b.nextID++
	event := LogEvent{
		EventID:  b.nextID,
		StreamID: b.streamID,
		Data:     line,
		Ts:       eventTime.UTC().Format(time.RFC3339Nano),
	}
	b.appendHistoryLocked(event)

	for sub := range b.subscribers {
		select {
		case sub.events <- event:
		default:
			delete(b.subscribers, sub)
			sub.close(LogSubscriptionSlowConsumer)
		}
	}
	return event.EventID
}

func (b *LogBroker) appendHistoryLocked(event LogEvent) {
	if len(b.history) < b.capacity {
		b.history = append(b.history, event)
		return
	}
	b.history[b.start] = event
	b.start = (b.start + 1) % b.capacity
}

func (b *LogBroker) historyLocked() []LogEvent {
	events := make([]LogEvent, len(b.history))
	for i := 0; i < len(b.history); i++ {
		events[i] = b.history[(b.start+i)%len(b.history)]
	}
	return events
}

func (b *LogBroker) SubscribeAfter(requestedStreamID string, lastEventID uint64) (LogReplay, *LogSubscription) {
	b.mu.Lock()
	defer b.mu.Unlock()

	sub := newLogSubscriber()
	subscription := &LogSubscription{
		Events: sub.events,
		Done:   sub.done,
		broker: b,
		sub:    sub,
	}
	replay := LogReplay{StreamID: b.streamID}
	if b.closed {
		sub.close(LogSubscriptionBrokerClosed)
		return replay, subscription
	}

	history := b.historyLocked()
	if len(history) > 0 {
		replay.OldestEventID = history[0].EventID
		replay.LatestEventID = history[len(history)-1].EventID
	}

	replay.Reset = requestedStreamID != "" && requestedStreamID != b.streamID
	if !replay.Reset && requestedStreamID == b.streamID && lastEventID > replay.LatestEventID {
		replay.Reset = true
	}

	switch {
	case requestedStreamID == "" || replay.Reset:
		replay.Events = history
	default:
		if len(history) > 0 && lastEventID > 0 && lastEventID+1 < replay.OldestEventID {
			replay.Gap = true
		}
		for _, event := range history {
			if event.EventID > lastEventID {
				replay.Events = append(replay.Events, event)
			}
		}
	}

	b.subscribers[sub] = struct{}{}
	return replay, subscription
}

func (b *LogBroker) removeSubscriber(sub *logSubscriber, reason string) {
	b.mu.Lock()
	if _, ok := b.subscribers[sub]; ok {
		delete(b.subscribers, sub)
	}
	sub.close(reason)
	b.mu.Unlock()
}

func (b *LogBroker) Close() {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return
	}
	b.closed = true
	for sub := range b.subscribers {
		delete(b.subscribers, sub)
		sub.close(LogSubscriptionBrokerClosed)
	}
	b.mu.Unlock()
}

type persistedLogLine struct {
	time     time.Time
	filename string
	line     int
	data     string
}

// HydrateBackendLogBroker loads timestamped entries from today's backend log
// files so a fresh process can still provide a useful initial replay window.
func HydrateBackendLogBroker(logDir string, day time.Time) error {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	prefix := day.Format("2006-01-02")
	lines := make([]persistedLogLine, 0, DefaultLogHistorySize)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), prefix) || !strings.HasSuffix(entry.Name(), ".log") {
			continue
		}
		path := filepath.Join(logDir, entry.Name())
		file, openErr := os.Open(path)
		if openErr != nil {
			continue
		}
		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		lineNumber := 0
		for scanner.Scan() {
			lineNumber++
			line := scanner.Text()
			if len(line) < 19 {
				continue
			}
			parsed, parseErr := time.ParseInLocation("2006-01-02 15:04:05", line[:19], day.Location())
			if parseErr != nil {
				continue
			}
			lines = append(lines, persistedLogLine{
				time:     parsed,
				filename: entry.Name(),
				line:     lineNumber,
				data:     "[" + line[:19] + "]" + line[19:],
			})
		}
		scanErr := scanner.Err()
		_ = file.Close()
		if scanErr != nil {
			return scanErr
		}
	}

	sort.SliceStable(lines, func(i, j int) bool {
		if lines[i].time.Equal(lines[j].time) {
			if lines[i].filename == lines[j].filename {
				return lines[i].line < lines[j].line
			}
			return lines[i].filename < lines[j].filename
		}
		return lines[i].time.Before(lines[j].time)
	})
	if len(lines) > DefaultLogHistorySize {
		lines = lines[len(lines)-DefaultLogHistorySize:]
	}
	for _, line := range lines {
		BackendLogBroker.PublishAt(line.data, line.time)
	}
	return nil
}

var BackendLogBroker = NewLogBroker()

package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

type LogMeta map[string]interface{}

type LogContext struct {
	Method      string
	OriginalURL string
	StatusCode  string
	PerformTime string
}

type Logger struct {
	filename string
	logDir   string
}

// LogBroker fans out lines to multiple WebSocket subscribers
type LogBroker struct {
	mu          sync.Mutex
	subscribers map[chan string]struct{}
	history     []string
}

func NewLogBroker() *LogBroker {
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

var BackendLogBroker = NewLogBroker()

func NewLogger(filename string) *Logger {
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		os.Mkdir("logs", os.ModePerm)
	}
	return &Logger{
		filename: filename,
		logDir:   "logs",
	}
}

func (l *Logger) Init(filename string) {
	if filename != "" {
		l.filename = filename
	}
	log.SetFlags(0)
	l.Info("[Logger] Initialized and pointing to file: \""+l.getFilename()+"\"", nil)
}

func (l *Logger) getFilename() string {
	name := l.logDir + "/" + time.Now().Format("2006-01-02")
	if l.filename != "" {
		name += "-" + l.filename
	}
	return name + ".log"
}

func (l *Logger) write(msg string, meta LogMeta) {
	entry := l.StripANSI(msg)
	if meta != nil {
		entry += "\n    " + l.stringify(meta)
	}
	f, err := os.OpenFile(l.getFilename(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file: ", err)
		return
	}
	defer f.Close()
	if _, err := f.WriteString(entry + "\n"); err != nil {
		fmt.Println("Error writing to file: ", err)
	}
}

func (l *Logger) getHeader(ctx LogContext, level string) string {
	return level + " \"" + ctx.Method + " " + ctx.OriginalURL + "\""
}

func (l *Logger) Log(level string, ctx interface{}, meta LogMeta) {
	var logMsg string
	stripLen := len(l.StripANSI(level))
	switch v := ctx.(type) {
	case LogContext:
		logMsg = l.getHeader(v, level) + " " + v.StatusCode + " " + v.PerformTime
	default:
		logMsg = level + strings.Repeat(" ", 5-stripLen) + ": " + v.(string)
	}
	log.Println("\033[37m[" + l.now() + "]\033[0m " + logMsg)
	l.write(l.now()+" "+logMsg, meta)
	BackendLogBroker.Publish("[" + l.now() + "] " + logMsg)
}

func (l *Logger) Info(message string, meta LogMeta) {
	l.Log("\033[32mINFO\033[0m", message, meta)
}

func (l *Logger) Warn(message string, meta LogMeta) {
	l.Log("\033[33mWARN\033[0m", message, meta)
}

func (l *Logger) Debug(message string, meta LogMeta) {
	if os.Getenv("DEBUG_MODE") == "true" {
		l.Log("\033[36mDEBUG\033[0m", message, meta)
	}
}

func (l *Logger) Error(message string, meta LogMeta) {
	l.Log("\033[31mERROR\033[0m", message, meta)
}

func (l *Logger) stringify(obj LogMeta) string {
	b, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		return "Error converting to JSON: " + err.Error()
	}
	return string(b)
}

func (l *Logger) StripANSI(str string) string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return ansiRegex.ReplaceAllString(str, "")
}

func (l *Logger) now() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

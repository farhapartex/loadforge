package web

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	logBufferSize    = 200
	logChannelBuffer = 64
	tailPollInterval = 300 * time.Millisecond
)

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

type LogBroadcaster struct {
	mu      sync.Mutex
	clients map[chan LogEntry]struct{}
	buffer  []LogEntry
}

func newLogBroadcaster() *LogBroadcaster {
	return &LogBroadcaster{
		clients: make(map[chan LogEntry]struct{}),
		buffer:  make([]LogEntry, 0, logBufferSize),
	}
}

func (lb *LogBroadcaster) subscribe() chan LogEntry {
	ch := make(chan LogEntry, logChannelBuffer)
	lb.mu.Lock()
	for _, entry := range lb.buffer {
		select {
		case ch <- entry:
		default:
		}
	}
	lb.clients[ch] = struct{}{}
	lb.mu.Unlock()
	return ch
}

func (lb *LogBroadcaster) unsubscribe(ch chan LogEntry) {
	lb.mu.Lock()
	delete(lb.clients, ch)
	close(ch)
	lb.mu.Unlock()
}

func (lb *LogBroadcaster) publish(entry LogEntry) {
	lb.mu.Lock()
	if len(lb.buffer) >= logBufferSize {
		lb.buffer = lb.buffer[1:]
	}
	lb.buffer = append(lb.buffer, entry)
	for ch := range lb.clients {
		select {
		case ch <- entry:
		default:
		}
	}
	lb.mu.Unlock()
}

// tailFile opens path (creating it if needed), seeks to the current end so
// existing content is skipped, then polls every tailPollInterval for new lines
// and broadcasts them. Stops when ctx is cancelled.
func (lb *LogBroadcaster) tailFile(ctx context.Context, path string) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		log.Printf("WARN tailFile: cannot open %q: %v", path, err)
		return
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	ticker := time.NewTicker(tailPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			lb.drainLines(reader)
		case <-ctx.Done():
			return
		}
	}
}

func (lb *LogBroadcaster) drainLines(r *bufio.Reader) {
	for {
		line, err := r.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")
		if line != "" {
			lb.publish(parseLogLine(line))
		}
		if err != nil {
			return
		}
	}
}

// parseLogLine handles the format written by log.Ltime: "15:04:05 message"
func parseLogLine(line string) LogEntry {
	if len(line) > 9 && line[8] == ' ' {
		msg := line[9:]
		return LogEntry{Timestamp: line[:8], Level: detectLevel(msg), Message: msg}
	}
	return LogEntry{
		Timestamp: time.Now().Format("15:04:05"),
		Level:     detectLevel(line),
		Message:   line,
	}
}

func (lb *LogBroadcaster) handleStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ch := lb.subscribe()
	defer lb.unsubscribe(ch)

	for {
		select {
		case entry, ok := <-ch:
			if !ok {
				return
			}
			data, err := json.Marshal(entry)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func detectLevel(msg string) string {
	up := strings.ToUpper(msg)
	switch {
	case strings.Contains(up, "ERROR") || strings.Contains(up, "FATAL"):
		return "ERROR"
	case strings.Contains(up, "WARN"):
		return "WARN"
	default:
		return "INFO"
	}
}

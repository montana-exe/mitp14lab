package social

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type JSONLWriter[T any] struct {
	path      string
	batchSize int
	flushEach time.Duration
	buffer    []T
	mu        sync.Mutex
	lastFlush time.Time
}

func NewJSONLWriter[T any](path string, batchSize int, flushEach time.Duration) *JSONLWriter[T] {
	if batchSize <= 0 {
		batchSize = 100
	}
	if flushEach <= 0 {
		flushEach = time.Second
	}
	return &JSONLWriter[T]{path: path, batchSize: batchSize, flushEach: flushEach, lastFlush: time.Now()}
}

func (w *JSONLWriter[T]) Add(item T) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buffer = append(w.buffer, item)
	if len(w.buffer) >= w.batchSize || time.Since(w.lastFlush) >= w.flushEach {
		return w.flushLocked()
	}
	return nil
}

func (w *JSONLWriter[T]) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.flushLocked()
}

func (w *JSONLWriter[T]) flushLocked() error {
	if len(w.buffer) == 0 {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(w.path), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(w.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	for _, item := range w.buffer {
		if err := encoder.Encode(item); err != nil {
			return err
		}
	}
	w.buffer = w.buffer[:0]
	w.lastFlush = time.Now()
	return nil
}

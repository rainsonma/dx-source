package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// SSEWriter wraps an http.ResponseWriter for Server-Sent Events
type SSEWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// NewSSEWriter creates a new SSE writer and sets appropriate headers
func NewSSEWriter(w http.ResponseWriter) *SSEWriter {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, _ := w.(http.Flusher)

	return &SSEWriter{w: w, flusher: flusher}
}

// Write sends a data event as JSON
func (s *SSEWriter) Write(data any) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal SSE data: %w", err)
	}

	_, err = fmt.Fprintf(s.w, "data: %s\n\n", jsonBytes)
	if err != nil {
		return fmt.Errorf("failed to write SSE data: %w", err)
	}

	s.flush()
	return nil
}

// WriteError sends an error event
func (s *SSEWriter) WriteError(message string) error {
	_, err := fmt.Fprintf(s.w, "event: error\ndata: %s\n\n", message)
	if err != nil {
		return err
	}
	s.flush()
	return nil
}

// Close sends a done event
func (s *SSEWriter) Close() {
	fmt.Fprintf(s.w, "event: done\ndata: \n\n")
	s.flush()
}

// flush safely flushes the response writer, recovering from panics
// caused by writing to a closed connection.
func (s *SSEWriter) flush() {
	if s.flusher == nil {
		return
	}
	defer func() { recover() }()
	s.flusher.Flush()
}

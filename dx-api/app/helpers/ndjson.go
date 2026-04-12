package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// NDJSONWriter wraps an http.ResponseWriter for Newline-Delimited JSON streaming.
type NDJSONWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// NewNDJSONWriter creates a new NDJSON writer and sets appropriate headers.
func NewNDJSONWriter(w http.ResponseWriter) *NDJSONWriter {
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, _ := w.(http.Flusher)

	return &NDJSONWriter{w: w, flusher: flusher}
}

// Write sends data as a JSON line.
func (n *NDJSONWriter) Write(data any) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal NDJSON data: %w", err)
	}

	_, err = fmt.Fprintf(n.w, "%s\n", jsonBytes)
	if err != nil {
		return fmt.Errorf("failed to write NDJSON data: %w", err)
	}

	n.flush()
	return nil
}

// WriteError sends an error line.
func (n *NDJSONWriter) WriteError(message string) error {
	payload := map[string]string{"error": message}
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(n.w, "%s\n", jsonBytes)
	if err != nil {
		return err
	}
	n.flush()
	return nil
}

// Close sends a done sentinel line.
func (n *NDJSONWriter) Close() {
	payload := map[string]string{"done": "true"}
	jsonBytes, _ := json.Marshal(payload)
	fmt.Fprintf(n.w, "%s\n", jsonBytes)
	n.flush()
}

// flush safely flushes the response writer, recovering from panics
// caused by writing to a closed connection.
func (n *NDJSONWriter) flush() {
	if n.flusher == nil {
		return
	}
	defer func() { recover() }()
	n.flusher.Flush()
}

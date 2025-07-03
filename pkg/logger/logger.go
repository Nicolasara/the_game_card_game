package logger

import (
	"encoding/json"
	"os"
)

// Logger provides structured JSON logging to a file.
type Logger struct {
	encoder *json.Encoder
	file    *os.File
}

// New creates a new Logger that writes to the specified file.
func New(filePath string) (*Logger, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{
		encoder: json.NewEncoder(file),
		file:    file,
	}, nil
}

// Log writes a structured log entry.
func (l *Logger) Log(v interface{}) error {
	return l.encoder.Encode(v)
}

// Close closes the underlying log file.
func (l *Logger) Close() error {
	return l.file.Close()
}

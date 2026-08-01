package logger

import (
    "encoding/json"
    "io"
    "time"
)

// JSONLogger writes structured JSON logs.
type JSONLogger struct {
    writer io.Writer
}

func NewJSONLogger(writer io.Writer) *JSONLogger {
    return &JSONLogger{writer: writer}
}

func (l *JSONLogger) log(level string, message string, fields map[string]any) {
    record := map[string]any{
        "timestamp":   time.Now().UTC().Format(time.RFC3339),
        "level":       level,
        "message":     message,
        "moduleFields": fields,
    }
    data, _ := json.Marshal(record)
    data = append(data, byte('\n'))
    _, _ = l.writer.Write(data)
}

func (l *JSONLogger) Info(message string, fields map[string]any) {
    l.log("info", message, fields)
}

func (l *JSONLogger) Error(message string, fields map[string]any) {
    l.log("error", message, fields)
}

func (l *JSONLogger) Debug(message string, fields map[string]any) {
    l.log("debug", message, fields)
}

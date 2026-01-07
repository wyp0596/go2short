package logger

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

var (
	encoder = json.NewEncoder(os.Stdout)
	mu      sync.Mutex
)

type Entry struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Msg     string `json:"msg"`
	ReqID   string `json:"req_id,omitempty"`
	Code    string `json:"code,omitempty"`
	Latency float64 `json:"latency_ms,omitempty"`
	Status  int    `json:"status,omitempty"`
	Error   string `json:"error,omitempty"`
	Extra   map[string]any `json:"extra,omitempty"`
}

func Info(msg string, fields ...func(*Entry)) {
	log("INFO", msg, fields...)
}

func Error(msg string, fields ...func(*Entry)) {
	log("ERROR", msg, fields...)
}

func log(level, msg string, fields ...func(*Entry)) {
	e := &Entry{
		Time:  time.Now().UTC().Format(time.RFC3339),
		Level: level,
		Msg:   msg,
	}
	for _, f := range fields {
		f(e)
	}

	mu.Lock()
	encoder.Encode(e)
	mu.Unlock()
}

// Field helpers
func ReqID(id string) func(*Entry) {
	return func(e *Entry) { e.ReqID = id }
}

func Code(code string) func(*Entry) {
	return func(e *Entry) { e.Code = code }
}

func Latency(ms float64) func(*Entry) {
	return func(e *Entry) { e.Latency = ms }
}

func Status(status int) func(*Entry) {
	return func(e *Entry) { e.Status = status }
}

func Err(err error) func(*Entry) {
	return func(e *Entry) {
		if err != nil {
			e.Error = err.Error()
		}
	}
}

func Extra(key string, val any) func(*Entry) {
	return func(e *Entry) {
		if e.Extra == nil {
			e.Extra = make(map[string]any)
		}
		e.Extra[key] = val
	}
}

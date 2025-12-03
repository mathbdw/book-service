package graylog

import (
	"encoding/json"
	"net"
	"time"
)

type GELFWriter struct {
	conn    net.Conn
	host    string
	version string
}

// Constructor - build gelfWriter
func NewGELFWriter(opts ...Option) (*GELFWriter, error) {
	g := &GELFWriter{}

	for _, opt := range opts {
		opt(g)
	}

	conn, err := net.Dial("udp", g.host)
	if err != nil {
		return nil, err
	}

	g.conn = conn
	return g, nil
}

// Write - Implements the io.Writer interface
func (g *GELFWriter) Write(p []byte) (n int, err error) {
	var logEntry map[string]any
	if err := json.Unmarshal(p, &logEntry); err != nil {
		return 0, err
	}

	gelfMessage := map[string]any{
		"version":       g.version,
		"host":          g.host,
		"short_message": logEntry["message"],
		"timestamp":     time.Now().Unix(),
		"level":         g.convertLevel(logEntry["level"]),
	}

	// Add field
	for key, value := range logEntry {
		if key != "message" && key != "level" && key != "time" {
			gelfMessage["_"+key] = value
		}
	}

	data, err := json.Marshal(gelfMessage)
	if err != nil {
		return 0, err
	}

	return g.conn.Write(data)
}

// convertLevel - convert level
func (g *GELFWriter) convertLevel(level any) int {
	switch level {
	case "debug":
		return 7
	case "info":
		return 6
	case "warn":
		return 4
	case "error":
		return 3
	case "fatal":
		return 2
	default:
		return 6
	}
}

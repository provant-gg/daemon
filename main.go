package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS events (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	received_at TEXT NOT NULL,
	event_name  TEXT,
	raw_json    TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_events_event_name  ON events(event_name);
CREATE INDEX IF NOT EXISTS idx_events_received_at ON events(received_at);
`

func main() {
	addr := flag.String("addr", "127.0.0.1:49123", "Rocket League stats socket")
	dbPath := flag.String("db", "events.db", "SQLite database path")
	reconnect := flag.Duration("reconnect", 2*time.Second, "delay between reconnect attempts")
	flag.Parse()

	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)", *dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec(schema); err != nil {
		log.Fatalf("init schema: %v", err)
	}
	insert, err := db.Prepare(`INSERT INTO events(received_at, event_name, raw_json) VALUES (?, ?, ?)`)
	if err != nil {
		log.Fatalf("prepare: %v", err)
	}
	defer insert.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		log.Println("shutdown signal received")
		cancel()
	}()

	log.Printf("logging events from %s -> %s", *addr, *dbPath)

	for ctx.Err() == nil {
		if err := stream(ctx, *addr, insert); err != nil && ctx.Err() == nil {
			log.Printf("stream ended: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(*reconnect):
		}
	}
}

func stream(ctx context.Context, addr string, insert *sql.Stmt) error {
	d := net.Dialer{Timeout: 5 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	log.Printf("connected to %s", addr)
	defer conn.Close()

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	r := bufio.NewReaderSize(conn, 1<<16)
	count := 0
	defer func() { log.Printf("disconnected after %d events", count) }()

	for {
		line, err := r.ReadString('\n')
		if line = strings.TrimRight(line, "\r\n"); line != "" {
			name := peekEventName(line)
			if _, dbErr := insert.Exec(time.Now().UTC().Format(time.RFC3339Nano), name, line); dbErr != nil {
				log.Printf("insert: %v", dbErr)
			}
			count++
			if count%1000 == 0 {
				log.Printf("captured %d events", count)
			}
		}
		if err != nil {
			return err
		}
	}
}

// peekEventName extracts the event name from the JSON envelope without
// fully decoding the payload. Stats API frames look like
// {"event":"Initialized","data":{...}} but we tolerate "name" as a fallback.
func peekEventName(s string) string {
	var probe struct {
		Event string `json:"event"`
		Name  string `json:"name"`
	}
	if err := json.Unmarshal([]byte(s), &probe); err != nil {
		return ""
	}
	if probe.Event != "" {
		return probe.Event
	}
	return probe.Name
}

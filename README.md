# provantgg-daemon

Reads Rocket League's local Stats API stream and writes every event to a SQLite database for analysis.

## Install

### macOS, Linux, FreeBSD

```sh
curl -fsSL https://raw.githubusercontent.com/provant-gg/daemon/master/install.sh | bash
```

The installer auto-detects your OS and architecture, downloads the matching release, verifies the checksum, and drops the binary into `/usr/local/bin` (or `~/.local/bin` if `/usr/local/bin` isn't writable).

Override defaults via env vars:

```sh
# Pin to a specific version
curl -fsSL https://raw.githubusercontent.com/provant-gg/daemon/master/install.sh | VERSION=v0.1.0 bash

# Custom install location
curl -fsSL https://raw.githubusercontent.com/provant-gg/daemon/master/install.sh | BIN_DIR=$HOME/bin bash
```

### Windows (x64)

Download `provantgg-daemon_Windows_x86_64_setup.exe` from the [latest release](https://github.com/provant-gg/daemon/releases/latest) and double-click it. The installer copies `provantgg-daemon.exe` to `C:\Program Files\provantgg-daemon\`, adds it to the system `PATH`, and registers an uninstaller in **Add or Remove Programs**. Open a new terminal afterwards so the updated `PATH` takes effect.

### Windows (ARM64 or x86)

Grab the matching `provantgg-daemon_Windows_<arch>.zip` from the [releases page](https://github.com/provant-gg/daemon/releases/latest), extract `provantgg-daemon.exe`, and put it somewhere on your `PATH`. (No installer is built for these architectures.)

## Usage

Make sure Rocket League is running with the Stats API enabled (default in recent builds — port `49123`), then:

```sh
provantgg-daemon -db events.db
```

Flags:

| Flag | Default | Description |
|------|---------|-------------|
| `-addr` | `127.0.0.1:49123` | Rocket League stats socket |
| `-db` | `events.db` | SQLite database path |
| `-reconnect` | `2s` | Delay between reconnect attempts when the game socket drops |

The daemon reconnects automatically as you start/stop matches, so you can leave it running.

## Schema

Single table, one row per event:

```sql
CREATE TABLE events (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  received_at TEXT NOT NULL,    -- RFC3339Nano UTC
  event_name  TEXT,             -- peeked from JSON envelope ("event" or "name")
  raw_json    TEXT NOT NULL     -- verbatim line from the socket
);
```

Quick analysis:

```sh
sqlite3 events.db 'SELECT event_name, count(*) FROM events GROUP BY 1 ORDER BY 2 DESC;'
```

## Build from source

Requires Go 1.22+.

```sh
git clone https://github.com/provant-gg/daemon.git
cd daemon
go build -o provantgg-daemon .
```

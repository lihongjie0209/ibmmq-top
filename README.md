<div align="center">

# mq-top

**`htop` for IBM MQ** — a real-time, interactive terminal UI for monitoring queues, channels, topics, and subscriptions.

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![IBM MQ](https://img.shields.io/badge/IBM%20MQ-9.x-054ADA?style=flat)](https://www.ibm.com/products/mq)
[![License](https://img.shields.io/badge/license-MIT-green?style=flat)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-linux%2Famd64-lightgrey?style=flat)]()

</div>

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ mq-top  QMgr: QM1   Status: RUNNING   Uptime: 2d3h5m   Conns: 12           │
├──────────────────────────────────────────────────────────────────────────────┤
│ 1:Queues  2:Channels  3:Topics  4:Subscriptions                              │
├──────────────────────────────────────────────────────────────────────────────┤
│ Name               Depth  MaxDepth  %Full              InHdl OutHdl MsgAge  │
│ APP.INPUT.QUEUE    1523   5000      [████████░░░░] 30%  2     1      0       │
│ APP.OUTPUT.QUEUE   4899   5000      [████████████] 97%  0     3      42      │
│ SYSTEM.DEAD.LETTER 0      5000      [░░░░░░░░░░░░]  0%  0     0      0       │
├──────────────────────────────────────────────────────────────────────────────┤
│ q:Quit  Tab:Next  1-4:Jump  ↑↓:Scroll                                        │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Features

- **4 live panels** — Queues, Channels, Topics, Subscriptions, all auto-refreshing
- **Visual queue depth bar** — progress bar + percentage for instant capacity awareness
- **Color-coded health** — green = healthy · yellow = warning · red = critical
- **Channel status at a glance** — RUNNING / RETRYING / STOPPED mapped to colors
- **Sortable columns** — click any column header to sort or reverse
- **Demo mode** — try the UI without any MQ connection (`-demo`)
- **Client-mode support** — monitor remote queue managers over TCP
- **Configurable refresh** — default 5 s, tunable to any Go duration

---

## Quick Start

### 1. Try it instantly (demo mode)

No IBM MQ installation needed:

```bash
./out/mq_top -demo
```

### 2. Monitor a local queue manager

```bash
# Copy binary into your running MQ container
docker cp ./out/mq_top <mq-container>:/tmp/

# Launch
docker exec -it <mq-container> /tmp/mq_top -ibmmq.queueManager QM1
```

### 3. Connect remotely (client mode)

```bash
./out/mq_top \
  -ibmmq.queueManager QM1 \
  -ibmmq.connName     "mqhost(1414)" \
  -ibmmq.channel      SYSTEM.DEF.SVRCONN \
  -ibmmq.userId       mquser \
  -ibmmq.password     mqpassword
```

Or via the `MQSERVER` environment variable:

```bash
export MQSERVER="SYSTEM.DEF.SVRCONN/TCP/mqhost(1414)"
./out/mq_top -ibmmq.queueManager QM1
```

---

## Building

> **Prerequisites:** Docker — the binary is cross-compiled inside a Linux container so you don't need a local MQ SDK.

**Windows**
```bat
build.bat
```

**Linux / macOS**
```bash
chmod +x build.sh && ./build.sh
```

The output binary lands at `./out/mq_top` (Linux amd64, CGO + MQ Redist Client).

**Air-gapped / custom Go proxy**
```bash
docker build \
  --build-arg GOPROXY=https://your-proxy.example.com \
  -t mq-top-builder -f Dockerfile.build .
```

---

## Testing with a Local IBM MQ Container

```bash
# 1. Start IBM MQ
docker run -d \
  --name mq1 \
  -e LICENSE=accept \
  -e MQ_QMGR_NAME=QM1 \
  -p 1414:1414 \
  icr.io/ibm-messaging/mq:latest

# 2. Build
./build.sh          # Linux/macOS
# build.bat         # Windows

# 3. Run inside the container (local mode)
docker cp ./out/mq_top mq1:/tmp/
docker exec -it mq1 /tmp/mq_top -ibmmq.queueManager QM1

# 4. Run from outside (client mode)
docker exec -it mq1 \
  runmqsc QM1 <<< "DEFINE CHANNEL(SYSTEM.DEF.SVRCONN) CHLTYPE(SVRCONN) MCAUSER('') REPLACE"

./out/mq_top \
  -ibmmq.queueManager QM1 \
  -ibmmq.connName     "localhost(1414)" \
  -ibmmq.channel      SYSTEM.DEF.SVRCONN
```

---

## Keyboard Shortcuts

| Key | Action |
|---|---|
| `q` / `Q` | Quit |
| `Tab` | Next tab |
| `Shift+Tab` | Previous tab |
| `1` | Jump to Queues |
| `2` | Jump to Channels |
| `3` | Jump to Topics |
| `4` | Jump to Subscriptions |
| `↑` / `↓` | Scroll rows |

---

## Configuration Reference

| Flag | Default | Description |
|---|---|---|
| `-ibmmq.queueManager` | *(default QMgr)* | Queue manager name |
| `-ibmmq.connName` | — | Remote connection string `host(port)` |
| `-ibmmq.channel` | — | Client channel name |
| `-ibmmq.userId` | — | MQ user ID |
| `-ibmmq.password` | — | MQ password |
| `-ibmmq.replyQueue` | `SYSTEM.DEFAULT.MODEL.QUEUE` | Model queue for PCF replies |
| `-ibmmq.monitoredQueues` | `*` | Comma-separated queue name patterns |
| `-ibmmq.monitoredChannels` | `*` | Comma-separated channel name patterns |
| `-ibmmq.monitoredTopics` | `*` | Comma-separated topic string patterns |
| `-ibmmq.monitoredSubscriptions` | `*` | Comma-separated subscription name patterns |
| `-refresh` | `5s` | Refresh interval (any Go duration, e.g. `10s`, `1m`) |
| `-demo` | `false` | Run with synthetic data — no MQ connection required |

---

## Project Structure

```
mq-top/
├── cmd/mq_top/
│   ├── main.go         Entry point, MQ connection, goroutine wiring
│   ├── config.go       CLI flags and ConnectionConfig
│   ├── collector.go    MQ PCF polling and data snapshot model
│   └── demo.go         Synthetic data generator for demo mode
├── ui/
│   ├── types.go        Shared data types (Snapshot, QueueInfo, …)
│   ├── app.go          Bubble Tea application, layout, key bindings
│   ├── model.go        Top-level Bubble Tea model
│   ├── header.go       Status bar: QMgr name, status, uptime
│   ├── footer.go       Help bar: key bindings
│   ├── queues.go       Queues panel
│   ├── channels.go     Channels panel
│   ├── topics.go       Topics panel
│   ├── subs.go         Subscriptions panel
│   └── styles.go       Lipgloss color/style definitions
├── Dockerfile.build    Reproducible Linux build image
├── build.bat           Windows build script
├── build.sh            Linux/macOS build script
└── go.mod
```

---

## Requirements

| Dependency | Version |
|---|---|
| Docker | any recent version (build only) |
| IBM MQ queue manager | 9.x |
| Target platform | Linux amd64 |

---

## Acknowledgements

- [mq-golang](https://github.com/ibm-messaging/mq-golang) — Go bindings for IBM MQ
- [mq-metric-samples](https://github.com/ibm-messaging/mq-metric-samples) — IBM MQ metrics exporter framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — Elm-inspired TUI framework for Go
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Style definitions for terminal layouts

---

## License

Released under the [MIT License](LICENSE).

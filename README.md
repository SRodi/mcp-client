# Network Telemetry MCP Client

A unified CLI tool that provides MCP (Model Context Protocol) server capabilities for real-time network connectivity analytics and AI-powered insights.

## Features

- **Process Identification**: Query by PID or process name
- **Network Connection Tracking**: Monitor connection attempts and patterns
- **AI-Powered Insights**: OpenAI GPT-3.5-turbo integration for network behavior analysis
- **Multiple Output Modes**: Summary, detailed listings, and pattern analysis
- **HTTP REST API**: Simple and reliable communication with the eBPF server
- **MCP Integration**: Self-contained MCP server with internal client communication
- **Interactive Mode**: Command-line interface for real-time network analysis

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   netspy CLI    │───▶│   MCP Server    │───▶│  eBPF Server    │
│  (MCP Client)   │    │  (Internal)     │    │ (HTTP API)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │  OpenAI API     │
                       │ (GPT-3.5-turbo) │
                       └─────────────────┘
```

## Prerequisites

1. **eBPF Network Monitor Server**: Build from [ebpf-server repository](https://github.com/SRodi/ebpf-server)
   ```bash
   git clone git@github.com:SRodi/ebpf-server.git
   cd ebpf-server
   make build  # Compiles both Go code AND eBPF programs
   ```

2. **Root Privileges**: Required for eBPF operations on the server

3. **Optional**: OpenAI API key for insights (set `OPENAI_API_KEY` environment variable)

## Installation

```bash
go build -o netspy ./cmd/netspy
```

## Usage

### Interactive Mode (Default)

Start an interactive session with the integrated MCP server:

```bash
# Start interactive mode
./netspy
```

Interactive commands:
```
netspy-mcp> summary --pid 1234 --duration 120
netspy-mcp> list --process curl --max-events 20
netspy-mcp> analyze --process nginx
netspy-mcp> insights "curl made 5 connection attempts in 60 seconds"
netspy-mcp> tools
netspy-mcp> help
netspy-mcp> quit
```

### Single Command Mode

Execute specific MCP tools directly:

```bash
# Get network summary
./netspy --tool get_network_summary --pid 1234 --duration 120

# List connections
./netspy --tool list_connections --process curl --max-events 15

# Analyze patterns
./netspy --tool analyze_patterns --process ssh

# Get AI insights (requires summary text)
./netspy --tool ai_insights --summary-text "nginx made 25 connections in 300 seconds"

# Get packet drop summary
./netspy --tool get_packet_drop_summary --process nginx --duration 300

# List packet drops
./netspy --tool list_packet_drops --pid 1234 --max-events 10
```

### Quick Start

The application works with a persistent eBPF HTTP API server:

```bash
# 1. Start the eBPF API server (run once and keep running)
cd /path/to/ebpf-server
sudo ./bin/ebpf-server --http --port 8080

# 2. Generate some network traffic
curl -s http://google.com
curl -s http://github.com

# 3. Use netspy to analyze connections
./netspy
# Then in interactive mode:
# netspy-mcp> summary --process curl
# netspy-mcp> list
# netspy-mcp> analyze --process curl
```

## MCP Tools

The integrated MCP server provides these tools:

- **get_network_summary**: Get connection statistics for processes/PIDs
- **list_connections**: List recent network connection events with filtering
- **analyze_patterns**: Analyze connection patterns and provide insights
- **ai_insights**: Generate AI-powered insights using OpenAI GPT-3.5-turbo
- **get_packet_drop_summary**: Get summary of packet drop events for processes
- **list_packet_drops**: List recent packet drop events with filtering

## Command Line Options

### General Options
- `--server URL`: eBPF server URL (default: http://localhost:8080)
- `--verbose`: Enable verbose logging
- `--help`: Show help information

### Tool Execution
- `--tool TOOL`: Run specific MCP tool and exit

### Tool Parameters
- `--pid PID`: Process ID to monitor
- `--process NAME`: Process name to monitor
- `--duration SECONDS`: Duration in seconds (default: 60)
- `--max-events COUNT`: Maximum events to retrieve (default: 100)
- `--summary-text TEXT`: Summary text for AI insights

## API Endpoints

The eBPF server provides the following HTTP REST API endpoints that the MCP server communicates with:

- `GET /health` - Health check
- `POST /api/connection-summary` - Get connection statistics
- `GET /api/list-connections` - List connections (simple queries)
- `POST /api/list-connections` - List connections (complex queries)

## Sample Output

### Interactive Mode
```bash
$ ./netspy
🔗 Network Telemetry MCP Server
Starting interactive mode...

Available commands:
  summary      - Get network connection summary
  list         - List recent connections
  analyze      - Analyze connection patterns
  insights     - Get AI insights about network behavior
  tools        - Show available MCP tools
  help         - Show this help message
  quit/exit    - Exit interactive mode

netspy-mcp> summary --process curl
Process 'curl' made 5 outbound connection attempts over the last 60 seconds

netspy-mcp> list --max-events 5
Recent connection events (15 total):
  21:05:53 | 127.0.0.1:8080 | TCP | netspy
  21:03:38 | (local socket) | UNIX | snapd
  21:01:08 | (local socket) | UNIX | snapd
  20:58:42 | 192.168.120.2:53 | UDP | systemd-resolve
  20:56:15 | 172.217.164.78:443 | TCP | curl
  ... and 10 more events

netspy-mcp> analyze --process curl
Connection Analysis:
  Top destinations:
    172.217.164.78:443 (3 connections)
    8.8.8.8:53 (2 connections)
  Protocols: TCP (3), UDP (2)

netspy-mcp> insights "curl made 5 connections in 60 seconds"
🤖 AI Network Insights:
This connection pattern shows typical web client behavior...
```

### Single Command Mode
```bash
$ ./netspy --tool get_network_summary --process curl
Process 'curl' made 5 outbound connection attempts over the last 60 seconds
```

## Troubleshooting

**"Connection refused" or failed connection**
- Ensure the eBPF API server is running with `--http --port 8080`
- Check that nothing else is using port 8080
- Verify the server started successfully (check for error messages)
- Test server health: `curl http://localhost:8080/health`

**"No connections found" when server has data**
- **Solution**: Ensure the eBPF API server is running:
  ```bash
  # Start server
  cd /path/to/ebpf-server
  sudo ./bin/ebpf-server --http --port 8080
  
  # Generate some traffic
  curl -s http://google.com
  
  # Query netspy
  ./netspy
  # netspy-mcp> list
  ```
- **Check server status**: Use `netstat -tlnp | grep 8080` to verify server is listening

**"open bpf/connection.o: no such file or directory"**
- Build server with `make build` in ebpf-server repository
- Server needs compiled eBPF programs, not just Go binary

**"permission denied"**  
- Use `sudo` for the eBPF server - eBPF operations require root privileges
- netspy client can run without sudo when using HTTP API mode

**Tool failures**
- Ensure eBPF server is running and accessible
- Check `--server` URL parameter points to correct eBPF server
- Use `--verbose` flag for detailed error information

## Developer Usage

### MCP Integration
```go
import "github.com/srodi/netspy/internal/mcp"

// Create MCP client with embedded server
mcpClient := mcp.NewMCPClient("http://localhost:8080", true)

// Start interactive mode
ctx := context.Background()
err := mcpClient.StartInteractiveMode(ctx)

// Or execute single command
arguments := map[string]any{
    "pid": 1234,
    "duration": 120,
}
result, err := mcpClient.RunSingleCommand(ctx, "get_network_summary", arguments)
```

### HTTP Client
```go
import "github.com/srodi/netspy/internal/netclient"

// Create HTTP client
client := netclient.NewClient("http://localhost:8080")

// Connect and use
ctx := context.Background()
if err := client.Connect(ctx); err != nil {
    return err
}
defer client.Close()

// Get connection summary
summary, err := client.GetConnectionSummary(ctx, pid, process, duration)

// List connections
connections, err := client.ListConnections(ctx, nil, nil)
```

### Key Components:
- **MCP Server**: Self-contained Model Context Protocol implementation
- **MCP Client**: Interactive and programmatic interface to MCP tools
- **HTTP REST API**: Communication layer with eBPF server
- **AI Integration**: OpenAI GPT-3.5-turbo for network behavior insights
- **JSON Request/Response**: Standard HTTP content types
- **Health Monitoring**: Built-in health check endpoint
- **Error Handling**: Proper HTTP status codes and error messages

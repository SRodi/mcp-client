# mcp-client

A flexible CLI tool that queries the MCP eBPF server to provide real-time network connectivity analytics and insights.

## Features

- **Flexible Process Identification**: Query by PID or process name
- **Network Connection Tracking**: Monitor connection attempts and patterns
- **AI-Powered Insights**: OpenAI integration for network behavior analysis and recommendations
- **Connection Listing**: View detailed connection events  
- **Configurable Duration**: Analyze connections over custom time periods
- **Multiple Output Modes**: Summary, detailed listings, and pattern analysis

## Prerequisites

- [mcp-ebpf server](https://github.com/SRodi/mcp-ebpf) running (default: localhost:8080)
- Optional: OpenAI API key for network insights (set `OPENAI_API_KEY` environment variable)

**Note**: The mcp-ebpf server captures `connect()` syscall attempts, not actual network latency. This tool provides insights about connection patterns and frequency, which is valuable for understanding application behavior and network usage patterns.

## Installation

```bash
go build -o mcp-client ./cmd/client
```

## Usage

### Basic Usage

```bash
# Analyze by process name (recommended for most use cases)
./mcp-client --process curl

# Analyze by PID
./mcp-client --pid 1234

# Analyze over custom time period
./mcp-client --process ssh --duration 300  # 5 minutes

# List all recent connections
./mcp-client --list

# List connections for specific process
./mcp-client --process nginx --list

# Analyze connection patterns
./mcp-client --process database --analyze
```

### Advanced Usage

```bash
# Verbose output with server details
./mcp-client --process curl --verbose

# Custom MCP server URL
./mcp-client --process ssh --server http://remote-server:8080/mcp

# Combine listing with pattern analysis
./mcp-client --process web-server --list --analyze --max-events 20

# Quick debugging - list recent connections
./mcp-client --list --max-events 5
```

### Command Line Options

- `--pid <number>`: Analyze specific process ID
- `--process <name>`: Analyze by process name (e.g., 'curl', 'ssh', 'nginx')
- `--duration <seconds>`: Time window to analyze (default: 60 seconds)
- `--server <url>`: MCP server URL (default: http://localhost:8080/mcp)
- `--list`: List recent connection events instead of summary
- `--analyze`: Show connection pattern analysis
- `--max-events <number>`: Maximum events to show in list mode (default: 10)
- `--verbose`: Enable verbose output

## Use Cases

### Network Analytics & Monitoring

```bash
# Get network activity insights for a service
./mcp-client --process nginx --analyze

# Monitor database connection attempts over time
./mcp-client --process postgres --duration 1800  # 30 minutes

# Analyze connection patterns for optimization
./mcp-client --process web-service --list --analyze

# Real-time network activity overview  
./mcp-client --list --max-events 20
```

### Application Behavior Analysis

```bash
# Analyze network connection patterns for insights
./mcp-client --process api-server --duration 3600 --verbose

# Compare connection patterns between services  
./mcp-client --process service-v1 --analyze
./mcp-client --process service-v2 --analyze

# Monitor connection frequency and patterns
./mcp-client --process cache-service --duration 600
```

### Development & Integration

```bash
# Understand application network behavior
./mcp-client --process myapp --list --analyze

# Validate expected connections during development
./mcp-client --process test-service --verbose

# Monitor integration points
./mcp-client --process integration-service --duration 900
```

## Sample Output

### Summary Mode
```
🔍 Network Telemetry Summary:
Process 'curl' made 5 outbound connection attempts over the last 60 seconds

🤖 AI Network Insights:
This connection pattern shows typical web client behavior. The process made 5 connection attempts over 60 seconds, suggesting multiple HTTP requests or redirects. This frequency is normal for a command-line HTTP client. Consider connection pooling if this becomes a high-frequency service...
```

### List Mode
```
Recent connection events (12 total):
  14:30:45 | 93.184.216.34:80 | TCP | curl
  14:30:40 | 8.8.8.8:53 | UDP | curl
  14:30:35 | 151.101.1.140:443 | TCP | curl
  ... and 9 more events
```

### Analysis Mode
```
Connection Analysis:
  Top destinations:
    93.184.216.34:80 (3 connections)
    8.8.8.8:53 (2 connections)
    151.101.1.140:443 (1 connections)
  Protocols: TCP (4), UDP (1)
```

## Integration with mcp-ebpf

This client is designed to work seamlessly with the [mcp-ebpf server](https://github.com/SRodi/mcp-ebpf). The eBPF server must be running with root privileges to monitor network connections:

```bash
# Start the eBPF server (in another terminal)
cd /path/to/mcp-ebpf
sudo make run

# Then use this client
./mcp-client --process your-app
```

## Error Handling

The client provides helpful error messages and suggestions:

- **No connections found**: Suggests checking process name or increasing duration
- **Server unavailable**: Provides server connection details
- **Invalid process**: Suggests using `--list` to see available processes
- **OpenAI unavailable**: Gracefully continues without AI insights

## Contributing

Feel free to submit issues and enhancement requests!

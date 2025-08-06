# Netspy - Contextual Network Telemetry Analyzer

A unified CLI tool that provides MCP (Model Context Protocol) server capabilities for real-time network connectivity analytics with AI-powered insights and OpenAI function calling integration.

## 🚀 Features

### Core Capabilities
- **Real-time Network Monitoring**: Track connection attempts, patterns, and packet drops
- **Process-Specific Analysis**: Query by PID or process name
- **AI-Powered Insights**: Advanced OpenAI function calling with contextual tool orchestration
- **Interactive & Programmatic**: Both CLI and API interfaces available
- **Multiple Output Modes**: Summary, detailed listings, pattern analysis, and AI insights

### AI Integration Highlights
- **OpenAI Function Calling**: LLM automatically selects and chains multiple analysis tools
- **Contextual Tool Orchestration**: Dynamic tool selection based on query context
- **Comprehensive Analysis**: Multi-tool data synthesis for actionable insights
- **Natural Language Queries**: Ask questions in plain English about network behavior

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   netspy CLI    │───▶│   MCP Server    │───▶│  eBPF Server    │
│  (MCP Client)   │    │  (Internal)     │    │ (HTTP API)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │  OpenAI API     │
                       │   (Function     │
                       │    Calling)     │
                       └─────────────────┘
```

### Function Calling Flow
```
User Query → Conversation Manager → OpenAI API → Function Calls → MCP Tools → Results → Final Response
```

1. **User Input**: Natural language query about network behavior
2. **OpenAI Processing**: LLM analyzes query and determines needed tools
3. **Function Calls**: System executes appropriate MCP tools with validated parameters
4. **Result Integration**: Tool outputs are fed back to the LLM
5. **Final Response**: Comprehensive analysis based on real data

## 📋 Prerequisites

1. **eBPF Network Monitor Server**: Build from [ebpf-server repository](https://github.com/SRodi/ebpf-server)
   ```bash
   git clone git@github.com:SRodi/ebpf-server.git
   cd ebpf-server
   make build  # Compiles both Go code AND eBPF programs
   ```

2. **Root Privileges**: Required for eBPF operations on the server

3. **OpenAI API Key**: Required for AI insights (set `OPENAI_API_KEY` environment variable)
   ```bash
   export OPENAI_API_KEY=your_openai_api_key_here
   ```

## 🛠️ Installation

```bash
go build -o netspy ./cmd/netspy
```

## 🎯 Quick Start

```bash
# 1. Start the eBPF API server (run once and keep running)
cd /path/to/ebpf-server
sudo ./bin/ebpf-server --http --port 8080

# 2. Set OpenAI API key
export OPENAI_API_KEY=your_key_here

# 3. Generate some network traffic
curl -s http://google.com

# 4. Use contextual analysis
./netspy
netspy-mcp> contextual "analyze my system"
```

## 💬 Usage Examples

### Contextual Analysis (Recommended)

The AI-powered analysis automatically selects and chains multiple tools:

```bash
# Interactive mode with contextual analysis
./netspy
netspy-mcp> contextual "What's happening with my network connections?"
netspy-mcp> contextual "Are there any packet drops or connection issues?"
netspy-mcp> contextual "Analyze the network behavior of process nginx"

# Command line mode
./netspy --tool contextual_analysis --query "Analyze my network activity"
./netspy --tool contextual_analysis --query "How is curl behaving?" --process curl
```

### Traditional Tool Commands

```bash
# Interactive mode
./netspy
netspy-mcp> summary --pid 1234 --duration 120
netspy-mcp> list --process curl --max-events 20
netspy-mcp> analyze --process nginx
netspy-mcp> insights "curl made 5 connections in 60 seconds"

# Single command mode
./netspy --tool get_network_summary --pid 1234 --duration 120
./netspy --tool list_connections --process curl --max-events 15
./netspy --tool analyze_patterns --process ssh
```

## 🔧 Available Tools

### Core Analysis Tools
- **get_network_summary**: Aggregated connection statistics for processes
- **list_connections**: Recent network connection events with filtering
- **get_packet_drop_summary**: Packet loss analysis for connectivity issues
- **list_packet_drops**: Detailed packet drop events
- **analyze_patterns**: Connection pattern analysis and behavioral insights

### AI-Powered Tools
- **contextual_analysis**: Advanced AI analysis with automatic tool selection
- **ai_insights**: Generate insights from provided summary text

## 📊 Sample Output

### Intelligent Analysis
```bash
### Contextual Analysis

netspy-mcp> contextual "analyze my system"

### Comprehensive Network Analysis

1. **Network Summary**:
   - Over the last 60 seconds, there were **10 outbound connection attempts** recorded.

2. **Detailed Connection Events**:
   - A total of **29 connection events** were logged recently.
   - Notable connections include:
     - **10 connections** to `127.0.0.1:8080` by the process `netspy`.
     - Additional connections to DNS servers and local processes.

3. **Packet Drops**:
   - There have been **3 packet drops** across all processes in the last 60 seconds.

4. **Connection Patterns**:
   - **Top Destinations**: Local connections dominate with 8 connections to `:0`
   - **Protocols Used**: Predominantly TCP (9 connections) and UDP (12 connections)

### Insights & Recommendations:
- **Monitor `netspy`**: Heavy localhost usage detected
- **Investigate Packet Drops**: Monitor for recurring losses
- **Consider Network Capacity**: Optimize settings if under heavy load
- **Regular Monitoring**: Implement ongoing monitoring for these metrics
```

### Traditional Commands
```bash
netspy-mcp> summary --process curl
Process 'curl' made 5 outbound connection attempts over the last 60 seconds

netspy-mcp> list --max-events 5
Recent connection events (15 total):
  21:05:53 | 127.0.0.1:8080 | TCP | netspy
  21:03:38 | (local socket) | UNIX | snapd
  21:01:08 | (local socket) | UNIX | snapd
  20:58:42 | 192.168.120.2:53 | UDP | systemd-resolve
  20:56:15 | 172.217.164.78:443 | TCP | curl
```

## 🎛️ Command Line Options

### General Options
- `--server URL`: eBPF server URL (default: http://localhost:8080)
- `--verbose`: Enable verbose logging
- `--help`: Show help information

### Tool Execution
- `--tool TOOL`: Run specific MCP tool and exit
  - Available tools: `get_network_summary`, `list_connections`, `get_packet_drop_summary`, `list_packet_drops`, `analyze_patterns`, `ai_insights`, `contextual_analysis`

### Tool Parameters
- `--pid PID`: Process ID to monitor
- `--process NAME`: Process name to monitor
- `--duration SECONDS`: Duration in seconds (default: 60)
- `--max-events COUNT`: Maximum events to retrieve (default: 100)
- `--summary-text TEXT`: Summary text for AI insights
- `--query TEXT`: Natural language query for contextual analysis

## 🤖 AI Function Calling Details

### How It Works

The system automatically registers all MCP tools as OpenAI functions, enabling the LLM to:
- **Automatically select relevant tools** based on user queries
- **Chain multiple tools** for comprehensive analysis
- **Validate parameters** and handle errors gracefully
- **Maintain conversation context** across interactions

### Function Parameters

Each tool accepts these parameters (all optional unless specified):

**Network Analysis Functions:**
- `pid` (integer): Process ID to analyze
- `process_name` (string): Process name to analyze  
- `duration` (integer, default: 60): Duration in seconds to analyze
- `max_events` (integer, default: 10): Maximum number of events to return

**AI Functions:**
- `query` (string, required for contextual_analysis): Natural language query
- `summary_text` (string, required for ai_insights): Summary text to analyze

### Key Improvements

1. **No Tool Context Integration** → **Live Tool Access**
   - ✅ LLM has direct access to live network data through function calls

2. **Missing Function Calling** → **Full OpenAI Function Support**
   - ✅ Proper OpenAI function calling with parameter validation
   - ✅ All 7 MCP tools registered as OpenAI functions

3. **Static Approach** → **Dynamic Tool Usage**
   - ✅ LLM contextually selects and chains multiple tools
   - ✅ Demonstrated: 4 tools used automatically for system analysis

4. **Limited Context** → **Structured Context Management**
   - ✅ Multi-turn conversations with full context retention

## 🔧 Troubleshooting

**"Connection refused" or failed connection**
- Ensure the eBPF API server is running with `--http --port 8080`
- Test server health: `curl http://localhost:8080/health`
- Check that nothing else is using port 8080

**"No connections found" when server has data**
- Ensure the eBPF API server is running and generating data
- Generate some traffic: `curl -s http://google.com`
- Use `netstat -tlnp | grep 8080` to verify server is listening

**"permission denied"**  
- Use `sudo` for the eBPF server - eBPF operations require root privileges
- netspy client can run without sudo when using HTTP API mode

**OpenAI API errors**
- Check that `OPENAI_API_KEY` environment variable is set
- Verify API key is valid and has sufficient credits
- Check rate limits if experiencing frequent failures

**"open bpf/connection.o: no such file or directory"**
- Build server with `make build` in ebpf-server repository
- Server needs compiled eBPF programs, not just Go binary

## 👨‍💻 Developer Usage

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

### OpenAI Integration
```go
import "github.com/srodi/netspy/internal/openai"

// Create contextual network analyst
analyst := openai.NewContextualNetworkAnalyst(mcpExecutor)

// Analyze with natural language
analysis, err := analyst.AnalyzeNetworkQuery(ctx, "What's happening with my network?")

// Process-specific analysis
analysis, err := analyst.AnalyzeProcess(ctx, "nginx", 0, 60)
```

## 🏗️ Architecture Details

### Core Components

1. **Function Call Manager** (`internal/openai/functions.go`)
   - Automatic MCP tool discovery and registration
   - OpenAI function definition generation
   - Parameter validation and type conversion
   - Error handling for function execution

2. **Conversation Manager** (`internal/openai/client.go`)
   - Multi-turn conversation management
   - Automatic function call detection and execution
   - Tool result integration back into conversation
   - Using `gpt-4o-mini` for optimal function calling

3. **Intelligent Network Analyst** (`internal/openai/analyst.go`)
   - Context-aware query processing
   - Automatic tool selection and chaining
   - Specialized network analysis prompting
   - Conversation history management

4. **MCP Server** (`internal/mcp/server.go`)
   - Model Context Protocol implementation
   - Tool registration and execution
   - HTTP communication with eBPF server

## 🔮 Future Enhancements

The architecture supports easy extension for:
- Additional MCP tools (automatically discovered)
- Custom analysis workflows
- Different AI model providers
- Streaming responses
- Function call caching
- Multi-modal inputs (images, files)

## 🚀 Performance

- **Function Call Latency**: Sub-second tool execution
- **Tool Coverage**: 100% of MCP tools available to LLM
- **Model Efficiency**: Using `gpt-4o-mini` for optimal function calling
- **Zero Code Duplication**: Automatic tool discovery eliminates duplicate definitions

## 📜 License

This project is part of the network telemetry ecosystem and integrates with the [ebpf-server](https://github.com/SRodi/ebpf-server) for comprehensive network monitoring capabilities.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/srodi/mcp-client/internal/mcp"
	"github.com/srodi/mcp-client/internal/openai"
	"github.com/srodi/mcp-client/internal/utils"
)

func main() {
	pid := flag.Int("pid", 0, "PID to analyze")
	processName := flag.String("process", "", "Process name to analyze (e.g., 'curl', 'ssh')")
	duration := flag.Int("duration", 60, "Duration in seconds to analyze (default: 60)")
	serverURL := flag.String("server", "http://localhost:8080/mcp", "MCP server URL")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	listConnections := flag.Bool("list", false, "List recent connection events")
	maxEvents := flag.Int("max-events", 10, "Maximum number of events to show when listing")
	analyzePatterns := flag.Bool("analyze", false, "Analyze connection patterns")
	flag.Parse()

	// Validate input - require either PID or process name (unless just listing all connections)
	if !*listConnections && *pid == 0 && *processName == "" {
		fmt.Println("Usage: mcp-client [options]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  mcp-client --pid 1234                    # Analyze specific PID")
		fmt.Println("  mcp-client --process curl                # Analyze all curl processes")
		fmt.Println("  mcp-client --process ssh --duration 300  # Analyze SSH for 5 minutes")
		fmt.Println("  mcp-client --list                        # List all recent connections")
		fmt.Println("  mcp-client --process curl --list         # List curl connections")
		fmt.Println("  mcp-client --process nginx --analyze     # Analyze nginx patterns")
		os.Exit(1)
	}

	if *pid != 0 && *processName != "" {
		log.Fatal("Please specify either --pid OR --process, not both")
	}

	// Handle list connections mode
	if *listConnections {
		connections, err := mcp.ListConnections(*serverURL)
		if err != nil {
			log.Fatalf("Failed to list connections: %v", err)
		}

		if len(connections.Result) == 0 {
			fmt.Println("No connections found")
			return
		}

		// Filter connections if specific PID or process requested
		if *pid != 0 || *processName != "" {
			var filteredEvents []mcp.ConnectionEvent
			for _, events := range connections.Result {
				for _, event := range events {
					if *pid != 0 && event.PID == uint32(*pid) {
						filteredEvents = append(filteredEvents, event)
					} else if *processName != "" && event.Command == *processName {
						filteredEvents = append(filteredEvents, event)
					}
				}
			}

			if len(filteredEvents) == 0 {
				var target string
				if *pid != 0 {
					target = fmt.Sprintf("PID %d", *pid)
				} else {
					target = fmt.Sprintf("process '%s'", *processName)
				}
				fmt.Printf("No connections found for %s\n", target)
				return
			}

			fmt.Println(utils.FormatConnectionEvents(filteredEvents, *maxEvents))
			if *analyzePatterns {
				fmt.Println()
				fmt.Println(utils.AnalyzeConnectionPatterns(filteredEvents))
			}
		} else {
			// Show all connections
			var allEvents []mcp.ConnectionEvent
			for _, events := range connections.Result {
				allEvents = append(allEvents, events...)
			}
			fmt.Println(utils.FormatConnectionEvents(allEvents, *maxEvents))
			if *analyzePatterns {
				fmt.Println()
				fmt.Println(utils.AnalyzeConnectionPatterns(allEvents))
			}
		}
		return
	}

	// Prepare request parameters for summary
	var queryType, queryValue string
	if *pid != 0 {
		queryType = "PID"
		queryValue = fmt.Sprintf("%d", *pid)
	} else {
		queryType = "process"
		queryValue = *processName
	}

	if *verbose {
		fmt.Printf("Querying MCP server at %s\n", *serverURL)
		fmt.Printf("Analyzing %s: %s for %d seconds\n", queryType, queryValue, *duration)
		fmt.Println()
	}

	// Get connection summary from MCP server
	summary, err := mcp.GetConnectionSummary(*pid, *processName, *duration, *serverURL)
	if err != nil {
		log.Fatalf("Failed to get summary: %v", err)
	}

	// Format the summary text using utility function
	text := utils.FormatConnectionSummary(*pid, *processName, *duration, summary)

	fmt.Println("🔍 Network Telemetry Summary:")
	fmt.Println(text)

	// Show additional analysis if requested and there's data
	if *analyzePatterns && summary.Result.Total > 0 {
		// Need to get detailed connection data for pattern analysis
		connections, err := mcp.ListConnections(*serverURL)
		if err == nil {
			var filteredEvents []mcp.ConnectionEvent
			for _, events := range connections.Result {
				for _, event := range events {
					if *pid != 0 && event.PID == uint32(*pid) {
						filteredEvents = append(filteredEvents, event)
					} else if *processName != "" && event.Command == *processName {
						filteredEvents = append(filteredEvents, event)
					}
				}
			}
			if len(filteredEvents) > 0 {
				fmt.Println()
				fmt.Println(utils.AnalyzeConnectionPatterns(filteredEvents))
			}
		}
	}

	// Only call OpenAI if we have connection data
	if summary.Result.Total > 0 {
		answer, err := openai.AskLLM(text)
		if err != nil {
			if *verbose {
				fmt.Printf("\n⚠️  OpenAI insights unavailable: %v\n", err)
				fmt.Println("(Set OPENAI_API_KEY environment variable to enable AI-powered network insights)")
			}
		} else {
			fmt.Println("\n🤖 AI Network Insights:")
			fmt.Println(answer)
		}
	} else {
		fmt.Printf("\n💡 No connection attempts found for %s '%s' in the last %d seconds.\n", queryType, queryValue, *duration)
		if queryType == "PID" {
			fmt.Println("   - Check if the PID is correct and the process is making network connections")
			fmt.Println("   - Try using --process instead if the process name is known")
		} else {
			fmt.Println("   - Check if the process name is correct (case-sensitive)")
			fmt.Println("   - Ensure the process has made network connections recently")
			fmt.Println("   - Try increasing --duration for historical data")
		}
		fmt.Println("   - Use --list to see all current connections")
	}
}

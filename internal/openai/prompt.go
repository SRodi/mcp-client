package openai

import "fmt"

// CreateNetworkInsightsPrompt generates a prompt focused on network connection pattern analysis
func CreateNetworkInsightsPrompt(summary string) string {
	return fmt.Sprintf(`You are a network connectivity analyst. Given this network connection data, provide insights about the connection patterns and behavior:

%s

Focus on:
- Connection frequency and patterns
- Network behavior assessment (is this normal for this type of process?)
- Destination analysis and what it suggests about the application
- Optimization opportunities for connection efficiency
- Potential monitoring or architecture recommendations

Be technical but practical. This data shows connection attempts, not network performance metrics. Provide actionable insights about the application's network behavior.`, summary)
}

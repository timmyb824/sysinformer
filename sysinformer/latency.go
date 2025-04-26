package sysinformer

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

const LATENCY_TIMEOUT = 3 * time.Second

var hosts = []string{
	"github.com",
	"google.com",
	"cloudflare.com",
	"amazon.com",
	"microsoft.com",
}

func checkPing(host string) (string, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), LATENCY_TIMEOUT)
	defer cancel()

	// Try curl first as it's more reliable
	curlCmd := exec.CommandContext(ctx, "curl", "-o", "/dev/null", "-s", "-w", "%{time_total}", fmt.Sprintf("https://%s", host))
	curlOutput, err := curlCmd.Output()
	if err == nil {
		latency, err := strconv.ParseFloat(strings.TrimSpace(string(curlOutput)), 64)
		if err == nil {
			return fmt.Sprintf("%.2f ms", latency*1000), nil // Convert seconds to milliseconds
		}
	}

	// If curl fails, try ping
	cmd := exec.CommandContext(ctx, "ping", "-c", "1", "-W", "2", host)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "Timeout", err
	}

	// Parse ping output to get round-trip time
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "time=") {
			parts := strings.Split(line, "time=")
			if len(parts) > 1 {
				timeStr := strings.Split(parts[1], " ")[0]
				return fmt.Sprintf("%s ms", timeStr), nil
			}
		}
	}

	return "Timeout", fmt.Errorf("could not parse ping output")
}

func calculateAverageLatency(pingResults [][]string) float64 {
	var total float64
	var count int

	for _, result := range pingResults {
		if len(result) == 2 && result[1] != "Timeout" {
			latencyStr := strings.TrimSuffix(result[1], " ms")
			latency, err := strconv.ParseFloat(latencyStr, 64)
			if err == nil && latency > 0 {
				total += latency
				count++
			}
		}
	}

	if count == 0 {
		return 0
	}
	return total / float64(count)
}

func performPing(hosts []string) [][]string {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var pingResults [][]string

	// Create a channel to collect results
	resultChan := make(chan []string, len(hosts))

	// Start a goroutine for each host
	for _, host := range hosts {
		wg.Add(1)
		go func(h string) {
			defer wg.Done()
			latency, err := checkPing(h)
			if err != nil {
				latency = "Timeout"
			}
			resultChan <- []string{h, latency}
		}(host)
	}

	// Close result channel when all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results as they come in
	for result := range resultChan {
		if result[1] != "Timeout" {
			mu.Lock()
			pingResults = append(pingResults, result)
			mu.Unlock()
		}
	}

	return pingResults
}

func PrintLatencyInfo() {
	fmt.Println("") // Add space before section
	// headers and data preparation logic remains

	PrintSectionHeader("===== Latency Information =====")

	// Perform ping tests concurrently
	pingResults := performPing(hosts)

	// Print warning if no results
	if len(pingResults) == 0 {
		fmt.Println("Warning: No latency information available (all hosts timed out)")
		return
	}

	// Calculate average latency
	avgLatency := calculateAverageLatency(pingResults)

	headers := []string{"Host", "Latency (ms)"}
	var data [][]string
	for _, result := range pingResults {
		row := []string{result[0], result[1]}
		data = append(data, row)
	}
	RenderTable(headers, data)
	fmt.Printf("Average Round-Trip Delay: %.2f ms\n", avgLatency)

}

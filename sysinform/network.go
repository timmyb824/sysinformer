package sysinform

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	GET_WAN_IP = "https://api.ipify.org"
	NETWORK_TIMEOUT = 3 * time.Second
)

func getNetworkInfo() (map[string]string, [2]string, error) {
	ipLanDict := make(map[string]string)

	// Get network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, [2]string{}, err
	}

	for _, iface := range interfaces {
		// Skip loopback interface
		if iface.Name == "lo" || iface.Name == "lo0" {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipv4 := ipNet.IP.To4(); ipv4 != nil {
					ipLanDict[iface.Name] = ipv4.String()
				}
			}
		}
	}

	// Get WAN IP with timeout
	ipWan := "N/A"
	client := &http.Client{Timeout: NETWORK_TIMEOUT}
	resp, err := client.Get(GET_WAN_IP)
	if err == nil {
		defer resp.Body.Close()
		body, err := io.ReadAll(io.LimitReader(resp.Body, 1024)) // Limit read to 1KB
		if err == nil {
			ipWan = strings.TrimSpace(string(body))
		}
	}

	return ipLanDict, [2]string{"WAN", ipWan}, nil
}

func getNetworkActivity() (map[string]map[string]float64, error) {
	networkActivity := make(map[string]map[string]float64)

	// Get network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		// Skip loopback interface
		if iface.Name == "lo" || iface.Name == "lo0" {
			continue
		}

		// Use netstat to get bytes sent/received with timeout
		ctx, cancel := context.WithTimeout(context.Background(), NETWORK_TIMEOUT)
		defer cancel()
		cmd := exec.CommandContext(ctx, "netstat", "-I", iface.Name, "-b")
		output, err := cmd.Output()
		if err != nil {
			// If command times out or fails, set zeros for this interface
			networkActivity[iface.Name] = map[string]float64{
				"bytes_sent": 0,
				"bytes_recv": 0,
			}
			continue
		}

		// Parse netstat output
		lines := strings.Split(string(output), "\n")
		if len(lines) < 2 {
			continue
		}

		// Split the second line into fields
		fields := strings.Fields(lines[1])
		if len(fields) < 10 {
			continue
		}

		// Fields[6] is bytes received, Fields[9] is bytes sent
		bytesRecv, err := strconv.ParseFloat(fields[6], 64)
		if err != nil {
			bytesRecv = 0
		}

		bytesSent, err := strconv.ParseFloat(fields[9], 64)
		if err != nil {
			bytesSent = 0
		}

		networkActivity[iface.Name] = map[string]float64{
			"bytes_sent": bytesSent,
			"bytes_recv": bytesRecv,
		}
	}

	return networkActivity, nil
}

func PrintNetworkInfo() {
	fmt.Println("===== Network Information =====")

	// Get network info with error handling
	ipLan, ipWan, err := getNetworkInfo()
	if err != nil {
		fmt.Printf("Warning: Limited network info available (%v)\n", err)
		ipLan = make(map[string]string)
		ipWan = [2]string{"WAN", "N/A"}
	}

	// Get network activity with error handling
	networkActivity, err := getNetworkActivity()
	if err != nil {
		fmt.Printf("Warning: Limited network activity info available (%v)\n", err)
		networkActivity = make(map[string]map[string]float64)
	}

	// Create slice for table rows
	var table [][]string

	// Add LAN interfaces to table
	for iface, ip := range ipLan {
		bytesSent := networkActivity[iface]["bytes_sent"]
		bytesRecv := networkActivity[iface]["bytes_recv"]

		// Convert bytes to MB
		mbSent := fmt.Sprintf("%.2f", bytesSent/(1024*1024))
		mbRecv := fmt.Sprintf("%.2f", bytesRecv/(1024*1024))

		table = append(table, []string{iface, ip, mbSent, mbRecv})
	}

	// Add WAN to table
	table = append(table, []string{ipWan[0], ipWan[1], "-", "-"})

	// Sort table by interface name
	sort.Slice(table, func(i, j int) bool {
		return table[i][0] < table[j][0]
	})

	// Print table header
	fmt.Println("┌─────────────┬────────────────┬───────────┬───────────────┐")
	fmt.Println("│ Interface   │ IP             │ MB Sent   │ MB Received   │")
	fmt.Println("├─────────────┼────────────────┼───────────┼───────────────┤")

	// Print table rows
	for i, row := range table {
		fmt.Printf("│ %-9s │ %-12s │ %-7s │ %-11s │\n",
			row[0], row[1], row[2], row[3])
		if i < len(table)-1 {
			fmt.Println("├─────────────┼────────────────┼───────────┼───────────────┤")
		}
	}

	fmt.Println("└─────────────┴────────────────┴───────────┴───────────────┘")
}

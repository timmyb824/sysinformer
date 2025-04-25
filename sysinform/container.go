package sysinform

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const CONTAINER_TIMEOUT = 3 * time.Second

type Container struct {
	ID       string
	Name     string
	Image    string
	Command  string
	Created  string
	Status   string
	Ports    string
	Platform string // "docker" or "podman"
}

func checkContainerRuntime() string {
	// Try docker first
	ctx, cancel := context.WithTimeout(context.Background(), CONTAINER_TIMEOUT)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "--version")
	if err := cmd.Run(); err == nil {
		return "docker"
	}

	// Try podman
	cmd = exec.CommandContext(ctx, "podman", "--version")
	if err := cmd.Run(); err == nil {
		return "podman"
	}

	return ""
}

func getContainers() ([]Container, error) {
	platform := checkContainerRuntime()
	if platform == "" {
		return nil, fmt.Errorf("no container runtime found (docker or podman)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), CONTAINER_TIMEOUT)
	defer cancel()

	// Get running containers with format string
	format := "{{.ID}}\t{{.Names}}\t{{.Image}}\t{{.Command}}\t{{.CreatedAt}}\t{{.Status}}\t{{.Ports}}"
	cmd := exec.CommandContext(ctx, platform, "ps", "--format", format)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error getting container list: %v", err)
	}

	var containers []Container
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 5 {
			continue
		}

			// Parse ports
		ports := fields[6]
		if ports == "" {
			ports = "N/A"
		} else {
			// Extract port mappings using regex
			re := regexp.MustCompile(`0.0.0.0:(\d+)->(\d+)/tcp`)
			matches := re.FindStringSubmatch(ports)
			if len(matches) > 2 {
				ports = fmt.Sprintf("%s:%s", matches[1], matches[2])
			}
		}

		container := Container{
			ID:       fields[0],
			Name:     fields[1],
			Image:    fields[2],
			Command:  fields[3],
			Created:  fields[4],
			Status:   fields[5],
			Ports:    ports,
			Platform: platform,
		}
		containers = append(containers, container)
	}

	return containers, nil
}

func PrintContainerInfo() {
	fmt.Println("===== Container Information =====")

	containers, err := getContainers()
	if err != nil {
		fmt.Printf("Warning: %v\n", err)
		return
	}

	if len(containers) == 0 {
		fmt.Println("No running containers found")
		return
	}

	// Print runtime information
	fmt.Printf("%s containers running on your system:\n", strings.Title(containers[0].Platform))

	// Print table header
	fmt.Println("┌───────────┬───────────┬───────────┬───────────┐")
	fmt.Println("│ ID         │ Name       │ Ports      │ Status     │")
	fmt.Println("├───────────┼───────────┼───────────┼───────────┤")

	// Print table rows
	for i, container := range containers {
		// Truncate ID to first 12 characters
		id := container.ID
		if len(id) > 12 {
			id = id[:12]
		}

		// Extract status duration
		status := container.Status
		if strings.HasPrefix(status, "Up") {
			parts := strings.Fields(status)
			if len(parts) > 1 {
				status = strings.Join(parts[1:], " ")
			}
		}

		fmt.Printf("│ %-10s │ %-10s │ %-10s │ %-10s │\n",
			id,
			truncate(container.Name, 10),
			truncate(container.Ports, 10),
			truncate(status, 10))

		if i < len(containers)-1 {
			fmt.Println("├───────────┼───────────┼───────────┼───────────┤")
		}
	}

	fmt.Println("└───────────┴───────────┴───────────┴───────────┘")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

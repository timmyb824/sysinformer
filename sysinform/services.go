package sysinform

import (
	"fmt"
	"net"
	"time"
)

type Service struct {
	Name string
	Port int
}

var commonServices = []Service{
	{"FTP", 21},
	{"SSH", 22},
	{"HTTP", 80},
	{"HTTPS", 443},
	{"MySQL", 3306},
	{"PostgreSQL", 5432},
	{"Redis", 6379},
	{"MongoDB", 27017},
	{"HTTP-Alt", 8080},
	{"SQL Server", 1433},
}

func checkService(port int) bool {
	// Try to connect to localhost with a 1 second timeout
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func PrintServicesInfo() {
	fmt.Println("===== Services Information =====")

	// Print table header
	fmt.Println("┌──────────────┬──────────┬──────────┐")
	fmt.Println("│ Service Name │ Port     │ Status   │")
	fmt.Println("├──────────────┼──────────┼──────────┤")

	// Check each service
	for i, service := range commonServices {
		status := "Down"
		if checkService(service.Port) {
			status = "\033[92mUp\033[0m"  // Green for Up
		} else {
			status = "\033[91mDown\033[0m" // Red for Down
		}

		fmt.Printf("│ %-11s │ %-8d │ %-8s │\n", service.Name, service.Port, status)

		// Print separator line except for the last row
		if i < len(commonServices)-1 {
			fmt.Println("├──────────────┼──────────┼──────────┤")
		}
	}

	fmt.Println("└──────────────┴──────────┴──────────┘")
}

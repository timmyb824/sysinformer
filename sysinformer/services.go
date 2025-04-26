package sysinformer

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
	fmt.Println("") // Add space before section
	// headers and data preparation logic remains

	PrintSectionHeader("===== Services Information =====")
	headers := []string{"Service Name", "Port", "Status"}
	var data [][]string
	for _, service := range commonServices {
		status := "\033[91mDown\033[0m" // Red for Down
		if checkService(service.Port) {
			status = "\033[92mUp\033[0m" // Green for Up
		}
		row := []string{service.Name, fmt.Sprintf("%d", service.Port), status}
		data = append(data, row)
	}
	RenderTable(headers, data)

	fmt.Println("└──────────────┴──────────┴──────────┘")
}

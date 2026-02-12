package sysinformer

import (
	"fmt"
	"runtime"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/shirou/gopsutil/v3/host"
)

func getSystemInfo() (map[string]interface{}, error) {
	hostInfo, err := host.Info()
	if err != nil {
		return nil, err
	}

	users, err := host.Users()
	usersNb := 0
	if err == nil {
		usersNb = len(users)
	}

	var dist string
	switch runtime.GOOS {
	case "darwin":
		dist = "macOS"
	case "linux":
		dist = hostInfo.Platform
	case "windows":
		dist = "Windows"
	default:
		dist = hostInfo.Platform
	}

	systemInfo := map[string]interface{}{
		"os_type":        cases.Title(language.English).String(runtime.GOOS),
		"hostname":       hostInfo.Hostname,
		"kernel_info":    hostInfo.KernelVersion,
		"architecture":   runtime.GOARCH,
		"dist":           dist,
		"dist_version":   hostInfo.PlatformVersion,
		"uptime":         fmt.Sprintf("%d days, %d hours, %d minutes", hostInfo.Uptime/86400, (hostInfo.Uptime%86400)/3600, (hostInfo.Uptime%3600)/60),
		"last_boot_date": time.Unix(int64(hostInfo.BootTime), 0).Format("2006-01-02 15:04:05"),
		"users_nb":       usersNb,
		"current_date":   time.Now().Format("2006-01-02 15:04:05"),
	}
	return systemInfo, nil
}

func PrintSystemInfo() {
	systemInfo, err := getSystemInfo()
	if err != nil {
		fmt.Println("Error getting system info:", err)
		return
	}
	PrintSectionHeader("===== System Information =====")
	fmt.Printf("Hostname: %v\n", systemInfo["hostname"])
	fmt.Printf("OS: %v %v %v\n", systemInfo["os_type"], systemInfo["dist"], systemInfo["dist_version"])
	fmt.Printf("Kernel: %v\n", systemInfo["kernel_info"])
	fmt.Printf("Architecture: %v\n", systemInfo["architecture"])
	fmt.Printf("Uptime: %v\n", systemInfo["uptime"])
	fmt.Printf("Last boot: %v\n", systemInfo["last_boot_date"])
	fmt.Printf("Users: %v\n", systemInfo["users_nb"])
	fmt.Printf("Server datetime: %v\n", systemInfo["current_date"])
}

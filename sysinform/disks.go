package sysinform

import (
	"fmt"
	"strings"

	"github.com/shirou/gopsutil/v3/disk"
)

func getDiskSpace() ([]map[string]interface{}, error) {
	partitions, err := disk.Partitions(false) // false means physical devices only
	if err != nil {
		return nil, err
	}

	var diskSpaceInfo []map[string]interface{}

	for _, partition := range partitions {
		// Skip snap-related mountpoints
		if strings.HasPrefix(partition.Mountpoint, "/snap") || strings.HasPrefix(partition.Mountpoint, "/var/snap") {
			continue
		}

		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue // Skip this partition if we can't get usage info
		}

		// Convert bytes to GB
		total := float64(usage.Total) / (1024 * 1024 * 1024)
		used := float64(usage.Used) / (1024 * 1024 * 1024)
		free := float64(usage.Free) / (1024 * 1024 * 1024)

		diskSpaceInfo = append(diskSpaceInfo, map[string]interface{}{
			"device":     partition.Device,
			"mountpoint": partition.Mountpoint,
			"total":      total,
			"used":       used,
			"free":       free,
			"percentage": usage.UsedPercent,
		})
	}

	return diskSpaceInfo, nil
}

func PrintDiskInfo() {
	diskInfo, err := getDiskSpace()
	if err != nil {
		fmt.Println("Error getting disk info:", err)
		return
	}

	fmt.Println("===== Disk Information =====")
	fmt.Println("┌──────────────┬──────────────┬────────────┬────────────┬────────────┬────────────┐")
	fmt.Println("│ Device       │ Mountpoint   │ Total      │ Used       │ Free       │ Percentage │")
	fmt.Println("├──────────────┼──────────────┼────────────┼────────────┼────────────┼────────────┤")

	for i, info := range diskInfo {
		fmt.Printf("│ %-10s │ %-10s │ %8.2f GB │ %8.2f GB │ %8.2f GB │ %8.2f %% │\n",
			info["device"],
			info["mountpoint"],
			info["total"],
			info["used"],
			info["free"],
			info["percentage"])

		if i < len(diskInfo)-1 {
			fmt.Println("├──────────────┼──────────────┼────────────┼────────────┼────────────┼────────────┤")
		}
	}

	fmt.Println("└──────────────┴──────────────┴────────────┴────────────┴────────────┴────────────┘")
}

package sysinformer

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
	fmt.Println("") // Add space before section
	// headers and data preparation logic remains

	diskInfo, err := getDiskSpace()
	if err != nil {
		fmt.Println("Error getting disk info:", err)
		return
	}

	PrintSectionHeader("===== Disk Information =====")
	headers := []string{"Device", "Mountpoint", "Total", "Used", "Free", "Percentage"}
	var data [][]string
	for _, info := range diskInfo {
		row := []string{
			info["device"].(string),
			info["mountpoint"].(string),
			fmt.Sprintf("%.2f", info["total"].(float64)),
			fmt.Sprintf("%.2f", info["used"].(float64)),
			fmt.Sprintf("%.2f", info["free"].(float64)),
			fmt.Sprintf("%.2f", info["percentage"].(float64)),
		}
		data = append(data, row)
	}
	RenderTable(headers, data)
}

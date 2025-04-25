package sysinformer

import (
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/process"
)

type procInfo struct {
	pid           int32
	name          string
	cpuPercent    float64
	memoryPercent float32
}

func getCPUInfo() (map[string]interface{}, error) {
	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, err
	}

	cpuStats, err := cpu.Percent(0, false)
	if err != nil {
		return nil, err
	}

	loadAvg, err := load.Avg()
	if err != nil {
		loadAvg = &load.AvgStat{} // Use empty stats if error
	}

	cpuCount, err := cpu.Counts(true)
	if err != nil {
		cpuCount = 0
	}

	// Get CPU cache size for macOS
	cpuCache := "UNKNOWN"
	cpuBogomips := "N/A (No macOS Equivalent)"
	if runtime.GOOS == "darwin" {
		out, err := exec.Command("sysctl", "-n", "hw.l2cachesize").Output()
		if err == nil {
			cacheSize := strings.TrimSpace(string(out))
			cpuCache = fmt.Sprintf("%d KB", stringToInt(cacheSize)/1024)
		}
	}

	var cpuModel string
	if len(cpuInfo) > 0 {
		cpuModel = cpuInfo[0].ModelName
	} else {
		cpuModel = "UNKNOWN"
	}

	var cpuFreq float64
	if len(cpuInfo) > 0 {
		cpuFreq = float64(cpuInfo[0].Mhz)
	}

	info := map[string]interface{}{
		"cpu_count":    cpuCount,
		"cpu_model":    cpuModel,
		"cpu_freq":     cpuFreq,
		"cpu_cache":    cpuCache,
		"cpu_bogomips": cpuBogomips,
		"cpu_usage":    cpuStats[0],
		"load_1":       loadAvg.Load1,
		"load_5":       loadAvg.Load5,
		"load_15":      loadAvg.Load15,
	}

	// Get process information
	processes, err := process.Processes()
	if err == nil {
			// Create a slice to store process info
		var procs []procInfo

		// Collect info for each process
		for _, p := range processes {
			name, err := p.Name()
			if err != nil {
			continue
			}

			cpu, err := p.CPUPercent()
			if err != nil {
			continue
			}

			mem, err := p.MemoryPercent()
			if err != nil {
			continue
			}

			procs = append(procs, procInfo{
				pid:           p.Pid,
				name:          name,
				cpuPercent:    cpu,
				memoryPercent: mem,
			})
		}

		// Sort by CPU usage
		sort.Slice(procs, func(i, j int) bool {
			return procs[i].cpuPercent > procs[j].cpuPercent
		})

		// Get top 5 processes
		top5 := procs
		if len(top5) > 5 {
			top5 = top5[:5]
		}

		info["process_count"] = len(processes)
		info["top_processes"] = top5
	}

	return info, nil
}

func stringToInt(s string) int64 {
	var n int64
	fmt.Sscanf(s, "%d", &n)
	return n
}

func PrintCPUInfo() {
	fmt.Println("") // Add space before section
	// headers and data preparation logic remains

	cpuInfo, err := getCPUInfo()
	if err != nil {
		fmt.Println("Error getting CPU info:", err)
		return
	}

	fmt.Println("===== CPU Information =====")
	headers := []string{"Model", "Cores", "Speed", "Usage"}
	var data [][]string
	data = append(data, []string{fmt.Sprintf("%v", cpuInfo["cpu_model"]), fmt.Sprintf("%v", cpuInfo["cpu_count"]), fmt.Sprintf("%.2f", cpuInfo["cpu_freq"]), fmt.Sprintf("%.1f%%", cpuInfo["cpu_usage"])})
	RenderTable(headers, data)
	fmt.Printf("Cache Size: %v\n", cpuInfo["cpu_cache"])
	fmt.Printf("BogoMips: %v\n", cpuInfo["cpu_bogomips"])
	if count, ok := cpuInfo["process_count"].(int); ok {
		fmt.Printf("Process Count: %d\n", count)
	}
	fmt.Printf("Load Average: %.2f, %.2f, %.2f (1, 5, 15 min)\n",
		cpuInfo["load_1"], cpuInfo["load_5"], cpuInfo["load_15"])

	// Print top processes
	if top5, ok := cpuInfo["top_processes"].([]procInfo); ok && len(top5) > 0 {
		fmt.Println("\nTop 5 Processes by CPU Usage:")
		headers := []string{"pid", "name", "cpu_percent", "memory_percent"}
		var data [][]string
		for _, proc := range top5 {
			row := []string{
				fmt.Sprintf("%d", proc.pid),
				proc.name,
				fmt.Sprintf("%.1f", proc.cpuPercent),
				fmt.Sprintf("%.6f", proc.memoryPercent),
			}
			data = append(data, row)
		}
		RenderTable(headers, data)
	}
}

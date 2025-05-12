package sysinformer

import (
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

func formatMemoryValue(valueInMB float64) string {
	if valueInMB > 1000 {
		return fmt.Sprintf("%.2fGB", valueInMB/1024)
	}
	return fmt.Sprintf("%.0fMB", valueInMB)
}

func getTotalMemoryOfAllProcesses() float64 {
	processes, err := process.Processes()
	if err != nil {
		return 0
	}

	var totalMemory float64
	for _, p := range processes {
		memInfo, err := p.MemoryInfo()
		if err != nil || memInfo == nil {
			continue
		}
		// Convert bytes to MB
		totalMemory += float64(memInfo.RSS) / (1024 * 1024)
	}
	return totalMemory
}

func getMemoryInfo() (map[string]interface{}, error) {
	virtualMemory, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	swapMemory, err := mem.SwapMemory()
	if err != nil {
		return nil, err
	}

	var memTotal, memFree float64
	var warning string

	if runtime.GOOS == "darwin" {
		// Get total physical memory
		out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
		if err == nil {
			totalBytes, _ := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
			memTotal = float64(totalBytes) / (1024 * 1024) // Convert bytes to MB
		}

		// Get memory usage from processes
		memUsageProcess := getTotalMemoryOfAllProcesses()
		memFree = memTotal - memUsageProcess

		warning = "WARNING: calc memory usage derived from process info; sys usage from psutil"
	} else {
		memTotal = float64(virtualMemory.Total) / (1024 * 1024)
		memFree = float64(virtualMemory.Available) / (1024 * 1024)
		warning = "WARNING: memory usage derived from '/proc/meminfo' and psutil"
	}

	// Get swap information
	var swapTotal, swapFree float64
	if runtime.GOOS == "darwin" {
		out, err := exec.Command("sysctl", "-n", "vm.swapusage").Output()
		if err == nil {
			re := regexp.MustCompile(`(\w+) = (\d+\.?\d*)M`)
			matches := re.FindAllStringSubmatch(string(out), -1)
			for _, match := range matches {
				value, _ := strconv.ParseFloat(match[2], 64)
				switch match[1] {
				case "total":
					swapTotal = value
				case "free":
					swapFree = value
				}
			}
		}
	} else {
		swapTotal = float64(swapMemory.Total) / (1024 * 1024)
		swapFree = float64(swapMemory.Free) / (1024 * 1024)
	}

	// Calculate usage percentages
	memUsedPercentageCalc := 0.0
	if memTotal > 0 {
		memUsedPercentageCalc = ((memTotal - memFree) / memTotal) * 100
	}

	swapUsedPercentageCalc := 0.0
	if swapTotal > 0 {
		swapUsedPercentageCalc = ((swapTotal - swapFree) / swapTotal) * 100
	}

	// Get top processes by memory usage
	processes, err := process.Processes()
	if err == nil {
		// Create a slice to store process info
		var procs []procInfo
		for _, p := range processes {
			name, err := p.Name()
			if err != nil {
				continue
			}

			memPercent, err := p.MemoryPercent()
			if err != nil {
				continue
			}

			cpuPercent, err := p.CPUPercent()
			if err != nil {
				continue
			}

			procs = append(procs, procInfo{
				pid:           p.Pid,
				name:          name,
				cpuPercent:    cpuPercent,
				memoryPercent: memPercent,
			})
		}

		// Filter out 'sysinformer' or 'sysinfo' process
		var filteredProcs []procInfo
		for _, proc := range procs {
			if proc.name != "sysinformer" && proc.name != "sysinfo" {
				filteredProcs = append(filteredProcs, proc)
			}
		}

		// Sort by memory usage
		sort.Slice(filteredProcs, func(i, j int) bool {
			return filteredProcs[i].memoryPercent > filteredProcs[j].memoryPercent
		})

		// Get top 5 processes (excluding sysinformer)
		if len(filteredProcs) > 5 {
			filteredProcs = filteredProcs[:5]
		}

		return map[string]interface{}{
			"mem_total":               formatMemoryValue(memTotal),
			"mem_free":                formatMemoryValue(memFree),
			"mem_usage":               virtualMemory.UsedPercent,
			"mem_used_percentage_calc": memUsedPercentageCalc,
			"swap_total":              formatMemoryValue(swapTotal),
			"swap_free":               formatMemoryValue(swapFree),
			"swap_usage":              swapMemory.UsedPercent,
			"swap_used_percentage_calc": swapUsedPercentageCalc,
			"warning":                  warning,
			"top_processes":            filteredProcs,
		}, nil
	}

	return map[string]interface{}{
		"mem_total":               formatMemoryValue(memTotal),
		"mem_free":                formatMemoryValue(memFree),
		"mem_usage":               virtualMemory.UsedPercent,
		"mem_used_percentage_calc": memUsedPercentageCalc,
		"swap_total":              formatMemoryValue(swapTotal),
		"swap_free":               formatMemoryValue(swapFree),
		"swap_usage":              swapMemory.UsedPercent,
		"swap_used_percentage_calc": swapUsedPercentageCalc,
		"warning":                  warning,
	}, nil
}

func PrintMemoryInfo() {
	fmt.Println("") // Add space before section
	// headers and data preparation logic remains

	memInfo, err := getMemoryInfo()
	if err != nil {
		fmt.Println("Error getting memory info:", err)
		return
	}

	PrintSectionHeader("===== Memory Information =====")
fmt.Println("Note: 'Actual Usage %' is calculated as (1 - Available/Total) and matches htop-style memory usage (excludes cache/buffers reclaimed by the OS).")
	headers := []string{"Type", "Free", "Total", "Usage %", "Actual Usage %"}
	memRow := []string{
		"Mem",
		memInfo["mem_free"].(string),
		memInfo["mem_total"].(string),
		fmt.Sprintf("%.2f", memInfo["mem_usage"].(float64)), // Traditional usage (includes buffers/cache)
		fmt.Sprintf("%.2f", memInfo["mem_used_percentage_calc"].(float64)), // htop-style usage (Available)
	}
	swapRow := []string{
		"Swap",
		memInfo["swap_free"].(string),
		memInfo["swap_total"].(string),
		fmt.Sprintf("%.2f", memInfo["swap_usage"].(float64)),
		fmt.Sprintf("%.2f", memInfo["swap_used_percentage_calc"].(float64)),
	}
	RenderTable(headers, [][]string{memRow, swapRow})

	// Print top processes in pretty table format
	if top5, ok := memInfo["top_processes"].([]procInfo); ok && len(top5) > 0 {
		fmt.Println("") // Space before top processes
		fmt.Println("Top 5 Processes by Memory Usage:")
		headers := []string{"PID", "Name", "CPU %", "Memory %"}
		var data [][]string
		for _, p := range top5 {
			row := []string{
				fmt.Sprintf("%d", p.pid),
				p.name,
				fmt.Sprintf("%.1f", p.cpuPercent),
				fmt.Sprintf("%.6f", p.memoryPercent),
			}
			data = append(data, row)
		}
		RenderTable(headers, data)
	}
}

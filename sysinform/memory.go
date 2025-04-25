package sysinform

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

		// Sort by memory usage
		sort.Slice(procs, func(i, j int) bool {
			return procs[i].memoryPercent > procs[j].memoryPercent
		})

		// Get top 5 processes
		if len(procs) > 5 {
			procs = procs[:5]
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
			"top_processes":            procs,
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
	memInfo, err := getMemoryInfo()
	if err != nil {
		fmt.Println("Error getting memory info:", err)
		return
	}

	fmt.Println("===== Memory Information =====")
	if warning, ok := memInfo["warning"].(string); ok {
		fmt.Println(warning)
	}

	// Print memory table
	fmt.Println("\n┌──────┬──────────┬──────────┬─────────────┬────────────┐")
	fmt.Println("│ Type │ Free     │ Total    │ Calc Usage  │ Sys Usage  │")
	fmt.Println("├──────┼──────────┼──────────┼─────────────┼────────────┤")
	fmt.Printf("│ Mem  │ %-8s │ %-8s │ %9.2f%% │ %8.2f%% │\n",
		memInfo["mem_free"], memInfo["mem_total"],
		memInfo["mem_used_percentage_calc"], memInfo["mem_usage"])
	fmt.Println("├──────┼──────────┼──────────┼─────────────┼────────────┤")
	fmt.Printf("│ Swap │ %-8s │ %-8s │ %9.2f%% │ %8.2f%% │\n",
		memInfo["swap_free"], memInfo["swap_total"],
		memInfo["swap_used_percentage_calc"], memInfo["swap_usage"])
	fmt.Println("└──────┴──────────┴──────────┴─────────────┴────────────┘")

	// Print top processes
	if top5, ok := memInfo["top_processes"].([]procInfo); ok && len(top5) > 0 {
		fmt.Println("\nTop 5 Processes by Memory Usage:")
		fmt.Println("┌───────┬─────────────────────────────────────────┬───────────────┬──────────────────┐")
		fmt.Println("│   pid │ name                                    │   cpu_percent │   memory_percent │")
		fmt.Println("├───────┼─────────────────────────────────────────┼───────────────┼──────────────────┤")
		for _, p := range top5 {
			fmt.Printf("│ %5d │ %-37s │ %11.1f │ %12.6f │\n",
				p.pid, p.name, p.cpuPercent, p.memoryPercent)
			if p != top5[len(top5)-1] {
				fmt.Println("├───────┼─────────────────────────────────────────┼───────────────┼──────────────────┤")
			}
		}
		fmt.Println("└───────┴─────────────────────────────────────────┴───────────────┴──────────────────┘")
	}
}

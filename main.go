package main

import (
	"fmt"
	"os"

	"github.com/timmyb824/sysinformer/sysinformer"
	"github.com/urfave/cli/v2"
)

var Version = "dev"

func main() {
	app := &cli.App{
		Version: Version,
		Name:  "sysinformer",
		Usage: "Show system info",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "system", Aliases: []string{"s"}, Usage: "Show system information"},
			&cli.BoolFlag{Name: "cpu", Aliases: []string{"c"}, Usage: "Show CPU information"},
			&cli.BoolFlag{Name: "memory", Aliases: []string{"m"}, Usage: "Show memory information"},
			&cli.BoolFlag{Name: "disks", Aliases: []string{"d"}, Usage: "Show disk information"},
			&cli.BoolFlag{Name: "network", Aliases: []string{"n"}, Usage: "Show network information"},
			&cli.BoolFlag{Name: "latency", Aliases: []string{"l"}, Usage: "Show latency information"},
			&cli.BoolFlag{Name: "services", Aliases: []string{"S"}, Usage: "Show services information"},
			&cli.BoolFlag{Name: "containers", Aliases: []string{"C"}, Usage: "Show container information"},
			&cli.BoolFlag{Name: "all", Aliases: []string{"a"}, Usage: "Show all information"},
		},
		Action: func(c *cli.Context) error {
			showAll := c.Bool("all")
			showSystem := c.Bool("system")
			showCPU := c.Bool("cpu")
			showMemory := c.Bool("memory")
			showDisks := c.Bool("disks")
			showNetwork := c.Bool("network")
			showLatency := c.Bool("latency")
			showServices := c.Bool("services")
			showContainers := c.Bool("containers")

			if !showAll && !showSystem && !showCPU && !showMemory && !showDisks && !showNetwork && !showLatency && !showServices && !showContainers {
				fmt.Println("No arguments given. Use --help for help.")
				return nil
			}

			if showAll || showSystem {
				sysinformer.PrintSystemInfo()
			}

			if showAll || showCPU {
				sysinformer.PrintCPUInfo()
			}

			if showAll || showMemory {
				sysinformer.PrintMemoryInfo()
			}

			if showAll || showDisks {
				sysinformer.PrintDiskInfo()
			}

			if showAll || showNetwork {
				sysinformer.PrintNetworkInfo()
			}

			if showAll || showLatency {
				sysinformer.PrintLatencyInfo()
			}

			if showAll || showServices {
				sysinformer.PrintServicesInfo()
			}

			if showAll || showContainers {
				sysinformer.PrintContainerInfo()
			}
			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println("Error:", err)
	}
}

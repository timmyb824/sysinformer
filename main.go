package main

import (
	"fmt"
	"os"

	"github.com/timmyb824/sysinformer/sysinformer"
	"github.com/urfave/cli/v2"
)

var Version = "v1.2.0"

func main() {
	app := &cli.App{
		Version: Version,
		Name:  "sysinformer",
		Usage: "Show system info",
		Commands: []*cli.Command{
			{
				Name:      "web",
				Usage:     "Website diagnostics (ping, HTTP, DNS, SSL, WHOIS, traceroute)",
				ArgsUsage: "<url-or-domain>",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "ping", Usage: "Ping the website"},
					&cli.BoolFlag{Name: "latency", Usage: "Check website latency"},
					&cli.BoolFlag{Name: "dns", Usage: "Check DNS information"},
					&cli.BoolFlag{Name: "http", Usage: "Check HTTP status code and headers"},
					&cli.BoolFlag{Name: "ssl", Usage: "Check SSL/TLS certificate"},
					&cli.BoolFlag{Name: "whois", Usage: "Look up WHOIS information"},
					&cli.BoolFlag{Name: "trace", Usage: "Perform traceroute to the website"},
					&cli.BoolFlag{Name: "full", Usage: "Run all checks"},
					&cli.IntFlag{Name: "timeout", Value: 10, Usage: "Timeout in seconds"},
					&cli.IntFlag{Name: "count", Value: 4, Usage: "Ping count"},
				},
				Action: func(c *cli.Context) error {
					target := ""
					if c.Args().Len() > 0 {
						target = c.Args().Get(0)
					}
					if target == "" {
						return cli.Exit("missing target. Example: sysinformer web example.com --full", 1)
					}
					return sysinformer.RunWebDiagnostics(sysinformer.WebDiagOptions{
						Target:     target,
						Ping:       c.Bool("ping"),
						Latency:    c.Bool("latency"),
						DNS:        c.Bool("dns"),
						HTTP:       c.Bool("http"),
						SSL:        c.Bool("ssl"),
						Whois:      c.Bool("whois"),
						Trace:      c.Bool("trace"),
						Full:       c.Bool("full"),
						TimeoutSec: c.Int("timeout"),
						Count:      c.Int("count"),
					})
				},
			},
		},
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
				return cli.ShowAppHelp(c)
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

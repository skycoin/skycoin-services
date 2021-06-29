package main

import (
	"encoding/json"
	"fmt"
	"github.com/aeden/traceroute"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"system-survey/cmd"
)

func main() {
	jsonFlag := &cli.BoolFlag{
		Name:        "json",
		Usage:       "Use this flag to get machine readable JSON format",
		DefaultText: "Will output in human readable format",
	}

	app := &cli.App{
		Name:        "System survey tool",
		Usage:       "",
		Description: "",
		Commands: cli.Commands{
			&cli.Command{
				Name:        "apps",
				Usage:       "Check list of installed applications",
				UsageText:   "apps [command options]",
				Description: "Reads out the PATH variable and prints all installed applications.",
				Action:      actionApps,
				Flags: cli.FlagsByName{
					jsonFlag,
					&cli.StringFlag{
						Name:        "filter",
						Aliases:     []string{"f"},
						Usage:       "Comma separated application names",
						DefaultText: "All applications",
					},
				},
			},
			&cli.Command{
				Name:        "golang",
				Usage:       "Check golang version",
				UsageText:   "golang [command options]",
				Description: "Runs the `go version` command and get output as a string or error on failure.",
				Action:      actionGolang,
				Flags: cli.FlagsByName{
					jsonFlag,
				},
			},
			&cli.Command{
				Name:        "hwinfo",
				Usage:       "Get information about the host machine",
				UsageText:   "hwinfo [command options]",
				Description: "Collects system and hardware information.",
				Action:      actionHwinfo,
				Flags: cli.FlagsByName{
					jsonFlag,
				},
			},
			&cli.Command{
				Name:        "traceroute",
				Usage:       "TraceRoute utility",
				UsageText:   "traceroute[command options]",
				Description: "Tracing a route to the host.",
				Action:      actionTraceroute,
				Flags: cli.FlagsByName{
					jsonFlag,
					&cli.StringFlag{
						Name:     "dest",
						Aliases:  []string{"d"},
						Usage:    "Destination address",
						Required: true,
					},
					&cli.IntFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Usage:   "Destination port",
						Value:   traceroute.DEFAULT_PORT,
					},
					&cli.IntFlag{
						Name:  "max-hops",
						Usage: "Maximum hops",
						Value: traceroute.DEFAULT_MAX_HOPS,
					},
					&cli.IntFlag{
						Name:  "retries",
						Usage: "Maximum retries for each hop",
						Value: traceroute.DEFAULT_RETRIES,
					},
					&cli.IntFlag{
						Name:  "timeout",
						Usage: "Timeout (ms) for each hop",
						Value: traceroute.DEFAULT_TIMEOUT_MS,
					},
					&cli.IntFlag{
						Name:  "first-hop",
						Usage: "Start with hop n",
						Value: traceroute.DEFAULT_FIRST_HOP,
					},
					&cli.IntFlag{
						Name:  "packet-size",
						Usage: "Size of the sending packet",
						Value: traceroute.DEFAULT_PACKET_SIZE,
					},
				},
			},
			&cli.Command{
				Name:        "network-info",
				Usage:       "Get information about the network",
				UsageText:   "network-info [command options]",
				Description: "Collects network information.",
				Action:      actionNetworkInfo,
				Flags: cli.FlagsByName{
					jsonFlag,
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func actionApps(c *cli.Context) error {
	apps := cmd.GetAppList(c.String("filter"))
	if c.Bool("json") {
		fmt.Println(string(apps.JSON()))
	} else {
		fmt.Println(apps.String())
	}
	return nil
}

func actionGolang(c *cli.Context) error {
	gv, err := cmd.GetGolangVersion()
	if err != nil {
		return err
	}
	if c.Bool("json") {
		fmt.Println(string(gv.JSON()))
	} else {
		fmt.Println(gv.String())
	}
	return nil
}

func actionHwinfo(c *cli.Context) error {
	hi := cmd.GetHwinfo()
	if c.Bool("json") {
		fmt.Println(string(hi.JSON()))
	} else {
		fmt.Println(hi.String())
	}
	return nil
}

func actionTraceroute(c *cli.Context) error {
	result, err := cmd.Traceroute(cmd.TracerouteInput{
		MaxHops:         c.Int("max-hops"),
		Retries:         c.Int("retries"),
		Timeout:         c.Int("timeout"),
		FirstHop:        c.Int("first-hop"),
		PacketSize:      c.Int("packet-size"),
		DestinationIP:   c.String("dest"),
		DestinationPort: c.Int("port"),
	})
	if err != nil {
		return err
	}

	if c.Bool("json") {
		b, _ := json.Marshal(result)
		fmt.Println(string(b))
		return nil
	}

	fmt.Println(result)
	return nil
}

func actionNetworkInfo(c *cli.Context) error {
	_, _ = cmd.GetNetworkInfo()
	return nil
}

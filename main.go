package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/urfave/cli"
)

var (
	snapshot_len int32 = 1024
	promiscuous  bool  = false
	err          error
	timeout      time.Duration = 100 * time.Millisecond
	handle       *pcap.Handle
)

func findDevice(c *cli.Context) string {
	if c.String("interface") != "" {
		return c.String("interface")
	}
	devices, err := pcap.FindAllDevs()
	if err != nil {
		panic(err)
	}

	return devices[0].Name
}

func main() {
	app := cli.NewApp()
	app.Name = "tcpterm"
	app.Usage = "tcpdump for human"
	app.Version = "0.0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "interface, i",
			Usage: "If unspecified, use lowest numbered interface.",
		},
	}

	app.Action = func(c *cli.Context) error {
		device := findDevice(c)
		handle, err = pcap.OpenLive(device, snapshot_len, promiscuous, timeout)
		if err != nil {
			log.Fatal(err)
		}
		defer handle.Close()

		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

		tcpterm := NewTcpterm(packetSource)
		tcpterm.Run()
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
)

var network string

func init() {
	flag.StringVar(&network, "network", "ip", "Network type (ip, ip4, ip6)")
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: resolv [-network <ip|ip4|ip6>] <host>...\n")
}

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(2)
	}

	for i, host := range args {
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("Host: %s\nAddresses:\n", host)
		ips, err := net.DefaultResolver.LookupNetIP(context.Background(), network, host)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to resolve host %s: %v\n", host, err)
			continue
		}
		for _, ip := range ips {
			fmt.Printf("  %s\n", ip)
		}
	}
}

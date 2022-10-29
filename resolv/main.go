package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/netip"
	"os"
	"strconv"
)

type AddressFamily byte

const (
	AddressFamilyUnspecified AddressFamily = 0
	AddressFamilyIPv4        AddressFamily = 4
	AddressFamilyIPv6        AddressFamily = 6
)

func (af AddressFamily) String() string {
	return strconv.Itoa(int(af))
}

func (af *AddressFamily) Set(s string) error {
	switch s {
	case "4":
		*af = AddressFamilyIPv4
	case "6":
		*af = AddressFamilyIPv6
	default:
		return fmt.Errorf("invalid address family: %s", s)
	}
	return nil
}

var host = flag.String("host", "example.com", "Host to resolve")

var inet AddressFamily

func init() {
	flag.Var(&inet, "inet", "address family")
}

func main() {
	flag.Parse()

	ip, err := ResolveAddr(*host, inet)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(ip)
}

// ResolveAddr resolves a domain name string into netip.Addr.
// String representations of IP addresses are not supported.
func ResolveAddr(host string, preferredAF AddressFamily) (netip.Addr, error) {
	ips, err := net.DefaultResolver.LookupNetIP(context.Background(), "ip", host)
	if err != nil {
		return netip.Addr{}, err
	}

	fmt.Println(ips)

	if len(ips) > 1 {
		switch preferredAF {
		case AddressFamilyUnspecified:
		case AddressFamilyIPv4:
			for _, ip := range ips {
				if ip.Is4() || ip.Is4In6() {
					return ip, nil
				}
			}
		case AddressFamilyIPv6:
			for _, ip := range ips {
				if !ip.Is4() && !ip.Is4In6() {
					return ip, nil
				}
			}
		default:
			return netip.Addr{}, fmt.Errorf("invalid address family: %d", preferredAF)
		}
	}

	return ips[0], nil
}

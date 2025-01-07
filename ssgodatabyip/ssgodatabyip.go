// Package ssgodatabyip implements a simple command-line tool that reads shadowsocks-go
// log entries from stdin, and prints peer data usage by IP address to stdout.
package main

import (
	"bufio"
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"net/netip"
	"os"
	"slices"
	"strconv"
)

func main() {
	var event Event
	addrByCSID := make(map[uint64]netip.Addr)
	bytesByIP := make(map[netip.Addr]uint64)
	r := bufio.NewReader(os.Stdin)

	for {
		line, err := r.ReadSlice('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "Failed to read line: %v\n", err)
			os.Exit(1)
		}

		start := bytes.IndexByte(line, '{')
		if start < 0 {
			continue
		}
		data := line[start:]

		event = Event{}
		if err := json.Unmarshal(data, &event); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to unmarshal event %q: %v\n", data, err)
			continue
		}

		addr := event.ClientAddress.Addr()

		if event.ClientSessionID != 0 {
			if addr.IsValid() {
				addrByCSID[event.ClientSessionID] = addr
			} else {
				addr = addrByCSID[event.ClientSessionID]
			}
		}

		if addr.IsValid() {
			bytesByIP[addr] += event.NL2R + event.NR2L + event.PayloadBytesSent
		}
	}

	usage := make([]DataUsage, 0, len(bytesByIP))
	for ip, bytes := range bytesByIP {
		usage = append(usage, DataUsage{IP: ip, Bytes: bytes})
	}
	slices.SortFunc(usage, func(a, b DataUsage) int {
		return -cmp.Compare(a.Bytes, b.Bytes)
	})

	const maxLineLength = len("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff%enp5s0\t18446744073709551615\n")
	output := make([]byte, 0, len(usage)*maxLineLength)
	for _, usage := range usage {
		output = usage.IP.AppendTo(output)
		output = append(output, '\t')
		output = strconv.AppendUint(output, usage.Bytes, 10)
		output = append(output, '\n')
	}
	os.Stdout.Write(output)
}

type Event struct {
	ClientAddress    netip.AddrPort `json:"clientAddress"`
	ClientSessionID  uint64         `json:"clientSessionID"`
	NL2R             uint64         `json:"nl2r"`
	NR2L             uint64         `json:"nr2l"`
	PayloadBytesSent uint64         `json:"payloadBytesSent"`
}

type DataUsage struct {
	IP    netip.Addr `json:"ip"`
	Bytes uint64     `json:"bytes"`
}

// Package ssgodatabyip implements a simple command-line tool that reads shadowsocks-go
// log entries from stdin, and prints peer data usage by IP address to stdout.
package main

import (
	"bufio"
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
	r := newEventJSONReader(os.Stdin)
	dec := json.NewDecoder(r)

	for {
		event = Event{}
		if err := dec.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "Failed to decode JSON: %v\n", err)
			os.Exit(1)
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

// eventJSONReader streams JSON values from the inner reader,
// discarding non-JSON data at the beginning of each line.
type eventJSONReader struct {
	r        *bufio.Reader
	leftover []byte
	err      error
}

func newEventJSONReader(r io.Reader) *eventJSONReader {
	return &eventJSONReader{
		r: bufio.NewReaderSize(r, 128*1024),
	}
}

func (r *eventJSONReader) Read(p []byte) (n int, err error) {
	if len(r.leftover) > 0 {
		n = copy(p, r.leftover)
		p = p[n:]
		r.leftover = r.leftover[n:]
	}

	if r.err != nil {
		return n, r.err
	}

	for len(p) > 0 {
		if _, err = r.r.ReadSlice('{'); err != nil {
			return n, err
		}
		n++
		p[0] = '{'
		p = p[1:]

		line, err := r.r.ReadSlice('\n')
		nn := copy(p, line)
		n += nn
		p = p[nn:]
		if nn < len(line) {
			r.leftover = line[nn:]
			r.err = err
			return n, nil
		}
		if err != nil {
			return n, err
		}
	}

	return n, nil
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

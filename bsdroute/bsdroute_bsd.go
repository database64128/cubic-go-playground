//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package main

import (
	"flag"
	"log/slog"
	"net/netip"
	"os"
	"strconv"
	"sync"

	"github.com/database64128/cubic-go-playground/logging/tslog"
	"golang.org/x/net/route"
	"golang.org/x/sys/unix"
)

var (
	logNoColor bool
	logNoTime  bool
	logKVPairs bool
	logJSON    bool
	logLevel   slog.Level
)

func init() {
	flag.BoolVar(&logNoColor, "logNoColor", false, "Disable colors in log output")
	flag.BoolVar(&logNoTime, "logNoTime", false, "Disable timestamps in log output")
	flag.BoolVar(&logKVPairs, "logKVPairs", false, "Use key=value pairs in log output")
	flag.BoolVar(&logJSON, "logJSON", false, "Use JSON in log output")
	flag.TextVar(&logLevel, "logLevel", slog.LevelInfo, "Log level, one of: DEBUG, INFO, WARN, ERROR")
}

func main() {
	flag.Parse()

	logCfg := tslog.Config{
		Level:          logLevel,
		NoColor:        logNoColor,
		NoTime:         logNoTime,
		UseTextHandler: logKVPairs,
		UseJSONHandler: logJSON,
	}
	logger := logCfg.NewLogger(os.Stderr)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		monitorRoutingSocket(logger.WithAttrs(slog.String("source", "monitor")))
	}()

	rib, err := route.FetchRIB(unix.AF_UNSPEC, route.RIBTypeRoute, 0)
	if err != nil {
		logger.Error("Failed to fetch RIB", tslog.Err(err))
		os.Exit(1)
	}
	parseAndLogMsgs(logger.WithAttrs(slog.String("source", "route")), rib, true)

	rib, err = route.FetchRIB(unix.AF_UNSPEC, route.RIBTypeInterface, 0)
	if err != nil {
		logger.Error("Failed to fetch RIB", tslog.Err(err))
		os.Exit(1)
	}
	parseAndLogMsgs(logger.WithAttrs(slog.String("source", "interface")), rib, true)

	wg.Wait()
}

func monitorRoutingSocket(logger *tslog.Logger) {
	fd, err := unix.Socket(unix.AF_ROUTE, unix.SOCK_RAW, unix.AF_UNSPEC)
	if err != nil {
		logger.Error("Failed to create socket", tslog.Err(err))
		os.Exit(1)
	}
	defer unix.Close(fd)

	// rtm := route.RouteMessage{
	// 	Version: unix.RTM_VERSION,
	// 	Type:    unix.RTM_GET,
	// 	Addrs: []route.Addr{
	// 		&route.DefaultAddr{},
	// 		&route.DefaultAddr{},
	// 		&route.DefaultAddr{},
	// 		&route.DefaultAddr{},
	// 		&route.DefaultAddr{},
	// 		&route.DefaultAddr{},
	// 		&route.DefaultAddr{},
	// 		&route.DefaultAddr{},
	// 	},
	// }

	// b, err := rtm.Marshal()
	// if err != nil {
	// 	log.Fatalf("Failed to marshal route message: %v", err)
	// }

	// if _, err := unix.Write(fd, b); err != nil {
	// 	log.Fatalf("Failed to write route message: %v", err)
	// }

	b := make([]byte, 65535)

	for {
		n, err := unix.Read(fd, b)
		if err != nil {
			logger.Error("Failed to read route message", tslog.Err(err))
			continue
		}
		parseAndLogMsgs(logger, b[:n], false)
	}
}

func parseAndLogMsgs(logger *tslog.Logger, b []byte, filter bool) {
	rmsgs, err := route.ParseRIB(route.RIBTypeRoute, b)
	if err != nil {
		logger.Error("Failed to parse route message", tslog.Err(err))
		return
	}

	logger.Debug("Parsed route messages", slog.Int("count", len(rmsgs)))

	var ifindex int

	for _, m := range rmsgs {
		switch m := m.(type) {
		case *route.RouteMessage:
			if filter && (m.Flags&unix.RTF_UP == 0 ||
				m.Flags&unix.RTF_GATEWAY == 0 ||
				m.Flags&unix.RTF_HOST != 0) {
				continue
			}
			logger.Info("RouteMessage",
				logAddrs(m.Addrs, filter,
					slog.Any("type", msgType(m.Type)),
					slog.Int("ifindex", m.Index),
					tslog.Uint("id", m.ID),
				)...,
			)

		case *route.InterfaceMessage:
			if filter && (m.Flags&unix.IFF_UP == 0 ||
				m.Flags&unix.IFF_LOOPBACK != 0 ||
				m.Flags&unix.IFF_POINTOPOINT != 0 ||
				m.Flags&unix.IFF_RUNNING == 0) {
				continue
			}
			ifindex = m.Index
			logger.Info("InterfaceMessage",
				logAddrs(m.Addrs, false,
					slog.Any("type", msgType(m.Type)),
					slog.Int("ifindex", m.Index),
					slog.String("name", m.Name),
				)...,
			)

		case *route.InterfaceAddrMessage:
			if filter && m.Index != ifindex {
				continue
			}
			logger.Info("InterfaceAddrMessage",
				logAddrs(m.Addrs, false,
					slog.Any("type", msgType(m.Type)),
					slog.Int("ifindex", m.Index),
				)...,
			)

		case *route.InterfaceAnnounceMessage:
			logger.Info("InterfaceAnnounceMessage",
				slog.Any("type", msgType(m.Type)),
				slog.Int("ifindex", m.Index),
				slog.String("name", m.Name),
				slog.Int("what", m.What),
			)
		}
	}
}

type msgType int

func (m msgType) String() string {
	switch m {
	case unix.RTM_ADD:
		return "RTM_ADD"
	case unix.RTM_DELETE:
		return "RTM_DELETE"
	case unix.RTM_CHANGE:
		return "RTM_CHANGE"
	case unix.RTM_GET:
		return "RTM_GET"
	case unix.RTM_LOSING:
		return "RTM_LOSING"
	case unix.RTM_NEWADDR:
		return "RTM_NEWADDR"
	case unix.RTM_DELADDR:
		return "RTM_DELADDR"
	case unix.RTM_IFINFO:
		return "RTM_IFINFO"
	default:
		return strconv.Itoa(int(m))
	}
}

func logAddrs(addrs []route.Addr, defaultRouteOnly bool, attrs ...slog.Attr) []slog.Attr {
	dstAddr := addrFromRouteAddr(addrs[unix.RTAX_DST])
	if defaultRouteOnly && !dstAddr.IsUnspecified() {
		return attrs
	}
	gatewayAddr := addrFromRouteAddr(addrs[unix.RTAX_GATEWAY])
	netmaskAddr := addrFromRouteAddr(addrs[unix.RTAX_NETMASK])
	if defaultRouteOnly && !netmaskAddr.IsUnspecified() {
		return attrs
	}
	genmaskAddr := addrFromRouteAddr(addrs[unix.RTAX_GENMASK])
	ifpAddr := addrFromRouteAddr(addrs[unix.RTAX_IFP])
	ifaAddr := addrFromRouteAddr(addrs[unix.RTAX_IFA])
	return append(attrs,
		tslog.Addr("dst", dstAddr),
		tslog.Addr("gateway", gatewayAddr),
		tslog.Addr("netmask", netmaskAddr),
		tslog.Addr("genmask", genmaskAddr),
		tslog.Addr("ifp", ifpAddr),
		tslog.Addr("ifa", ifaAddr),
	)
}

func addrFromRouteAddr(ra route.Addr) netip.Addr {
	switch ra := ra.(type) {
	case *route.Inet4Addr:
		return netip.AddrFrom4(ra.IP)
	case *route.Inet6Addr:
		return netip.AddrFrom16(ra.IP)
	default:
		return netip.Addr{}
	}
}

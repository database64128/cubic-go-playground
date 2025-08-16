//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/netip"
	"os"
	"strconv"
	"unsafe"

	"github.com/database64128/cubic-go-playground/logging/tslog"
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

	fd, err := newRoutingSocket()
	if err != nil {
		logger.Error("Failed to open routing socket", tslog.Err(err))
		os.Exit(1)
	}
	f := os.NewFile(uintptr(fd), "route")
	defer f.Close()

	b, err := sysctlGetBytes([]int32{unix.CTL_NET, unix.AF_ROUTE, 0, unix.AF_UNSPEC, unix.NET_RT_IFLIST, 0})
	if err != nil {
		logger.Error("Failed to get interface dump", tslog.Err(err))
		os.Exit(1)
	}
	parseAndLogMsgs(logger.WithAttrs(slog.String("source", "interface")), b, true)

	b, err = sysctlGetBytes([]int32{unix.CTL_NET, unix.AF_ROUTE, 0, unix.AF_UNSPEC, unix.NET_RT_DUMP, 0})
	if err != nil {
		logger.Error("Failed to get route dump", tslog.Err(err))
		os.Exit(1)
	}
	parseAndLogMsgs(logger.WithAttrs(slog.String("source", "route")), b, true)

	monitorRoutingSocket(logger.WithAttrs(slog.String("source", "monitor")), f)
}

func monitorRoutingSocket(logger *tslog.Logger, f *os.File) {
	// route(8) monitor uses this buffer size.
	// Each read only returns a single message.
	const readBufSize = 2048
	b := make([]byte, readBufSize)
	for {
		n, err := f.Read(b)
		if err != nil {
			logger.Error("Failed to read route message", tslog.Err(err))
			continue
		}
		parseAndLogMsgs(logger, b[:n], false)
	}
}

func parseAndLogMsgs(logger *tslog.Logger, b []byte, filter bool) {
	var ifindex uint16

	for len(b) >= sizeofMsghdr {
		m := (*msghdr)(unsafe.Pointer(unsafe.SliceData(b)))
		if m.Msglen < sizeofMsghdr || int(m.Msglen) > len(b) || !m.isHdrlenOK() {
			logger.Error("Invalid message length",
				tslog.Uint("msglen", m.Msglen),
				tslog.Uint("version", m.Version),
				tslog.Uint("type", m.Type),
			)
			return
		}
		msgBuf := b[:m.Msglen]
		b = b[m.Msglen:]

		if m.Version != unix.RTM_VERSION {
			logger.Warn("Unsupported message version",
				tslog.Uint("msglen", m.Msglen),
				tslog.Uint("version", m.Version),
				tslog.Uint("type", m.Type),
			)
			continue
		}

		switch m.Type {
		case unix.RTM_ADD, unix.RTM_DELETE, unix.RTM_CHANGE, unix.RTM_GET:
			if len(msgBuf) < unix.SizeofRtMsghdr {
				logger.Error("Invalid rt_msghdr length",
					tslog.Uint("msglen", m.Msglen),
					tslog.Uint("version", m.Version),
					tslog.Uint("type", m.Type),
				)
				return
			}

			rtm := (*unix.RtMsghdr)(unsafe.Pointer(m))

			if filter && (rtm.Flags&unix.RTF_UP == 0 ||
				rtm.Flags&unix.RTF_GATEWAY == 0 ||
				rtm.Flags&unix.RTF_HOST != 0) {
				continue
			}

			var addrs [unix.RTAX_MAX]*unix.RawSockaddr
			if err := parseAddrs(&addrs, rtm.Addrs, m.addrsBuf(msgBuf, unix.SizeofRtMsghdr)); err != nil {
				logger.Error("Failed to parse addresses", tslog.Err(err))
				return
			}

			logger.Info("RouteMessage", appendAddrAttrs([]slog.Attr{
				slog.Any("type", msgType(rtm.Type)),
				tslog.Uint("ifindex", rtm.Index),
				slog.Any("flags", routeFlags(rtm.Flags)),
				tslog.Int("pid", rtm.Pid),
			}, &addrs, filter)...)

		case unix.RTM_IFINFO:
			if len(msgBuf) < unix.SizeofIfMsghdr {
				logger.Error("Invalid if_msghdr length",
					tslog.Uint("msglen", m.Msglen),
					tslog.Uint("version", m.Version),
					tslog.Uint("type", m.Type),
				)
				return
			}

			ifm := (*unix.IfMsghdr)(unsafe.Pointer(m))

			if filter && (ifm.Flags&unix.IFF_UP == 0 ||
				ifm.Flags&unix.IFF_LOOPBACK != 0 ||
				ifm.Flags&unix.IFF_POINTOPOINT != 0 ||
				ifm.Flags&unix.IFF_RUNNING == 0) {
				continue
			}

			ifindex = ifm.Index

			var addrs [unix.RTAX_MAX]*unix.RawSockaddr
			if err := parseAddrs(&addrs, ifm.Addrs, m.addrsBuf(msgBuf, unix.SizeofIfMsghdr)); err != nil {
				logger.Error("Failed to parse addresses", tslog.Err(err))
				return
			}

			logger.Info("InterfaceMessage", appendAddrAttrs([]slog.Attr{
				slog.Any("type", msgType(ifm.Type)),
				slog.Any("flags", ifaceFlags(ifm.Flags)),
				tslog.Uint("ifindex", ifm.Index),
			}, &addrs, false)...)

		case unix.RTM_NEWADDR, unix.RTM_DELADDR:
			if len(msgBuf) < unix.SizeofIfaMsghdr {
				logger.Error("Invalid ifa_msghdr length",
					tslog.Uint("msglen", m.Msglen),
					tslog.Uint("version", m.Version),
					tslog.Uint("type", m.Type),
				)
				return
			}

			ifam := (*unix.IfaMsghdr)(unsafe.Pointer(m))

			if filter && ifam.Index != ifindex {
				continue
			}

			var addrs [unix.RTAX_MAX]*unix.RawSockaddr
			if err := parseAddrs(&addrs, ifam.Addrs, m.addrsBuf(msgBuf, unix.SizeofIfaMsghdr)); err != nil {
				logger.Error("Failed to parse addresses", tslog.Err(err))
				return
			}

			logger.Info("InterfaceAddrMessage",
				appendAddrAttrs([]slog.Attr{
					slog.Any("type", msgType(ifam.Type)),
					tslog.Uint("ifindex", ifam.Index),
				}, &addrs, false)...)

		default:
			logger.Info("Unknown message type",
				tslog.Uint("msglen", m.Msglen),
				tslog.Uint("version", m.Version),
				tslog.Uint("type", m.Type),
			)
		}
	}
}

type msgType uint8

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

type routeFlags int32

func (f routeFlags) AppendText(b []byte) ([]byte, error) {
	for _, flag := range routeFlagNames {
		if f&flag.mask != 0 {
			b = append(b, flag.name)
		}
	}
	return b, nil
}

func (f routeFlags) MarshalText() ([]byte, error) {
	return f.AppendText(make([]byte, 0, len(routeFlagNames)))
}

type ifaceFlags int32

func (f ifaceFlags) AppendText(b []byte) ([]byte, error) {
	bLen := len(b)
	for _, flag := range ifaceFlagNames {
		if f&flag.mask != 0 {
			b = append(b, flag.name...)
			b = append(b, ',')
		}
	}
	if len(b) > bLen {
		b = b[:len(b)-1]
	}
	return b, nil
}

func (f ifaceFlags) MarshalText() ([]byte, error) {
	return f.AppendText(make([]byte, 0, len(ifaceFlagNames)))
}

func parseAddrs(dst *[unix.RTAX_MAX]*unix.RawSockaddr, addrs int32, b []byte) error {
	for i := range unix.RTAX_MAX {
		if addrs&(1<<i) == 0 {
			continue
		}
		if len(b) < unix.SizeofSockaddrInet4 {
			return fmt.Errorf("truncated rtaddr %d", i)
		}
		sa := (*unix.RawSockaddr)(unsafe.Pointer(unsafe.SliceData(b)))
		dst[i] = sa
		alignedLen := rtaAlign(int(sa.Len))
		if len(b) < alignedLen {
			return fmt.Errorf("truncated rtaddr %d", i)
		}
		b = b[alignedLen:]
	}
	return nil
}

func appendAddrAttrs(attrs []slog.Attr, addrs *[unix.RTAX_MAX]*unix.RawSockaddr, defaultRouteOnly bool) []slog.Attr {
	attrs = appendAddrAttr(attrs, "dst", addrs[unix.RTAX_DST], defaultRouteOnly)
	attrs = appendAddrAttr(attrs, "gateway", addrs[unix.RTAX_GATEWAY], false)
	attrs = appendAddrAttr(attrs, "netmask", addrs[unix.RTAX_NETMASK], defaultRouteOnly)
	attrs = appendAddrAttr(attrs, "genmask", addrs[unix.RTAX_GENMASK], false)
	attrs = appendAddrAttr(attrs, "ifp", addrs[unix.RTAX_IFP], false)
	attrs = appendAddrAttr(attrs, "ifa", addrs[unix.RTAX_IFA], false)
	return attrs
}

func appendAddrAttr(attrs []slog.Attr, name string, sa *unix.RawSockaddr, unspecifiedOnly bool) []slog.Attr {
	if sa == nil {
		return attrs
	}

	switch sa.Family {
	case unix.AF_INET:
		if sa.Len < unix.SizeofSockaddrInet4 {
			return attrs
		}
		sa4 := (*unix.RawSockaddrInet4)(unsafe.Pointer(sa))
		addr4 := netip.AddrFrom4(sa4.Addr)
		if unspecifiedOnly && !addr4.IsUnspecified() {
			return attrs
		}
		return append(attrs, tslog.Addr(name, addr4))

	case unix.AF_INET6:
		if sa.Len < unix.SizeofSockaddrInet6 {
			return attrs
		}
		sa6 := (*unix.RawSockaddrInet6)(unsafe.Pointer(sa))
		addr6 := netip.AddrFrom16(sa6.Addr)
		if unspecifiedOnly && !addr6.IsUnspecified() {
			return attrs
		}
		return append(attrs, tslog.Addr(name, addr6))

	case unix.AF_LINK:
		if sa.Len < unix.SizeofSockaddrDatalink {
			return attrs
		}
		sa := (*unix.RawSockaddrDatalink)(unsafe.Pointer(sa))
		if int(sa.Nlen) > len(sa.Data) {
			return attrs
		}
		ifnameBuf := unsafe.Slice((*byte)(unsafe.Pointer(&sa.Data)), sa.Nlen)
		ifname := string(ifnameBuf)
		return append(attrs, slog.String(name, ifname))

	default:
		return attrs
	}
}

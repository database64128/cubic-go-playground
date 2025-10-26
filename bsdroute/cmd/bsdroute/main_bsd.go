//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package main

import (
	"flag"
	"log/slog"
	"net/netip"
	"os"
	"unsafe"

	"github.com/database64128/cubic-go-playground/bsdroute"
	"github.com/database64128/cubic-go-playground/logging/tslog"
	"github.com/database64128/netx-go"
	"golang.org/x/sys/unix"
)

var (
	dumpAll    bool
	logNoColor bool
	logNoTime  bool
	logKVPairs bool
	logJSON    bool
	logLevel   slog.Level
)

func init() {
	flag.BoolVar(&dumpAll, "dumpAll", false, "Dump all routes and interfaces, not just the default routes and active interfaces")
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

	f, err := bsdroute.OpenRoutingSocket()
	if err != nil {
		logger.Error("Failed to open routing socket", tslog.Err(err))
		os.Exit(1)
	}
	defer f.Close()

	ioctlFd, err := bsdroute.Socket(unix.AF_INET6, unix.SOCK_DGRAM, 0)
	if err != nil {
		logger.Error("Failed to open ioctl socket", tslog.Err(err))
		os.Exit(1)
	}
	defer unix.Close(ioctlFd)

	b, err := bsdroute.SysctlGetBytes([]int32{unix.CTL_NET, unix.AF_ROUTE, 0, unix.AF_UNSPEC, unix.NET_RT_IFLIST, 0})
	if err != nil {
		logger.Error("Failed to get interface dump", tslog.Err(err))
		os.Exit(1)
	}
	parseAndLogMsgs(logger.WithAttrs(slog.String("source", "interface")), ioctlFd, b, !dumpAll)

	b, err = bsdroute.SysctlGetBytes([]int32{unix.CTL_NET, unix.AF_ROUTE, 0, unix.AF_UNSPEC, unix.NET_RT_DUMP, 0})
	if err != nil {
		logger.Error("Failed to get route dump", tslog.Err(err))
		os.Exit(1)
	}
	parseAndLogMsgs(logger.WithAttrs(slog.String("source", "route")), ioctlFd, b, !dumpAll)

	monitorRoutingSocket(logger.WithAttrs(slog.String("source", "monitor")), f, ioctlFd)
}

func monitorRoutingSocket(logger *tslog.Logger, f *os.File, ioctlFd int) {
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
		parseAndLogMsgs(logger, ioctlFd, b[:n], false)
	}
}

func parseAndLogMsgs(logger *tslog.Logger, ioctlFd int, b []byte, filter bool) {
	var ifindex uint16

	for len(b) >= bsdroute.SizeofMsghdr {
		m := (*bsdroute.Msghdr)(unsafe.Pointer(unsafe.SliceData(b)))
		if m.Msglen < bsdroute.SizeofMsghdr || int(m.Msglen) > len(b) {
			logger.Error("Invalid message length",
				tslog.Uint("msglen", m.Msglen),
				tslog.Uint("version", m.Version),
				slog.Any("type", bsdroute.MsgType(m.Type)),
			)
			return
		}
		msgBuf := b[:m.Msglen]
		b = b[m.Msglen:]

		if m.Version != unix.RTM_VERSION {
			logger.Warn("Unsupported message version",
				tslog.Uint("msglen", m.Msglen),
				tslog.Uint("version", m.Version),
				slog.Any("type", bsdroute.MsgType(m.Type)),
			)
			continue
		}

		switch m.Type {
		case unix.RTM_ADD, unix.RTM_DELETE, unix.RTM_CHANGE, unix.RTM_GET:
			if len(msgBuf) < unix.SizeofRtMsghdr {
				logger.Error("Invalid rt_msghdr length",
					tslog.Uint("msglen", m.Msglen),
					tslog.Uint("version", m.Version),
					slog.Any("type", bsdroute.MsgType(m.Type)),
				)
				return
			}

			rtm := (*unix.RtMsghdr)(unsafe.Pointer(m))

			if filter && (rtm.Flags&unix.RTF_UP == 0 ||
				rtm.Flags&unix.RTF_GATEWAY == 0 ||
				rtm.Flags&unix.RTF_HOST != 0) {
				continue
			}

			addrsBuf, ok := m.AddrsBuf(msgBuf, unix.SizeofRtMsghdr)
			if !ok {
				logger.Error("Invalid rtm_hdrlen",
					tslog.Uint("msglen", m.Msglen),
					tslog.Uint("version", m.Version),
					slog.Any("type", bsdroute.MsgType(m.Type)),
					tslog.Uint("hdrlen", m.HeaderLen()),
				)
				return
			}

			var addrs [unix.RTAX_MAX]*unix.RawSockaddr
			bsdroute.ParseAddrs(&addrs, addrsBuf, rtm.Addrs)

			logger.Info("rt_msghdr", appendAddrAttrs([]slog.Attr{
				slog.Any("type", bsdroute.MsgType(rtm.Type)),
				tslog.Uint("ifindex", rtm.Index),
				slog.Any("flags", bsdroute.RouteFlags(rtm.Flags)),
				tslog.Int("pid", rtm.Pid),
				tslog.Int("seq", rtm.Seq),
			}, &addrs, ioctlFd, rtm.Index, filter)...)

		case unix.RTM_IFINFO:
			if len(msgBuf) < unix.SizeofIfMsghdr {
				logger.Error("Invalid if_msghdr length",
					tslog.Uint("msglen", m.Msglen),
					tslog.Uint("version", m.Version),
					slog.Any("type", bsdroute.MsgType(m.Type)),
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

			addrsBuf, ok := m.AddrsBuf(msgBuf, unix.SizeofIfMsghdr)
			if !ok {
				logger.Error("Invalid ifm_hdrlen",
					tslog.Uint("msglen", m.Msglen),
					tslog.Uint("version", m.Version),
					slog.Any("type", bsdroute.MsgType(m.Type)),
					tslog.Uint("hdrlen", m.HeaderLen()),
				)
				return
			}

			var addrs [unix.RTAX_MAX]*unix.RawSockaddr
			bsdroute.ParseAddrs(&addrs, addrsBuf, ifm.Addrs)

			logger.Info("if_msghdr", appendAddrAttrs([]slog.Attr{
				slog.Any("type", bsdroute.MsgType(ifm.Type)),
				slog.Any("flags", bsdroute.IfaceFlags(ifm.Flags)),
				tslog.Uint("ifindex", ifm.Index),
			}, &addrs, ioctlFd, ifm.Index, false)...)

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

			addrsBuf, ok := m.AddrsBuf(msgBuf, unix.SizeofIfaMsghdr)
			if !ok {
				logger.Error("Invalid ifam_hdrlen",
					tslog.Uint("msglen", m.Msglen),
					tslog.Uint("version", m.Version),
					slog.Any("type", bsdroute.MsgType(m.Type)),
					tslog.Uint("hdrlen", m.HeaderLen()),
				)
				return
			}

			var addrs [unix.RTAX_MAX]*unix.RawSockaddr
			bsdroute.ParseAddrs(&addrs, addrsBuf, ifam.Addrs)

			logger.Info("ifam_msghdr",
				appendAddrAttrs([]slog.Attr{
					slog.Any("type", bsdroute.MsgType(ifam.Type)),
					tslog.Uint("ifindex", ifam.Index),
				}, &addrs, ioctlFd, ifam.Index, false)...)

		default:
			logger.Info("Unknown message type",
				tslog.Uint("msglen", m.Msglen),
				tslog.Uint("version", m.Version),
				tslog.Uint("type", m.Type),
			)
		}
	}
}

func appendAddrAttrs(attrs []slog.Attr, addrs *[unix.RTAX_MAX]*unix.RawSockaddr, ioctlFd int, ifindex uint16, defaultRouteOnly bool) []slog.Attr {
	attrs = appendAddrAttr(attrs, "dst", addrs[unix.RTAX_DST], ioctlFd, ifindex, defaultRouteOnly)
	attrs = appendAddrAttr(attrs, "gateway", addrs[unix.RTAX_GATEWAY], ioctlFd, ifindex, false)
	attrs = appendAddrAttr(attrs, "netmask", addrs[unix.RTAX_NETMASK], ioctlFd, ifindex, defaultRouteOnly)
	attrs = appendAddrAttr(attrs, "genmask", addrs[unix.RTAX_GENMASK], ioctlFd, ifindex, false)
	attrs = appendAddrAttr(attrs, "ifp", addrs[unix.RTAX_IFP], ioctlFd, ifindex, false)
	attrs = appendAddrAttr(attrs, "ifa", addrs[unix.RTAX_IFA], ioctlFd, ifindex, false)
	return attrs
}

func appendAddrAttr(attrs []slog.Attr, name string, sa *unix.RawSockaddr, ioctlFd int, ifindex uint16, unspecifiedOnly bool) []slog.Attr {
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
		addr6 := netip.AddrFrom16(sa6.Addr).WithZone(netx.ZoneCache.Name(int(sa6.Scope_id)))
		if unspecifiedOnly && !addr6.IsUnspecified() {
			return attrs
		}
		attrs = append(attrs, tslog.Addr(name, addr6))
		if name == "ifa" {
			if ifname := netx.ZoneCache.Name(int(ifindex)); ifname != "" {
				ifaFlags, err := bsdroute.IoctlGetIfaFlagInet6(ioctlFd, ifname, sa6)
				if err != nil {
					return attrs
				}
				attrs = append(attrs, slog.Any("ifaFlags", ifaFlags))
			}
		}
		return attrs

	case unix.AF_LINK:
		if sa.Len < unix.SizeofSockaddrDatalink {
			return attrs
		}
		sa := (*unix.RawSockaddrDatalink)(unsafe.Pointer(sa))
		if int(sa.Nlen) > len(sa.Data) {
			return attrs
		}
		ifname := unsafe.String((*byte)(unsafe.Pointer(&sa.Data)), sa.Nlen)
		return append(attrs, slog.GroupAttrs(name,
			tslog.Uint("index", sa.Index),
			slog.String("name", ifname),
		))

	default:
		return attrs
	}
}

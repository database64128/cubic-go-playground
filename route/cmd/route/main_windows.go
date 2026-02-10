package main

import (
	"context"
	"flag"
	"net/netip"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"unsafe"

	"github.com/database64128/cubic-go-playground/logging/tslog"
	"golang.org/x/sys/windows"
)

func main() {
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ctx.Done()
		stop()
	}()

	logCfg := tslog.Config{
		Level:          logLevel,
		NoColor:        logNoColor,
		NoTime:         logNoTime,
		UseTextHandler: logKVPairs,
		UseJSONHandler: logJSON,
	}
	logger = logCfg.NewLogger(os.Stderr)

	run(ctx)
}

type mibIpForwardRow2Notification struct {
	InterfaceLuid     uint64
	InterfaceIndex    uint32
	DestinationPrefix windows.IpAddressPrefix
	NextHop           windows.RawSockaddrInet
	NotificationType  uint32
}

var notifyRouteChange2Callback = sync.OnceValue(func() uintptr {
	return syscall.NewCallback(func(callerContext *chan<- mibIpForwardRow2Notification, row *windows.MibIpForwardRow2, notificationType uint32) uintptr {
		notifyCh := *callerContext
		var nmsg mibIpForwardRow2Notification
		if row != nil {
			nmsg.InterfaceLuid = row.InterfaceLuid
			nmsg.InterfaceIndex = row.InterfaceIndex
			nmsg.DestinationPrefix = row.DestinationPrefix
			nmsg.NextHop = row.NextHop
		}
		nmsg.NotificationType = notificationType
		notifyCh <- nmsg
		return 0
	})
})

var (
	logger                     *tslog.Logger
	initialNotificationHandled bool
)

func run(ctx context.Context) {
	// It's been observed that NotifyRouteChange2 sends 2 initial notifications and
	// blocks until the callback calls return. Be safe here and give it 2 extra slots.
	notifyCh := make(chan mibIpForwardRow2Notification, 4)

	var pinner runtime.Pinner
	pinner.Pin(&notifyCh)
	defer pinner.Unpin()

	var notificationHandle windows.Handle

	if err := windows.NotifyRouteChange2(
		windows.AF_UNSPEC,
		notifyRouteChange2Callback(),
		unsafe.Pointer(&notifyCh),
		true,
		&notificationHandle,
	); err != nil {
		logger.Error("Failed to register for route change notifications",
			tslog.Err(os.NewSyscallError("NotifyRouteChange2", err)),
		)
		return
	}

	logger.Info("Registered for route change notifications",
		tslog.Uint("notificationHandle", notificationHandle),
	)

	go func() {
		// Always close notifyCh, even if we failed to unregister,
		// because all bets are off anyways.
		defer close(notifyCh)

		<-ctx.Done()

		// Apparently, even on success, the notification handle can be NULL!
		// I mean, WTF, Microsoft?!
		if notificationHandle == 0 {
			logger.Debug("Skipping CancelMibChangeNotify2 because notification handle is NULL")
			return
		}

		if err := windows.CancelMibChangeNotify2(notificationHandle); err != nil {
			logger.Error("Failed to unregister for route change notifications",
				tslog.Uint("notificationHandle", notificationHandle),
				tslog.Err(os.NewSyscallError("CancelMibChangeNotify2", err)),
			)
			return
		}

		logger.Info("Unregistered for route change notifications")
	}()

	for nmsg := range notifyCh {
		handleMibIpForwardRow2Notification(nmsg)
	}
}

func handleMibIpForwardRow2Notification(nmsg mibIpForwardRow2Notification) {
	destinationPrefix := prefixFromIpAddressPrefix(&nmsg.DestinationPrefix)
	nextHop := ipFromSockaddrInet(&nmsg.NextHop)
	logger.Info("Received route change notification",
		tslog.Uint("InterfaceLuid", nmsg.InterfaceLuid),
		tslog.Uint("InterfaceIndex", nmsg.InterfaceIndex),
		tslog.Prefix("DestinationPrefix", destinationPrefix),
		tslog.Addr("NextHop", nextHop),
		tslog.Uint("NotificationType", nmsg.NotificationType),
	)

	switch nmsg.NotificationType {
	case windows.MibParameterNotification, windows.MibAddInstance, windows.MibDeleteInstance:
		// Retrieve full route information.
		row := windows.MibIpForwardRow2{
			InterfaceLuid:     nmsg.InterfaceLuid,
			InterfaceIndex:    nmsg.InterfaceIndex,
			DestinationPrefix: nmsg.DestinationPrefix,
			NextHop:           nmsg.NextHop,
		}

		if err := windows.GetIpForwardEntry2(&row); err != nil {
			if err == windows.ERROR_NOT_FOUND {
				logger.Debug("Skipping route change notification for route that no longer exists")
				return
			}
			logger.Error("Failed to get route information for route change notification",
				tslog.Err(os.NewSyscallError("GetIpForwardEntry2", err)),
			)
			return
		}

		logMibIpForwardRow2(&row, false)

	case windows.MibInitialNotification:
		// Skip subsequent initial notifications.
		if initialNotificationHandled {
			logger.Debug("Skipping subsequent initial notification")
			return
		}
		initialNotificationHandled = true

		var table *windows.MibIpForwardTable2
		if err := windows.GetIpForwardTable2(windows.AF_UNSPEC, &table); err != nil {
			logger.Error("Failed to get route table",
				tslog.Err(os.NewSyscallError("GetIpForwardTable2", err)),
			)
			return
		}

		rows := table.Rows()
		for i := range rows {
			logMibIpForwardRow2(&rows[i], !dumpAll)
		}

		windows.FreeMibTable(unsafe.Pointer(table))

	default:
		logger.Warn("Unknown notification type")
	}
}

func logMibIpForwardRow2(row *windows.MibIpForwardRow2, defaultOnly bool) {
	destinationPrefix := prefixFromIpAddressPrefix(&row.DestinationPrefix)
	if defaultOnly && (destinationPrefix.Bits() != 0 || !destinationPrefix.Addr().IsUnspecified()) {
		return
	}
	logger.Info("MibIpForwardRow2",
		tslog.Uint("InterfaceLuid", row.InterfaceLuid),
		tslog.Uint("InterfaceIndex", row.InterfaceIndex),
		tslog.Prefix("DestinationPrefix", destinationPrefix),
		tslog.Addr("NextHop", ipFromSockaddrInet(&row.NextHop)),
		tslog.Uint("SitePrefixLength", row.SitePrefixLength),
		tslog.Uint("ValidLifetime", row.ValidLifetime),
		tslog.Uint("PreferredLifetime", row.PreferredLifetime),
		tslog.Uint("Metric", row.Metric),
		tslog.Uint("Protocol", row.Protocol),
		tslog.Uint("Loopback", row.Loopback),
		tslog.Uint("AutoconfigureAddress", row.AutoconfigureAddress),
		tslog.Uint("Publish", row.Publish),
		tslog.Uint("Immortal", row.Immortal),
		tslog.Uint("Age", row.Age),
		tslog.Uint("Origin", row.Origin),
	)
}

func prefixFromIpAddressPrefix(p *windows.IpAddressPrefix) netip.Prefix {
	return netip.PrefixFrom(ipFromSockaddrInet(&p.Prefix), int(p.PrefixLength))
}

func ipFromSockaddrInet(sa *windows.RawSockaddrInet) netip.Addr {
	switch sa.Family {
	case windows.AF_INET:
		sa4 := (*windows.RawSockaddrInet4)(unsafe.Pointer(sa))
		return netip.AddrFrom4(sa4.Addr)
	case windows.AF_INET6:
		sa6 := (*windows.RawSockaddrInet6)(unsafe.Pointer(sa))
		return netip.AddrFrom16(sa6.Addr)
	default:
		return netip.Addr{}
	}
}

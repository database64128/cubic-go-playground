package main

import "golang.org/x/sys/unix"

/*
struct bits {
	int	b_mask;
	char	b_val;
};
static const struct bits bits[] = {
	{ RTF_UP,	'U' },
	{ RTF_GATEWAY,	'G' },
	{ RTF_HOST,	'H' },
	{ RTF_REJECT,	'R' },
	{ RTF_BLACKHOLE, 'B' },
	{ RTF_DYNAMIC,	'D' },
	{ RTF_MODIFIED,	'M' },
	{ RTF_DONE,	'd' },
	{ RTF_MASK,	'm' },
	{ RTF_CONNECTED, 'C' },
	{ RTF_LLDATA,	'L' },
	{ RTF_STATIC,	'S' },
	{ RTF_PROTO1,	'1' },
	{ RTF_PROTO2,	'2' },
	{ RTF_ANNOUNCE,	'p' },
	{ RTF_LOCAL, 'l'},
	{ RTF_BROADCAST, 'b'},
	{ 0, 0 }
};
*/

var routeFlagNames = [...]struct {
	mask routeFlags
	name byte
}{
	{unix.RTF_UP, 'U'},
	{unix.RTF_GATEWAY, 'G'},
	{unix.RTF_HOST, 'H'},
	{unix.RTF_REJECT, 'R'},
	{unix.RTF_BLACKHOLE, 'B'},
	{unix.RTF_DYNAMIC, 'D'},
	{unix.RTF_MODIFIED, 'M'},
	{unix.RTF_DONE, 'd'},
	{unix.RTF_MASK, 'm'},
	// {unix.RTF_CONNECTED, 'C'},
	// {unix.RTF_LLDATA, 'L'},
	{unix.RTF_STATIC, 'S'},
	{unix.RTF_PROTO1, '1'},
	{unix.RTF_PROTO2, '2'},
	{unix.RTF_ANNOUNCE, 'p'},
	// {unix.RTF_LOCAL, 'l'},
	// {unix.RTF_BROADCAST, 'b'},
}

/*
	IFF_UP                            = 0x1
	IFF_BROADCAST                     = 0x2
	IFF_DEBUG                         = 0x4
	IFF_LOOPBACK                      = 0x8
	IFF_POINTOPOINT                   = 0x10
	IFF_NOTRAILERS                    = 0x20
	IFF_RUNNING                       = 0x40
	IFF_NOARP                         = 0x80
	IFF_PROMISC                       = 0x100
	IFF_ALLMULTI                      = 0x200
	IFF_OACTIVE                       = 0x400
	IFF_SIMPLEX                       = 0x800
	IFF_LINK0                         = 0x1000
	IFF_LINK1                         = 0x2000
	IFF_LINK2                         = 0x4000
	IFF_MULTICAST                     = 0x8000
	IFF_CANTCHANGE                    = 0x8f52
*/

var ifaceFlagNames = [...]struct {
	mask ifaceFlags
	name string
}{
	{unix.IFF_UP, "UP"},
	{unix.IFF_BROADCAST, "BROADCAST"},
	{unix.IFF_DEBUG, "DEBUG"},
	{unix.IFF_LOOPBACK, "LOOPBACK"},
	{unix.IFF_POINTOPOINT, "POINTOPOINT"},
	{unix.IFF_NOTRAILERS, "NOTRAILERS"},
	{unix.IFF_RUNNING, "RUNNING"},
	{unix.IFF_NOARP, "NOARP"},
	{unix.IFF_PROMISC, "PROMISC"},
	{unix.IFF_ALLMULTI, "ALLMULTI"},
	{unix.IFF_OACTIVE, "OACTIVE"},
	{unix.IFF_SIMPLEX, "SIMPLEX"},
	{unix.IFF_LINK0, "LINK0"},
	{unix.IFF_LINK1, "LINK1"},
	{unix.IFF_LINK2, "LINK2"},
	{unix.IFF_MULTICAST, "MULTICAST"},
}

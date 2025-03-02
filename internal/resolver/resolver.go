package resolver

import (
	"context"
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/al002/zbittorrent/internal/blocklist"
)

var (
	ErrBlocked        = errors.New("ip is blocked")
	ErrNotIpv4Address = errors.New("not ipv4 address")
	ErrInvalidPort    = errors.New("invalid port number")
)

func Resolve(ctx context.Context, hostport string, timeout time.Duration, bl *blocklist.Blocklist) (net.IP, int, error) {
	host, portStr, err := net.SplitHostPort(hostport)
	if err != nil {
		return nil, 0, err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, 0, err
	}

	if port <= 0 || port > 65535 {
		return nil, 0, ErrInvalidPort
	}

	ip := net.ParseIP(host)
	if ip == nil {
		ip, err = ResolveIPv4(ctx, timeout, host)
		if err != nil {
			return nil, 0, err
		}
	}

	ip4 := ip.To4()
	if ip4 == nil {
		return nil, 0, ErrNotIpv4Address
	}
	if bl != nil && bl.Blocked(ip) {
		return nil, 0, ErrBlocked
	}

	return ip4, port, nil
}

func ResolveIPv4(ctx context.Context, timeout time.Duration, host string) (net.IP, error) {
	var cancel func()
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}

	for _, ia := range addrs {
		ip4 := ia.IP.To4()
		if ip4 != nil {
			return ip4, nil
		}
	}

	return nil, ErrNotIpv4Address
}

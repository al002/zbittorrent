package udpTracker

import (
	"bytes"
	"context"
	"encoding"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/al002/zbittorrent/internal/logger"
	trackerTypes "github.com/al002/zbittorrent/internal/tracker/types"
	"github.com/anacrolix/dht/v2/krpc"
	"github.com/protolambda/ctxlock"
)

type AnnounceResponsePeers interface {
	encoding.BinaryUnmarshaler
	NodeAddrs() []krpc.NodeAddr
}

type ClientOpts struct {
	Network                 string // udp, udp4, udp6
	Addr                    string // Tracker address
	Ipv6                    *bool
	ListenePacket           func(network, addr string) (net.PacketConn, error)
	ShouldReconnectOverride func() bool
	Logger                  logger.Logger
}

type Client struct {
	connId                  ConnectionId
	connIdIssued            time.Time
	conn                    net.PacketConn
	network                 string
	addr                    string
	ipv6                    *bool
	tm                      *TransactionManager
	shouldReconnectOverride func() bool
	closed                  bool
	readErr                 error
	logger                  logger.Logger
	mu                      ctxlock.Lock
}

func NewClient(opts ClientOpts) (*Client, error) {
	var conn net.PacketConn
	var err error

	if opts.ListenePacket != nil {
		conn, err = opts.ListenePacket(opts.Network, ":0")
	} else {
		conn, err = net.ListenPacket(opts.Network, ":0")
	}

	if err != nil {
		return nil, err
	}

	client := &Client{
		conn:                    conn,
		network:                 opts.Network,
		addr:                    opts.Addr,
		ipv6:                    opts.Ipv6,
		shouldReconnectOverride: opts.ShouldReconnectOverride,
		logger:                  opts.Logger,
	}

	go client.reader()

	return client, nil
}

// read data from network, dispatch data to transaction manager
func (c *Client) reader() {
	b := make([]byte, 0x800)

	for {
		n, addr, err := c.conn.ReadFrom(b)
		if err != nil {
			c.readErr = err
			if !c.closed {
				c.Close()
			}
			break
		}

		err = c.tm.Dispatch(b[:n], addr)
		if err != nil {
      fmt.Printf("dispatching packet received on %v: %v", c.conn.LocalAddr(), err)
			s := fmt.Sprintf("dispatching packet received on %v: %v", c.conn.LocalAddr(), err)
			c.logger.Info(s)
		}
	}
}

func (c *Client) Close() error {
	c.closed = true
	return c.conn.Close()
}

func (c *Client) Announce(ctx context.Context, req trackerTypes.AnnounceRequest) (
	respHeader AnnounceResponseHeader,
	peers AnnounceResponsePeers,
	err error,
) {
	respBody, addr, err := c.request(ctx, ActionAnnounce, mustMarshal(req))

	if err != nil {
		return
	}

	r := bytes.NewBuffer(respBody)
	err = Read(r, &respHeader)

	if err != nil {
		err = fmt.Errorf("reading response header: %w", err)
		return
	}

	if c.isIPv6(addr) {
		peers = &krpc.CompactIPv6NodeAddrs{}
	} else {
		peers = &krpc.CompactIPv4NodeAddrs{}
	}

	err = peers.UnmarshalBinary(r.Bytes())

	if err != nil {
		err = fmt.Errorf("reading response peers: %w", err)
	}

	return
}

const ConnectionIdMismatchNul = "Connection ID mismatch.\x00"

type ErrorResponse struct {
	Message string
}

func (me ErrorResponse) Error() string {
	return fmt.Sprintf("error response: %#q", me.Message)
}

func (c *Client) request(ctx context.Context, action Action, body []byte) (respBody []byte, addr net.Addr, err error) {
	respChan := make(chan DispatchedResponse, 1)

	tx := c.tm.NewTransaction(func(dr DispatchedResponse) {
		respChan <- dr
	})

	defer tx.End()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	writeErr := make(chan error, 1)

	go func() {
		writeErr <- c.requestUntilResponse(ctx, action, body, tx.Id())
	}()

	select {
	case dr := <-respChan:
		if dr.Header.Action == action {
			respBody = dr.Body
			addr = dr.Addr
		} else if dr.Header.Action == ActionError {
			stringBody := string(dr.Body[:])
			err = ErrorResponse{Message: stringBody}
			if stringBody == ConnectionIdMismatchNul {
				err = fmt.Errorf("Connection Id mismatched %w", err)
			}
			c.connIdIssued = time.Time{}
		} else {
			err = fmt.Errorf("unexpected response action %v", dr.Header.Action)
		}
	case err = <-writeErr:
		err = fmt.Errorf("write error: %w", err)
	case <-ctx.Done():
		err = context.Cause(ctx)
	}

	return
}

func (c *Client) requestUntilResponse(
	ctx context.Context,
	action Action,
	body []byte,
	tId TransactionId,
) (err error) {
	var buf bytes.Buffer

	for n := 0; ; n++ {
		err = c.doRequest(ctx, action, body, tId, &buf)
		if err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(timeout(n)):
		}
	}
}

func (c *Client) doRequest(
	ctx context.Context,
	action Action,
	body []byte,
	tId TransactionId,
	buf *bytes.Buffer,
) (err error) {
	var connId ConnectionId
	// if it is ActionConnect, use connection id magic number
	if action == ActionConnect {
		connId = ConnectRequestConnectionId
	} else {
		// Lock ctx while establishing a connection id, ensuring that the request
		// is written before connection id to change again.
		err = c.mu.LockCtx(ctx)
		if err != nil {
			return fmt.Errorf("locking connection id: %w", err)
		}
		defer c.mu.Unlock()

		// Get new connection id
		connId, err = c.connIdForRequest(ctx, action)
    fmt.Printf("connId %v", connId)
		if err != nil {
			return
		}
	}

	// Prepare request data
	buf.Reset()
	err = Write(buf, RequestHeader{
		ConnectionId:  connId,
		Action:        action,
		TransactionId: tId,
	})

	if err != nil {
		panic(err)
	}

	buf.Write(body)

	addr, err := net.ResolveUDPAddr(c.network, c.addr)
	if err != nil {
		return err
	}

	// Write data to udp socket is sent the data
	_, err = c.conn.WriteTo(buf.Bytes(), addr)
	return
}

func (c *Client) connIdForRequest(ctx context.Context, action Action) (id ConnectionId, err error) {
	if action == ActionConnect {
		id = ConnectRequestConnectionId
		return
	}

	err = c.connect(ctx)
	if err != nil {
		return
	}

	id = c.connId
	return
}

func (c *Client) connect(ctx context.Context) (err error) {
	if !c.shouldReconnect() {
		return nil
	}

	return c.doConnectRoundTrip(ctx)
}

func (c *Client) doConnectRoundTrip(ctx context.Context) (err error) {
	respBody, _, err := c.request(ctx, ActionConnect, nil)
	if err != nil {
		return err
	}

	var connResp ConnectionResponsne

	err = binary.Read(bytes.NewReader(respBody), binary.BigEndian, &connResp)
	if err != nil {
		return
	}

  fmt.Printf("connId %v", connResp.ConnectionId)
	c.connId = connResp.ConnectionId
	c.connIdIssued = time.Now()
	return
}

func (c *Client) isIPv6(addr net.Addr) bool {
	if c.ipv6 != nil {
		return *c.ipv6
	}

	switch c.network {
	case "udp4":
		return false
	case "udp6":
		return true
	}
	ip := AddrIP(addr)
	return ip.To16() != nil && ip.To4() == nil
}

func (c *Client) shouldReconnect() bool {
	if c.shouldReconnectOverride != nil {
		return c.shouldReconnectOverride()
	}

	return c.connIdIssued.IsZero() || time.Since(c.connIdIssued) >= time.Minute
}

func AddrIP(addr net.Addr) net.IP {
	if addr == nil {
		return nil
	}

	switch raw := addr.(type) {
	case *net.UDPAddr:
		return raw.IP
	case *net.TCPAddr:
		return raw.IP
	default:
		host, _, err := net.SplitHostPort(addr.String())
		if err != nil {
			panic(err)
		}

		return net.ParseIP(host)
	}
}

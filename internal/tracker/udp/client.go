package udp

import (
	"io"
	"time"
)

type Client struct {
	connId                  ConnectionId
	connIdIssued            time.Time
	shouldReconnectOverride func() bool
	// Dispatcher              *Dispatcher
	Writer                  io.Writer
}



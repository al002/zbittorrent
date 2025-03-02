package udptracker

import (
	"net"
	"time"
)

type connection struct {
  *requestBase
  *connectRequest
  requests []*transportRequest
  addr *net.UDPAddr
  id int64
  connectedAt time.Time
}

var _ udpRequest = (*connection)(nil)

func newConnection(req *transportRequest) *connection {
  return &connection{
    requestBase: newRequestBase(req.ctx, req.dest),
    connectRequest: newConnectRequest(),
    requests: []*transportRequest{req},
  }
}

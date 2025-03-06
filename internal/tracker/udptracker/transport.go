package udptracker

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/al002/zbittorrent/internal/blocklist"
	"github.com/al002/zbittorrent/internal/log"
	"github.com/al002/zbittorrent/internal/resolver"
	"github.com/al002/zbittorrent/internal/tracker"
	"github.com/al002/zbittorrent/pkg/bencode"
	"github.com/cenkalti/backoff/v5"
)

const (
	connectionIDMagic    = 0x41727101980
	connectionIDInterval = time.Minute
)

type Transport struct {
	dnsTimeout time.Duration
	blocklist  *blocklist.Blocklist
	requestC   chan *transportRequest
	readC      chan []byte
	closeC     chan struct{}
	doneC      chan struct{}
	log        log.Logger
}

func NewTransport(bl *blocklist.Blocklist, dnsTimeout time.Duration, logger log.Logger) *Transport {
	return &Transport{
		blocklist:  bl,
		dnsTimeout: dnsTimeout,
		log:        logger,
		requestC:   make(chan *transportRequest),
		readC:      make(chan []byte),
		closeC:     make(chan struct{}),
		doneC:      make(chan struct{}),
	}
}

func (t *Transport) Close() {
	close(t.closeC)
	<-t.doneC
}

func (t *Transport) Do(req *transportRequest) ([]byte, error) {
	var errTransportClosed = errors.New("udp transport closed")

	select {
	case t.requestC <- req:
	case <-req.ctx.Done():
		return nil, req.ctx.Err()
	case <-t.closeC:
		return nil, errTransportClosed
	}

	select {
	// when conn.SetResponse
	case <-req.done:
		return req.response, req.err
	case <-req.ctx.Done():
		return nil, req.ctx.Err()
	case <-t.closeC:
		return nil, errTransportClosed
	}
}

func (t *Transport) Run() {
	t.log.Debug("Starting udp transport run loop")
	var listening bool
	var uaddr net.UDPAddr

	udpConn, listenErr := net.ListenUDP("udp4", &uaddr)

	if listenErr != nil {
		t.log.Error(listenErr.Error())
	} else {
		listening = true
		t.log.Debug("Starting udp transport read loop")
		go t.readLoop(udpConn)
	}

	transactions := make(map[int32]*transaction)
	connections := make(map[string]*connection)
	connectDone := make(chan *connectionResult)
	connectionExpired := make(chan string)

	beginTransaction := func(i udpRequest) (*transaction, error) {
		trx := newTransaction(i)
		_, ok := transactions[trx.id]
		if ok {
			return nil, errors.New("transaction id collision")
		}
		transactions[trx.id] = trx
		return trx, nil
	}

	for {
		select {
		case req := <-t.requestC:
			// if cannot listen on the UDP port, break early
			if !listening {
				req.SetResponse(nil, listenErr)
				break
			}
			conn, ok := connections[req.dest]
			if !ok {
				// No connection for dest, make new one
				conn = newConnection(req)
				connections[req.dest] = conn
				// Make new transaction for this connection
				trx, err := beginTransaction(conn)
				if err != nil {
					conn.SetResponse(nil, err)
				} else {
					// sent `connect` action
					go resolveDestinationAndConnect(trx, req.dest, udpConn, t.dnsTimeout, t.blocklist, connectDone, t.closeC)
				}
			} else {
				if !conn.connectedAt.IsZero() {
					// connection is connected
					req.ConnectionID = conn.id
					// make new transaction ID
					trx, err := beginTransaction(req)
					if err != nil {
						req.SetResponse(nil, err)
					} else {
						// retry connect action
						go retryTransaction(trx, udpConn, conn.addr)
					}
				} else {
					// connection is in connecting state
					conn.requests = append(conn.requests, req)
				}
			}
		case res := <-connectDone:
			// connect finished, delete transaction ID
			conn := res.trx.request.(*connection)
			delete(transactions, res.trx.id)

			// Handle connection error
			if res.err != nil {
				delete(connections, res.dest)
				for _, req := range conn.requests {
					req.SetResponse(nil, res.err)
				}
				break
			}

			// connected
			conn.addr = res.addr
			conn.id = res.id
			conn.connectedAt = res.connectedAt

			// Expire the connection after efined period
			go func(dest string) {
				select {
				case <-time.After(connectionIDInterval):
				case <-t.closeC:
					return
				}

				select {
				case connectionExpired <- dest:
				case <-t.closeC:
				}
			}(res.dest)

			// Start announce transaction
			for _, req := range conn.requests {
				req.ConnectionID = conn.id
				trx, err := beginTransaction(req)
				if err != nil {
					req.SetResponse(nil, err)
				} else {
					go retryTransaction(trx, udpConn, conn.addr)
				}
			}

			// Clear waiting requests
			conn.requests = nil
		case dest := <-connectionExpired:
			// Connection expired
			delete(connections, dest)
		case buf := <-t.readC:
			// Read udp message, data from readLoop, include connect or other action
			var header udpMessageHeader
			err := binary.Read(bytes.NewReader(buf), binary.BigEndian, &header)
			if err != nil {
				t.log.Error(err.Error())
				continue
			}

			trx, ok := transactions[header.TransactionID]
			if !ok {
				t.log.Debug(
					"Unexpected transaction ID",
					"transaction_id", header.TransactionID,
				)
				continue
			}

			t.log.Debug(
				"Receive response for transaction ID",
				"transaction_id", header.TransactionID,
			)

			if header.Action == actionError {
				// error is after the header
				rest := buf[binary.Size(header):]
				var terr struct {
					FailureReason string `bencode:"failure reason"`
					RetryIn       string `bencode:"retry in"`
				}

				err = bencode.Unmarshal(rest, &terr)
				if err != nil {
					err = tracker.ErrDecode
				} else {
					retryIn, _ := strconv.Atoi(terr.RetryIn)
					err = &tracker.Error{
						FailureReason: terr.FailureReason,
						RetryIn:       time.Duration(retryIn) * time.Minute,
					}
				}
			}

			// It is will always set connect action response first
			// possibly trigger sendAndReceiveConnect trx.request.(*connection).done
			trx.request.SetResponse(buf, err)
			trx.cancel()
		case <-t.closeC:
			// Transport close
			for _, conn := range connections {
				for _, req := range conn.requests {
					req.SetResponse(nil, errors.New("transport closing"))
				}
			}

			for _, trx := range transactions {
				trx.cancel()
			}

			if listening {
				udpConn.Close()
			}
			close(t.doneC)
			return
		}
	}
}

func (t *Transport) readLoop(conn net.Conn) {
	const maxNumWant = 1000
	bigBuf := make([]byte, 20+6*maxNumWant)

	for {
		n, err := conn.Read(bigBuf)
		if err != nil {
			select {
			case <-t.closeC:
			default:
				t.log.Error(err.Error())
			}

			return
		}
		buf := make([]byte, n)
		copy(buf, bigBuf)

		select {
		case t.readC <- buf:
		case <-t.closeC:
			return
		}
	}
}

type connectionResult struct {
	id          int64
	trx         *transaction
	dest        string
	addr        *net.UDPAddr
	err         error
	connectedAt time.Time
}

func resolveDestinationAndConnect(trx *transaction, dest string, udpConn *net.UDPConn, dnsTimeout time.Duration, blocklist *blocklist.Blocklist, connectDoneC chan *connectionResult, stopC chan struct{}) {
	res := &connectionResult{
		trx:  trx,
		dest: dest,
	}

	ip, port, err := resolver.Resolve(trx.ctx, dest, dnsTimeout, blocklist)
	if err != nil {
		res.err = err
		select {
		// Trigger connectDone, but has err
		case connectDoneC <- res:
		case <-stopC:
		}
		return
	}

	res.addr = &net.UDPAddr{
		IP:   ip,
		Port: port,
	}

	res.id, res.err = sendAndReceiveConnect(trx, udpConn, res.addr)
	if res.err == nil {
		res.connectedAt = time.Now()
	}

	select {
	// Trigger connectDone, succesful status
	case connectDoneC <- res:
	case <-stopC:
	}
}

func sendAndReceiveConnect(trx *transaction, conn *net.UDPConn, addr net.Addr) (connectionID int64, err error) {
	go retryTransaction(trx, conn, addr)

	select {
	// Request is done
	case <-trx.request.(*connection).done:
	case <-trx.ctx.Done():
		return 0, trx.ctx.Err()
	}

	data, err := trx.request.GetResponse()
	if err != nil {
		return 0, err
	}

	var response connectResponse
	err = binary.Read(bytes.NewReader(data), binary.BigEndian, &response)
	if err != nil {
		return 0, err
	}

	if response.Action != actionConnect {
		return 0, errors.New("Invalid action in connect response")
	}

	return response.ConnectionID, nil
}

func retryTransaction(trx *transaction, conn *net.UDPConn, addr net.Addr) {
	// sent transaction with backoff retry
	var b bytes.Buffer
	_, _ = trx.request.WriteTo(&b)
	data := b.Bytes()

	ticker := backoff.NewTicker(new(udpBackOff))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Sent request
			_, _ = conn.WriteTo(data, addr)
		case <-trx.ctx.Done():
			return
		}
	}
}

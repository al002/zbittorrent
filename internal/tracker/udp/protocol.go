package udpTracker

import (
	"bytes"
	"encoding/binary"
	"io"
)

// UDP tracker protocol action type
type Action int32

const (
	ActionConnect  Action = iota // Connect
	ActionAnnounce               // Announce
	ActionScrape                 // Scrape
	ActionError                  // Error
)

const ConnectRequestConnectionId uint64 = 0x41727101980

// Messages

// Match response with request
type TransactionId int32

// Assigned by server, used for subsequent request
type ConnectionId = uint64

// Common header for all request
type RequestHeader struct {
	ConnectionId  ConnectionId
	Action        Action
	TransactionId TransactionId
} // 16 bytes

// Common header for all response
type ResponseHeader struct {
	Action        Action
	TransactionId TransactionId
}

type ConnectionResponsne struct {
	ConnectionId ConnectionId
}

type Event int32

const (
	EventNone Event = iota
	EventCompleted
	EventStarted
	EventStopped
)

type AnnounceResponseHeader struct {
	Interval int32
	Leechers int32
	Seeders  int32
}

func marshal(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := Write(&buf, data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func mustMarshal(data interface{}) []byte {
  b, err := marshal(data)
  if err != nil {
    panic(err)
  }

  return b
}

func Write(w io.Writer, data interface{}) error {
	return binary.Write(w, binary.BigEndian, data)
}

func Read(r io.Reader, data interface{}) error {
	return binary.Read(r, binary.BigEndian, data)
}

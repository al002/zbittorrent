package udpTracker

import (
	"bytes"
	"fmt"
	"net"
	"sync"
)

type DispatchedResponse struct {
	Header ResponseHeader
	Body   []byte
	Addr   net.Addr
}

type TransactionManager struct {
	mu           sync.RWMutex
	transactions map[TransactionId]Transaction
}

func (tm *TransactionManager) Dispatch(b []byte, addr net.Addr) error {
	buf := bytes.NewBuffer(b)
	var rh ResponseHeader

	err := Read(buf, &rh)

  if err != nil {
    return err
  }

	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if t, ok := tm.transactions[rh.TransactionId]; ok {
  fmt.Printf("transaction id %v", rh.TransactionId)
		t.h(DispatchedResponse{
			Header: rh,
			Body:   append([]byte(nil), buf.Bytes()...),
			Addr:   addr,
		})
		return nil
	} else {
  fmt.Printf("error transaction id %v\n", rh.TransactionId)
		return fmt.Errorf("unknown transaction id %v", rh.TransactionId)
	}
}

func (tm *TransactionManager) forgetTransaction(id TransactionId) {
  tm.mu.Lock()
  defer tm.mu.Unlock()
  delete(tm.transactions, id)
}

func (tm *TransactionManager) NewTransaction(h TransactionResponseHandler) Transaction {
  tm.mu.Lock()
  defer tm.mu.Unlock()

  for {
    id := RandomTransactionId()
    if _, ok := tm.transactions[id]; ok {
      continue
    }

    t := Transaction{
      tm: tm,
      h: h,
      id: id,
    }

    if tm.transactions == nil {
      tm.transactions = make(map[TransactionId]Transaction)
    }

    tm.transactions[id] = t
    return t
  }
}

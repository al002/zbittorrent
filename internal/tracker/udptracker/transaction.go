package udptracker

import (
	"context"
	"io"
	"math/rand"
)

type transaction struct {
	id      int32
	request udpRequest
	ctx     context.Context
	cancel  func()
}

type udpRequest interface {
	io.WriterTo
	SetTransactionID(int32)
	GetContext() context.Context
	GetResponse() (data []byte, err error)
	SetResponse(data []byte, err error)
}

func newTransaction(req udpRequest) *transaction {
	t := &transaction{
		id:      rand.Int31(),
		request: req,
	}
	req.SetTransactionID(t.id)
	t.ctx, t.cancel = context.WithCancel(req.GetContext())
	return t
}

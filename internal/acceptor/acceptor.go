package acceptor

import (
	"net"

	"github.com/al002/zbittorrent/internal/log"
)

type Acceptor struct {
	listener  net.Listener
	newConnsC chan net.Conn
	closeC    chan struct{}
	doneC     chan struct{}
	log       log.Logger
}

func New(lis net.Listener, newConnsC chan net.Conn, l log.Logger) *Acceptor {
	return &Acceptor{
		listener:  lis,
		newConnsC: newConnsC,
		closeC:    make(chan struct{}),
		doneC:     make(chan struct{}),
		log:       l,
	}
}

func (a *Acceptor) Close() {
	close(a.closeC)
	<-a.doneC
}

func (a *Acceptor) Run() {
	defer close(a.doneC)

	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-a.closeC:
			a.listener.Close()
		case <-done:
		}
	}()

	for {
		conn, err := a.listener.Accept()
		if err != nil {
			select {
			case <-a.closeC:
			default:
				a.log.Error(err.Error())
			}

			return
		}

		select {
		case a.newConnsC <- conn:
		case <-a.closeC:
			conn.Close()
			return
		}
	}
}

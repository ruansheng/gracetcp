package gracetcp

import (
	"net"
	"sync"
	"time"
)

func NewListener(listener *net.TCPListener) *Listener {
	return &Listener{listener, &sync.WaitGroup{}}
}

type Listener struct {
	*net.TCPListener
	wg *sync.WaitGroup
}

func (this *Listener) Accept() (*Conn, error) {
	tc, err := this.AcceptTCP()
	if err != nil {
		return nil, err
	}

	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)

	this.wg.Add(1)
	conn := &Conn{tc, this}
	return conn, nil
}

func (this *Listener) Wait() {
	this.wg.Wait()
}

func (this *Listener) GetFd() (uintptr, error) {
	file, err := this.TCPListener.File()
	if err != nil {
		return 0, err
	}
	return file.Fd(), nil
}

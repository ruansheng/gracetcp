package gracetcp

import (
	"fmt"
	"net"
)

type Conn struct {
	net.Conn
	listener *Listener
}

func (this *Conn) Close() error {
	this.listener.wg.Done()
	fmt.Println("Close")
	return this.Conn.Close()
}

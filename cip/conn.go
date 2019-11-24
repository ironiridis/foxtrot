package cip

import (
	"net"
	"sync"
)

// Conn represents an active communication with another CIP device
type Conn struct {
	mu     sync.RWMutex
	socket *net.TCPConn
	ipid   IPID
}

// RemoteAddr returns the result of calling RemoteAddr() on the underlying net.TCPConn.
func (c *Conn) RemoteAddr() net.Addr {
	return c.socket.RemoteAddr()
}

// LocalAddr returns the result of calling LocalAddr() on the underlying net.TCPConn.
func (c *Conn) LocalAddr() net.Addr {
	return c.socket.LocalAddr()
}

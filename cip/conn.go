package cip

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

const ErrorConnBadState Error = "cip.Conn in bad state"
const ErrorConnShortWrite Error = "failed to write entire packet"
const ErrorConnShortRead Error = "failed to read expected length"
const ErrorConnSocketNotReady Error = "socket is not ready"

// Conn represents an active communication channel with another CIP device
type Conn struct {
	Joins         *Joins
	mu            sync.RWMutex
	socket        *net.TCPConn
	ipid          IPID
	state         connState
	lastError     error
	lastHeartbeat time.Time
}

type connState int

const (
	connStateNew connState = iota
	connStateGotOnlineReq
	connStateSentHello
	connStateAwaitingHello
	connStateAwaitingOnlineOK
	connStateAwaitingSync
	connStateReady
	connStateClosed
)

type packetType byte

const (
	packetTypeOnlineReq = 0x01
	packetTypeOnlineOK  = 0x02
	packetTypeUpdate    = 0x05
	packetTypePing      = 0x0d
	packetTypePong      = 0x0e
	packetTypeHello     = 0x0f
)

// RemoteAddr returns the result of calling RemoteAddr() on the underlying net.TCPConn.
func (c *Conn) RemoteAddr() net.Addr {
	return c.socket.RemoteAddr()
}

// LocalAddr returns the result of calling LocalAddr() on the underlying net.TCPConn.
func (c *Conn) LocalAddr() net.Addr {
	return c.socket.LocalAddr()
}

func (c *Conn) err(e error) error {
	panic(e)
	c.mu.Lock()
	c.lastError = e
	if c.socket != nil {
		c.socket.Close()
		c.socket = nil
	}
	c.mu.Unlock()
	return e
}

func (c *Conn) stop() error {
	c.mu.Lock()
	c.state = connStateClosed
	if c.socket != nil {
		c.socket.Close()
		c.socket = nil
	}
	c.mu.Unlock()
	return nil
}

func (c *Conn) setState(s connState) {
	c.mu.Lock()
	c.state = s
	c.mu.Unlock()
}

func (c *Conn) readLoop() error {
	var n int
	var err error
	c.mu.RLock()
	s := c.socket
	c.mu.RUnlock()
	defer s.Close()
	for {
		b := make([]byte, 3)
		n, err = s.Read(b)
		if err == io.EOF {
			return c.stop()
		}
		if err != nil {
			return c.err(err)
		}
		if n != 3 {
			return c.err(ErrorConnShortRead)
		}
		p := make([]byte, int(b[1])<<8|int(b[2]))
		n, err = s.Read(p)
		if err != nil {
			return c.err(err)
		}
		if n != len(p) {
			return c.err(ErrorConnShortRead)
		}
		switch b[0] {
		case packetTypeOnlineReq:
			c.handleOnlineRequest(p)
		case packetTypeHello:
			c.handleHello(p)
		case packetTypeOnlineOK:
			c.handleOnlineOK(p)
		case packetTypeUpdate:
			c.unpackUpdates(p)
		case packetTypePing:
			c.handlePing(p)
		case packetTypePong:
			c.handlePong(p)
		default:
			return c.err(fmt.Errorf("unrecognized packet type %02x (payload length %d, payload=%x)", b[0], len(p), p))
		}
	}
}

func (c *Conn) inState(s connState) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state == s
}

func (c *Conn) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var sb strings.Builder
	sb.WriteString("CIP connection state:")
	switch c.state {
	case connStateNew:
		sb.WriteString("new")
	case connStateGotOnlineReq:
		sb.WriteString("got online request")
	case connStateSentHello:
		sb.WriteString("sent hello")
	case connStateAwaitingHello:
		sb.WriteString("awaiting hello")
	case connStateAwaitingSync:
		sb.WriteString("awaiting sync")
	case connStateReady:
		sb.WriteString("ready")
	case connStateClosed:
		sb.WriteString("closed")
	default:
		sb.WriteString(fmt.Sprintf("unknown state (%d)", c.state))
	}
	sb.WriteString(fmt.Sprintf(" IPID=%02x", c.ipid))
	if c.lastHeartbeat.IsZero() {
		sb.WriteString(" (no heartbeats)")
	} else {
		sb.WriteString(fmt.Sprintf(" last heartbeat: %s", time.Since(c.lastHeartbeat)))
	}
	return sb.String()
}

package cip

import (
	"bytes"
	"time"
)

const ErrorMalformedPacket Error = "malformed packet"

func (c *Conn) tx(b []byte) error {
	if c.lastError != nil {
		return c.lastError
	}
	if c.socket == nil {
		return c.err(ErrorConnSocketNotReady)
	}
	n, err := c.socket.Write(b)
	if err != nil {
		return c.err(err)
	}
	if n != len(b) {
		return c.err(ErrorConnShortWrite)
	}
	return nil
}

func (c *Conn) txPkt(t byte, p []byte) error {
	b := make([]byte, 3+len(p))
	b[0] = t
	b[1] = byte((len(p) & 0xff00) >> 8)
	b[2] = byte(len(p) & 0xff)
	copy(b[3:], p)
	return c.tx(b)
}

func (c *Conn) sendPing(p []byte) error {
	return c.txPkt(packetTypePing, []byte{0, 0})
}

func (c *Conn) handlePing(p []byte) error {
	c.mu.Lock()
	c.lastHeartbeat = time.Now()
	c.mu.Unlock()
	return c.sendPong(p)
}

func (c *Conn) sendPong(p []byte) error {
	return c.txPkt(packetTypePong, p)
}

func (c *Conn) handlePong(p []byte) error {
	c.mu.Lock()
	c.lastHeartbeat = time.Now()
	c.mu.Unlock()
	return nil
}

// client device requesting to come online
func (c *Conn) sendOnlineRequest() error {
	//TODO: determine purpose of mystery fields
	// first 5 bytes are some kind of device type, last byte (0x40) is unknown
	err := c.txPkt(packetTypeOnlineReq, []byte{0x00, 0x00, 0x00, 0x00, 0x00, byte(c.ipid), 0x40})
	if err == nil {
		c.state = connStateAwaitingHello
	}
	return err
}

func (c *Conn) handleOnlineRequest(p []byte) error {
	var err error

	if len(p) < 6 {
		return c.err(ErrorMalformedPacket)
	}
	c.mu.Lock()
	c.ipid = IPID(p[5])
	c.mu.Unlock()

	err = c.sendOnlineOK()
	if err != nil {
		return c.err(err)
	}
	return nil
}

// host device responding to online request
func (c *Conn) sendHello() error {
	//TODO: determine purpose of mystery field
	err := c.txPkt(packetTypeHello, []byte{0x02})
	if err == nil {
		c.setState(connStateSentHello)
	}
	return err
}

func (c *Conn) handleHello(p []byte) error {
	if len(p) != 1 {
		println("unexpected hello payload length: ", len(p))
	}
	if p[0] != 0x02 {
		println("unexpected hello payload: ", p)
	}
	c.setState(connStateAwaitingOnlineOK)
	return nil
}

// host device responding to online request
func (c *Conn) sendOnlineOK() error {
	//TODO: determine purpose of mystery fields
	err := c.txPkt(packetTypeOnlineOK, []byte{0x00, 0x00, 0x00, 0x03})
	if err == nil {
		c.setState(connStateAwaitingSync)
	}
	return err
}

func (c *Conn) handleOnlineOK(p []byte) error {
	if len(p) != 3 {
		println("unexpected online ok payload length: ", len(p))
	}
	if !bytes.Equal(p, []byte{0x00, 0x00, 0x00, 0x03}) {
		println("unexpected online ok payload: ", p)
	}
	c.setState(connStateAwaitingSync)
	return c.sendSync()
}

func (c *Conn) sendSync() error {
	//TODO: determine purpose of mystery fields
	return c.txPkt(packetTypeUpdate, []byte{0x00, 0x00, 0x02, 0x03, 0x00})
}

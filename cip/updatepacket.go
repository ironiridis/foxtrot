package cip

import (
	"bytes"
	"io"
)

func (c *Conn) readUpdateTask(r io.Reader) error {
	var n int
	var err error

	b := make([]byte, 3)
	n, err = r.Read(b)
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	if n < 3 {
		return ErrorConnShortRead
	}
	if b[0] != 0x00 {
		println("unexpected update task b[0]: ", b[0])
	}
	p := make([]byte, int(b[1])<<8|int(b[2]))
	n, err = r.Read(p)
	if err == io.EOF || n < len(p) {
		return ErrorConnShortRead
	}
	if err != nil {
		return err
	}
	switch p[0] {
	case 0x03: // sync
		if len(p) != 2 {
			println("unexpected update task sync payload length: ", len(p))
		}
		switch p[1] {
		case 0x00: // sync request
			c.Joins.Sync()
			if c.inState(connStateAwaitingSync) {
				c.setState(connStateReady)
				c.sendSync()
			}
		case 0x16: // sync complete
			if c.inState(connStateAwaitingSync) {
				c.setState(connStateReady)
			}
		default:
			println("unknown update task sync value:", p[1])
		}
	case 0x08: // time update
		// ignore
		//TODO: annotate received data format and weird decimal-as-hex quirk
	case 0x00, 0x27: // digital
		if len(p) != 3 {
			println("unexpected update task digital payload length: ", len(p))
		}
		j := Join(1 + (((uint16(p[2]) & 0x0f) << 8) | uint16(p[1])))
		c.Joins.DigitalIn(j, (p[2]&0x80) == 0)
	case 0x14: // analog
		if len(p) != 5 {
			println("unexpected update task analog payload length: ", len(p))
		}
		j := Join(1 + (uint16(p[1])<<8 | uint16(p[2])))
		v := Analog(uint16(p[3])<<8 | uint16(p[4]))
		c.Joins.AnalogIn(j, v)
	case 0x15: // serial
		//	j := Join(1 + (uint16(p[1])<<8 | uint16(p[2])))
		//	v := Analog(uint16(p[3])<<8 | uint16(p[4]))
		//	c.Joins.AnalogIn(j, v)
		println("serial update task unimplemented, payload is: ", p)
	default:
		println("unimplemented update task: ", p[0], " length: ", len(p))
	}
	return nil
}

func (c *Conn) unpackUpdates(p []byte) error {
	buf := bytes.NewBuffer(p)
	for buf.Len() > 0 {
		err := c.readUpdateTask(buf)
		if err != nil {
			return c.err(err)
		}
	}
	return nil
}

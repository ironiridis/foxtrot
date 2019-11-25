package cip

import "net"

const ErrorInvalidNetwork Error = "Network must be tcp or tcp4"
const ErrorInvalidIPID Error = "IPID must be between 0x03 and 0xFE inclusive when Dialing"

// Dial initiates a CIP connection to a listening CIP host. `network` must be
// either "tcp" or "tcp4" and will always use IPv4. `address` should be the
// host device, generally with port 41794 (eg "10.0.0.1:41794").
func Dial(network, address string, ipid IPID) (*Conn, error) {
	j := NewJoins()
	return DialWithJoins(network, address, ipid, j)
}

// DialWithJoins initiates a CIP connection to a listening CIP host. `network`
// must be either "tcp" or "tcp4" and will always use IPv4. `address` should be
// the host device, generally with port 41794 (eg "10.0.0.1:41794"). `j` must
// be an existing Joins structure which will be synchronized on connect.
func DialWithJoins(network, address string, ipid IPID, j *Joins) (*Conn, error) {
	var err error
	if network != "tcp" && network != "tcp4" {
		return nil, ErrorInvalidNetwork
	}
	if ipid < 0x03 || ipid > 0xfe {
		return nil, ErrorInvalidIPID
	}

	addr, err := net.ResolveTCPAddr(network, address)
	if err != nil {
		return nil, err
	}

	c := new(Conn)
	c.socket, err = net.DialTCP(network, nil, addr)
	if err != nil {
		return nil, err
	}
	c.Joins = j
	go c.readLoop()
	c.sendOnlineRequest()

	return c, nil
}
